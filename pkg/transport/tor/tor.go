package tor

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const LoggerTag = "TOR"

type TorHTTP struct {
	mu     sync.Mutex
	ctx    context.Context
	waitCh <-chan error

	topics   map[string]transport.Message
	consumer *consumer
	producer *producer
	signer   ethereum.Signer
	log      log.Logger

	msgCh   map[string]chan transport.ReceivedMessage              // Channels for received messages.
	msgChFO map[string]*chanutil.FanOut[transport.ReceivedMessage] // Fan-out channels for received messages.
}

type Config struct {
	// Port on which the TOR hidden service will listen for incoming messages.
	Port int

	// Consumers is an instance of consumer provider.
	Consumers Consumers

	// Transport used by feeders.
	Transport http.RoundTripper

	// Topics is a list of subscribed topics. A value of the map a type of
	// message given as a nil pointer, e.g.: (*Message)(nil).
	Topics map[string]transport.Message

	// Signer used to verify price messages. Ignored in bootstrap mode.
	Signer ethereum.Signer

	// FeedersAddrs is a list of price feeders. Only feeders can create new
	// messages in the network.
	FeedersAddrs []ethereum.Address

	// Logger is a custom logger instance. If not provided then null
	// logger is used.
	Logger log.Logger
}

func New(cfg Config) (*TorHTTP, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	logger := cfg.Logger.WithField("tag", LoggerTag)
	us := newURLSigner(cfg.Signer, rand.Reader)
	ms := newMessageSigner(cfg.Signer)
	return &TorHTTP{
		topics: maputil.Copy(cfg.Topics),
		consumer: newConsumer(consumerConfig{
			URLSigner:     us,
			MessageSigner: ms,
			Server: &http.Server{
				Addr:              fmt.Sprintf("127.0.0.1:%d", cfg.Port),
				ReadTimeout:       60 * time.Second,
				ReadHeaderTimeout: 60 * time.Second,
				WriteTimeout:      60 * time.Second,
				IdleTimeout:       60 * time.Second,
			},
			FeedersAddrs: cfg.FeedersAddrs,
		}),
		producer: newProducer(producerConfig{
			URLSigner:     us,
			MessageSigner: ms,
			Client: &http.Client{
				Transport: cfg.Transport,
				Timeout:   60 * time.Second,
			},
			Consumers:     cfg.Consumers,
			FlushInterval: timeutil.NewTicker(10 * time.Second),
			Logger:        logger,
		}),
		signer:  cfg.Signer,
		log:     logger,
		msgCh:   make(map[string]chan transport.ReceivedMessage),
		msgChFO: make(map[string]*chanutil.FanOut[transport.ReceivedMessage]),
	}, nil
}

func (t *TorHTTP) Start(ctx context.Context) error {
	if t.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	t.ctx = ctx
	t.log.Info("Starting")
	for topic := range t.topics {
		t.msgCh[topic] = make(chan transport.ReceivedMessage, 10000)
		t.msgChFO[topic] = chanutil.NewFanOut(t.msgCh[topic])
	}

	if err := t.consumer.Start(ctx); err != nil {
		return err
	}
	if err := t.producer.Start(ctx); err != nil {
		return err
	}
	waitChFI := chanutil.NewFanIn(t.consumer.Wait(), t.producer.Wait())
	waitChFI.AutoClose()
	t.waitCh = waitChFI.Chan()
	go t.consumeMessagesRoutine()
	return nil
}

func (t *TorHTTP) Wait() <-chan error {
	return t.waitCh
}

func (t *TorHTTP) ID() []byte {
	return nil
}

func (t *TorHTTP) Broadcast(topic string, message transport.Message) error {
	t.log.WithField("topic", topic).Debug("Broadcasting message")
	bin, err := message.MarshallBinary()
	if err != nil {
		return err
	}
	return t.producer.Send(topic, bin)
}

func (t *TorHTTP) Messages(topic string) <-chan transport.ReceivedMessage {
	if ch, ok := t.msgChFO[topic]; ok {
		return ch.Chan()
	}
	return nil
}

func (t *TorHTTP) consumeMessagesRoutine() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case raw := <-t.consumer.Messages():
			typ, ok := t.topics[raw.topic]
			if !ok {
				continue
			}
			tr := reflect.TypeOf(typ).Elem()
			msg := reflect.New(tr).Interface().(transport.Message)
			if err := msg.UnmarshallBinary(raw.data); err != nil {
				t.log.
					WithError(err).
					Warn("unable to unmarshal message")
				continue
			}
			if _, ok := t.msgCh[raw.topic]; !ok {
				// This should never happen.
				t.log.
					WithField("topic", raw.topic).
					Panic("topic channel is not initialized")
			}
			t.msgCh[raw.topic] <- transport.ReceivedMessage{
				Message: msg,
				Author:  raw.signer,
			}
		}
	}
}
