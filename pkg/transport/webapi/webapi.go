//nolint:gofmt,goimports // TODO(mdobak): remove nolint after upgrade Go to 1.19
package webapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	netURL "net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi/pb"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const (
	LoggerTag = "WEB_API"

	// consumePath is the URL path for the consume endpoint.
	consumePath = "/consume"

	// messageChanSize is the size of the message channel. It is used to buffer
	// messages before they are consumed.
	messageChanSize = 10000

	// defaultMaxClockSkew is the maximum allowed clock skew between the consumer
	// and the producer.
	defaultMaxClockSkew = 10 * time.Second

	// defaultTimeout is the default timeout for HTTP requests, both for
	// client and server.
	defaultTimeout = 60 * time.Second
)

// WebAPI is transport that uses HTTP API to send and receive messages.
// It is designed to use over secure network, e.g. Tor, I2P or VPN.
//
// Transport involves two main actors: message producers and consumers.
//
// Message producers are responsible for sending messages to the message
// consumers. List of addresses of message consumers is stored in the address
// book. Address book must implement the AddressBook interface.
//
// Message producers prepares a batch of messages and send them periodically
// to the message consumers. The batch is sent as a single HTTP POST request.
// Each request is signed using the private key of the producer. The signature
// is sent in the query string of the request. The signature is calculated
// using the following formula:
//
//   signature = sign(timestamp + rand)
//
// Where timestamp is the current UNIX timestamp in seconds encoded as a
// 10-base string, rand is a random 16-byte string encoded as a hex string.
// The sign function is provided by the Signer interface.
//
// Signature is stored in three query parameters:
//
//   s - hex encoded signature
//   t - timestamp in seconds encoded as a 10-base string
//   r - 16-byte random string encoded as a hex string
//
// Consumer verifies the signature and checks if message producer is on the
// allowlist. Timestamp must be within following ranges:
//
//   now - maxClockSkew <= timestamp <= now + flushInterval + maxClockSkew
//   lastRequest <= now + flushInterval - maxClockSkew
//
// Where now is the current UNIX timestamp in seconds, lastRequest is the
// timestamp of the last request received from the message producer and
// timestamp is the timestamp from the t query parameter.
//
// The purpose of the signature is to prevent replay attacks and to allow
// message consumers to quickly verify if the identity of the message producer.
//
// The request body is a protobuf message (see pb/tor.proto) that contains
// the following fields:
//
//   messages - a list of messages grouped by topic
//   signature - signature of the message pack
//
// The signature of the request body is calculated using Signer interface. The
// signing data is a concatenation of all messages in the batch prefixed with
// the topic name. Topics are sorted using sort.Strings function.
//
// The author recovered from the signature must be the same as the author
// recovered from the signature in the query string. Otherwise, the request
// is rejected.
//
// Request body is compressed using gzip compression. Requests that are not
// compressed or are not have a valid gzip header are rejected.
//
// All request must have Content-Type header set to application/x-protobuf.
//
// The HTTP server returns HTTP 200 OK response if the request is valid.
// Otherwise, it returns 429 Too Many Requests response if producer sends
// messages too often or 400 Bad Request response for any other error.
type WebAPI struct {
	mu     sync.RWMutex
	ctx    context.Context
	waitCh <-chan error

	// State fields:
	messagePack *pb.MessagePack                                        // Message pack to be sent on next flush.
	lastReqs    map[ethereum.Address]time.Time                         // Last timestamp received from each producer.
	msgCh       map[string]chan transport.ReceivedMessage              // Channels for received messages.
	msgChFO     map[string]*chanutil.FanOut[transport.ReceivedMessage] // Fan-out channels for received messages.

	// Configuration fields:
	addressBook  AddressBook
	topics       map[string]transport.Message
	allowlist    []ethereum.Address
	flushTicker  *timeutil.Ticker
	signer       ethereum.Signer
	client       *http.Client
	server       *httpserver.HTTPServer
	rand         io.Reader
	maxClockSkew time.Duration
	log          log.Logger
}

// Config is a configuration of WebAPI.
type Config struct {
	// ListenAddr is the address of the HTTP server that is used to receive
	// messages. If used with TOR, server should listen on localhost.
	//
	// Ignored if Server is not nil.
	//
	// Cannot be empty if Server is nil.
	ListenAddr string

	// AddressBook provides the list of message consumers. All produced messages
	// will be sent only to consumers from the address book.
	//
	// Cannot be nil.
	AddressBook AddressBook

	// Topics is a list of subscribed topics. A value of the map a type of
	// message given as a nil pointer, e.g.: (*Message)(nil).
	Topics map[string]transport.Message

	// AuthorAllowlist is a list of allowed message authors. Only messages from
	// these addresses will be accepted.
	AuthorAllowlist []ethereum.Address

	// FlushTicker specifies how often the producer will flush messages
	// to the consumers. If FlushTicker is nil, default ticker with 1 minute
	// interval is used.
	//
	// Flush interval must be less or equal to timeout.
	FlushTicker *timeutil.Ticker

	// Signer used to sign and verify messages.
	Signer ethereum.Signer

	// Timeout is a timeout for HTTP requests.
	//
	// If timeout is zero, default value will be used (60 seconds).
	Timeout time.Duration

	// Server is an optional custom HTTP server that will be used to receive
	// messages. If provided, ListenAddr and Timeout are ignored.
	Server *httpserver.HTTPServer

	// Client is an optional custom HTTP client that will be used to send
	// messages. If provided, Timeout is ignored.
	Client *http.Client

	// Rand is an optional random number generator. If not provided, Reader
	// from crypto/rand package will be used.
	Rand io.Reader

	// MaxClockSkew is the maximum allowed clock skew between the consumer
	// and the producer. If not provided, default value will be used (10 seconds).
	MaxClockSkew time.Duration

	// Logger is a custom logger instance. If not provided then null
	// logger is used.
	Logger log.Logger
}

// New returns a new instance of WebAPI.
func New(cfg Config) (*WebAPI, error) {
	if cfg.Server == nil && len(cfg.ListenAddr) == 0 {
		return nil, errors.New("listen address must be provided")
	}
	if cfg.AddressBook == nil {
		return nil, errors.New("address book must be provided")
	}
	if cfg.Signer == nil {
		return nil, errors.New("signer must be provided")
	}
	if cfg.FlushTicker == nil {
		return nil, errors.New("flush interval must be provided")
	}
	if cfg.FlushTicker.Duration() > 0 && cfg.Timeout >= cfg.FlushTicker.Duration() {
		return nil, errors.New("timeout must be less or equal to flush interval")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxClockSkew == 0 {
		cfg.MaxClockSkew = defaultMaxClockSkew
	}
	if cfg.Rand == nil {
		cfg.Rand = rand.Reader
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	server := cfg.Server
	client := cfg.Client
	if server == nil {
		server = httpserver.New(&http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       cfg.Timeout,
			ReadHeaderTimeout: cfg.Timeout,
			WriteTimeout:      cfg.Timeout,
			IdleTimeout:       cfg.Timeout,
		})
	}
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	w := &WebAPI{
		topics:       maputil.Copy(cfg.Topics),
		allowlist:    sliceutil.Copy(cfg.AuthorAllowlist),
		addressBook:  cfg.AddressBook,
		flushTicker:  cfg.FlushTicker,
		client:       client,
		server:       server,
		signer:       cfg.Signer,
		lastReqs:     make(map[ethereum.Address]time.Time),
		msgCh:        make(map[string]chan transport.ReceivedMessage),
		msgChFO:      make(map[string]*chanutil.FanOut[transport.ReceivedMessage]),
		maxClockSkew: cfg.MaxClockSkew,
		rand:         cfg.Rand,
		log:          cfg.Logger.WithField("tag", LoggerTag),
	}
	w.server.SetHandler(http.HandlerFunc(w.consumeHandler))
	return w, nil
}

// Start implements the supervisor.Supervisor interface.
func (w *WebAPI) Start(ctx context.Context) error {
	if w.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	w.ctx = ctx
	w.log.Info("Starting")
	for topic := range w.topics {
		w.msgCh[topic] = make(chan transport.ReceivedMessage, messageChanSize)
		w.msgChFO[topic] = chanutil.NewFanOut(w.msgCh[topic])
	}
	if err := w.server.Start(ctx); err != nil {
		return err
	}
	w.flushTicker.Start(ctx)
	go w.flushRoutine(ctx)
	go w.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Supervisor interface.
func (w *WebAPI) Wait() <-chan error {
	return w.waitCh
}

// Broadcast implements the transport.Transport interface.
func (w *WebAPI) Broadcast(topic string, message transport.Message) error {
	w.log.WithField("topic", topic).Debug("Broadcasting message")
	bin, err := message.MarshallBinary()
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	// Add message to the next batch. The batch will be sent to the consumers
	// on the next flush (see flushRoutine).
	if w.messagePack == nil {
		w.messagePack = &pb.MessagePack{Messages: make(map[string]*pb.MessagePack_Messages)}
	}
	if w.messagePack.Messages[topic] == nil {
		w.messagePack.Messages[topic] = &pb.MessagePack_Messages{}
	}
	w.messagePack.Messages[topic].Data = append(w.messagePack.Messages[topic].Data, bin)
	return nil
}

// Messages implements the transport.Transport interface.
func (w *WebAPI) Messages(topic string) <-chan transport.ReceivedMessage {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if ch, ok := w.msgChFO[topic]; ok {
		return ch.Chan()
	}
	return nil
}

// flushMessages sends the current batch of messages to the consumers.
// The batch is cleared after the messages are sent.
func (w *WebAPI) flushMessages(ctx context.Context, t time.Time) error {
	w.mu.Lock()
	defer func() {
		w.messagePack = nil
		w.mu.Unlock()
	}()
	if err := signMessage(w.messagePack, w.signer); err != nil {
		return err
	}
	bin, err := proto.Marshal(w.messagePack)
	if err != nil {
		return err
	}
	bin, err = gzipCompress(bin)
	if err != nil {
		return err
	}
	cons, err := w.addressBook.Consumers(ctx)
	if err != nil {
		return err
	}
	for _, addr := range cons {
		go w.doHTTPRequest(ctx, addr, bin, t)
	}
	return nil
}

// doHTTPRequest sends a POST request to the given address with the given
// data. The data must be gzipped protobuf-encoded MessagePack. The t parameter
// is the time used for the URL signature.
func (w *WebAPI) doHTTPRequest(ctx context.Context, addr string, data []byte, t time.Time) {
	w.log.WithField("addr", addr).Info("Sending messages to consumer")

	// Sign the URL.
	url, err := signURL(
		fmt.Sprintf("%s%s", strings.TrimRight(addr, "/"), consumePath),
		t,
		w.signer,
		w.rand,
	)
	if err != nil {
		w.log.WithError(err).WithField("addr", addr).Error("Failed to sign URL")
		return
	}

	// Prepare the request.
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		w.log.WithField("addr", addr).WithError(err).Error("Failed to create request")
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")

	// Send the request.
	res, err := w.client.Do(req)
	if err != nil {
		w.log.WithField("addr", addr).WithError(err).Error("Failed to send messages to consumer")
		return
	}
	defer res.Body.Close()
}

// consumeHandler handles incoming messages from consumers.
//
// Request must be a POST request to the /consume path with a protobuf-encoded
// MessagePack in the body.
//nolint:funlen,gocyclo
func (w *WebAPI) consumeHandler(res http.ResponseWriter, req *http.Request) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.ctx.Err() != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	fields := log.Fields{
		"method": req.Method,
		"addr":   req.RemoteAddr,
		"url":    req.URL.String(),
	}

	w.log.WithFields(fields).Debug("Received request")

	// Only POST requests are allowed.
	if req.Method != http.MethodPost {
		w.log.WithFields(fields).Debug("Invalid request method")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only requests to the /consume path are allowed.
	if req.URL.Path != consumePath {
		w.log.WithFields(fields).Debug("Invalid request path")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only requests with the protobuf content type are allowed.
	if req.Header.Get("Content-Type") != "application/x-protobuf" {
		w.log.WithFields(fields).Debug("Invalid content type")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only requests with the gzip content encoding are allowed.
	if req.Header.Get("Content-Encoding") != "gzip" {
		w.log.WithFields(fields).Debug("Invalid request encoding")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify the request URL signature.
	requestAuthor, timestamp, err := verifyURL(req.URL.String(), w.signer)
	if err != nil {
		w.log.WithFields(fields).WithError(err).Debug("Invalid request signature")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	fields["feeder"] = requestAuthor
	fields["timestamp"] = timestamp

	// Verify if the feeder is allowed to send messages.
	if !sliceutil.Contains(w.allowlist, *requestAuthor) {
		w.log.WithFields(fields).Debug("Feeder not allowed to send messages")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Message timestamp must be within the allowed time window:
	// [now - flushInterval - maxClockSkew, now + maxClockSkew].
	currentTimestamp := time.Now()
	if timestamp.After(currentTimestamp.Add(w.maxClockSkew)) {
		w.log.WithFields(fields).Warn("Timestamp too far in the future")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if timestamp.Before(currentTimestamp.Add(-w.flushTicker.Duration() - w.maxClockSkew)) {
		w.log.WithFields(fields).Warn("Timestamp too far in the past")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Message timestamp must be newer than the last received message by
	// flushInterval - defaultMaxClockSkew.
	if timestamp.Before(w.lastReqs[*requestAuthor].Add(w.flushTicker.Duration() - defaultMaxClockSkew)) {
		w.log.WithFields(fields).Warn("Too many messages received in a short time")
		res.WriteHeader(http.StatusTooManyRequests)
		return
	}
	w.lastReqs[*requestAuthor] = timestamp

	// Read the request body.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.log.WithFields(fields).WithError(err).Warn("Unable to read request body")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err = gzipDecompress(body)
	if err != nil {
		w.log.WithFields(fields).WithError(err).Warn("Unable to decompress request body")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Unmarshal protobuf-encoded MessagePack.
	mp := &pb.MessagePack{}
	if err := proto.Unmarshal(body, mp); err != nil {
		w.log.WithFields(fields).WithError(err).Warn("Unable to decode protobuf message")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify the message signature and verify that the author of the message
	// is same as the author of the request. There is no scenario in which the
	// author of the message and the request author can be different.
	messagePackAuthor, err := verifyMessage(mp, w.signer)
	if err != nil {
		w.log.WithFields(fields).WithError(err).Warn("Invalid message pack signature")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if *messagePackAuthor != *requestAuthor {
		w.log.WithFields(fields).Warn("Message pack author does not match request author")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Send messages from the MessagePack to the msgCh channel.
	for topic, msgs := range mp.Messages {
		for _, bin := range msgs.Data {
			typ, ok := w.topics[topic]
			if !ok {
				continue // Ignore messages for unknown topics.
			}
			ref := reflect.TypeOf(typ).Elem()
			msg := reflect.New(ref).Interface().(transport.Message)
			if err := msg.UnmarshallBinary(bin); err != nil {
				w.log.WithFields(fields).WithError(err).Warn("Unable to unmarshal message")
				continue
			}
			if _, ok := w.msgCh[topic]; !ok {
				// PANIC!
				// This should never happen because the keys of w.msgCh are
				// the same as the keys of w.topics.
				w.log.WithField("topic", topic).Panic("Topic channel is not initialized")
			}
			w.msgCh[topic] <- transport.ReceivedMessage{
				Message: msg,
				Author:  requestAuthor.Bytes(),
			}
		}
	}
}

// flushRoutine periodically sends the buffered messages to the
// consumers.
func (w *WebAPI) flushRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tm := <-w.flushTicker.TickCh():
			if err := w.flushMessages(ctx, tm); err != nil {
				w.log.
					WithError(err).
					Error("Failed to send messages")
			}
		}
	}
}

// contextCancelHandler handles context cancellation.
func (w *WebAPI) contextCancelHandler() {
	defer w.log.Info("Stopped")
	<-w.ctx.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.msgCh {
		close(ch)
	}
}

// signURL signs URLs using the given signer. It is used to quickly verify that a
// request comes from a specific feeder.
//
// The URL is signed by appending the following query parameters:
//  t: Unix timestamp of the signature as a 10-base integer.
//  r: Random 16-byte value encoded in hex.
//  s: Signature created by signing the concatenation of t and r.
func signURL(url string, tm time.Time, signer ethereum.Signer, rand io.Reader) (string, error) {
	r := make([]byte, 16)
	if _, err := rand.Read(r); err != nil {
		return "", err
	}
	t := strconv.FormatInt(tm.Unix(), 10)
	s, err := signer.Signature([]byte(t + hex.EncodeToString(r)))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s?t=%s&r=%x&s=%x", url, t, r, s.Bytes()), nil
}

// verifyURL verifies URL signature calculated by signURL and returns signer
// address and timestamp.
func verifyURL(url string, signer ethereum.Signer) (*ethereum.Address, time.Time, error) {
	p, err := netURL.Parse(url)
	if err != nil {
		return nil, time.Time{}, err
	}
	q := p.Query()
	t, err := strconv.ParseInt(q.Get("t"), 10, 64)
	if err != nil {
		return nil, time.Time{}, err
	}
	s, err := hex.DecodeString(q.Get("s"))
	if err != nil {
		return nil, time.Time{}, err
	}
	addr, err := signer.Recover(ethereum.SignatureFromBytes(s), []byte(q.Get("t")+q.Get("r")))
	if err != nil {
		return nil, time.Time{}, err
	}
	return addr, time.Unix(t, 0), nil
}

// signMessage signs the given message pack with the signer's private key.
func signMessage(msg *pb.MessagePack, signer ethereum.Signer) error {
	sig, err := signer.Signature(messageSigningData(msg))
	if err != nil {
		return err
	}
	msg.Signature = sig.Bytes()
	return nil
}

// verifyMessage verifies message pack signature and returns signer address.
func verifyMessage(msg *pb.MessagePack, signer ethereum.Signer) (*ethereum.Address, error) {
	return signer.Recover(ethereum.SignatureFromBytes(msg.Signature), messageSigningData(msg))
}

// messageSigningData returns the data used to sign the given message pack.
// The data is the concatenation of all topics followed by topic data. Topics
// are sorted in ascending order using sort.Strings.
func messageSigningData(msg *pb.MessagePack) []byte {
	var signingData []byte
	topics := maputil.SortKeys(msg.Messages, sort.Strings)
	for _, topic := range topics {
		signingData = append(signingData, topic...)
		for _, data := range msg.Messages[topic].Data {
			signingData = append(signingData, data...)
		}
	}
	return signingData
}

// gzipCompress compresses the given data using gzip.
func gzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	g := gzip.NewWriter(&b)
	_, err := g.Write(data)
	if err != nil {
		return nil, err
	}
	err = g.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// gzipDecompress decompresses the given data using gzip.
func gzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
