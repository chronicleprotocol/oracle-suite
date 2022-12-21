package tor

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/tor/pb"
)

const consumePath = "/consume"

type consumer struct {
	urlSigner     *urlSigner
	messageSigner *messageSigner
	server        *httpserver.HTTPServer
	feedersAddrs  []ethereum.Address
	log           log.Logger

	messages chan consumedMessage
}

type consumerConfig struct {
	URLSigner     *urlSigner
	MessageSigner *messageSigner
	Server        *http.Server
	FeedersAddrs  []ethereum.Address
	Logger        log.Logger
}

type consumedMessage struct {
	topic  string
	signer []byte
	data   []byte
}

func newConsumer(cfg consumerConfig) *consumer {
	c := &consumer{
		urlSigner:     cfg.URLSigner,
		messageSigner: cfg.MessageSigner,
		feedersAddrs:  cfg.FeedersAddrs,
		log:           cfg.Logger,
		messages:      make(chan consumedMessage),
	}
	cfg.Server.Handler = http.HandlerFunc(c.handler)
	c.server = httpserver.New(cfg.Server)
	return c
}

func (c *consumer) Start(ctx context.Context) error {
	return c.server.Start(ctx)
}

func (c *consumer) Wait() <-chan error {
	return c.server.Wait()
}

func (c *consumer) Messages() <-chan consumedMessage {
	return c.messages
}

func (c *consumer) handler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		// Only POST requests are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			Debug("Invalid request method")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if consumePath != req.URL.Path {
		// Only requests to the /consume path are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			Debug("Invalid request path")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Header.Get("Content-Encoding") != "gzip" {
		// Only gzip-encoded requests are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			Debug("Invalid request encoding")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	addr, err := c.urlSigner.VerifyURL(req.URL.String())
	if err != nil {
		// Only requests with a valid signature are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithError(err).
			Debug("Invalid request signature")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	validFeeder := false
	for _, feederAddr := range c.feedersAddrs {
		if feederAddr == *addr {
			validFeeder = true
			break
		}
	}
	if !validFeeder {
		// Only messages from known feeders are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithField("feeder", addr.Hex()).
			Debug("Invalid feeder")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		// Unable to read request body.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithError(err).
			Debug("Unable to read request body")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err = gzipDecompress(body)
	if err != nil {
		// Unable to decompress request body.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithError(err).
			Debug("Unable to decompress request body")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	mp := &pb.MessagePack{}
	if err := proto.Unmarshal(body, mp); err != nil {
		// Unable to decode protobuf message.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithError(err).
			Debug("Unable to decode protobuf message")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	addr, err = c.messageSigner.VerifyMessage(mp)
	if err != nil {
		// Unable to verify message signature.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithError(err).
			Debug("Unable to verify message signature")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	validFeeder = false
	for _, feederAddr := range c.feedersAddrs {
		if feederAddr == *addr {
			validFeeder = true
			break
		}
	}
	if !validFeeder {
		// Only messages from known feeders are allowed.
		c.log.
			WithField("path", req.URL.Path).
			WithField("method", req.Method).
			WithField("feeder", addr.Hex()).
			Debug("Invalid feeder")
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	for topic, msgs := range mp.Messages {
		for _, msg := range msgs.Data {
			c.messages <- consumedMessage{
				topic: topic,
				data:  msg,
			}
		}
	}
	res.WriteHeader(http.StatusOK)
}

func gzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
