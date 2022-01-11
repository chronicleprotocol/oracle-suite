package observer

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const LoggerTag = "EVENT_OBSERVER"

type EventObserver struct {
	ctx    context.Context
	waitCh chan error

	signers   []Signer
	listeners []Listener
	transport transport.Transport
	log       log.Logger
}

type Listener interface {
	Events() chan *messages.Event
	Start(ctx context.Context) error
}

type Signer interface {
	Sign(event *messages.Event) (bool, error)
	Start(ctx context.Context) error
}

type Config struct {
	Listeners []Listener
	// Signer is a list of Signers used to sign events.
	Signers []Signer
	// Transport is implementation of transport used to send events to relayers.
	Transport transport.Transport
	// Logger is a current logger interface used by the EventObserver. The Logger
	// helps to monitor asynchronous processes.
	Logger log.Logger
}

func NewEventObserver(ctx context.Context, cfg Config) (*EventObserver, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	return &EventObserver{
		ctx:       ctx,
		waitCh:    make(chan error),
		transport: cfg.Transport,
		listeners: cfg.Listeners,
		signers:   cfg.Signers,
		log:       cfg.Logger.WithField("tag", LoggerTag),
	}, nil
}

func (l *EventObserver) Start() error {
	l.log.Infof("Starting")
	l.listenerLoop()
	for _, lis := range l.listeners {
		err := lis.Start(l.ctx)
		if err != nil {
			return err
		}
	}
	for _, sig := range l.signers {
		err := sig.Start(l.ctx)
		if err != nil {
			return err
		}
	}
	go l.contextCancelHandler()
	return nil
}

func (l *EventObserver) Wait() error {
	return <-l.waitCh
}

func (l *EventObserver) listenerLoop() {
	for _, li := range l.listeners {
		li := li
		go func() {
			for {
				select {
				case <-l.ctx.Done():
					return
				case e := <-li.Events():
					l.broadcast(e)
				}
			}
		}()
	}
}

func (l *EventObserver) broadcast(event *messages.Event) {
	if !l.sign(event) {
		return
	}
	l.log.
		WithField("id", hex.EncodeToString(event.ID)).
		WithField("type", event.Type).
		WithField("index", hex.EncodeToString(event.Index)).
		Info("Event broadcast")
	err := l.transport.Broadcast(messages.EventMessageName, event)
	if err != nil {
		l.log.
			WithError(err).
			Error("Unable to broadcast the event")
	}
	return
}

func (l *EventObserver) sign(event *messages.Event) bool {
	var err error
	var signed bool
	for _, s := range l.signers {
		ok, err := s.Sign(event)
		if !ok {
			continue
		}
		if err != nil {
			l.log.
				WithError(err).
				Error("Unable to sign event")
			continue
		}
		signed = true
	}
	if !signed {
		l.log.
			WithError(err).
			Error("There are no signers that supports the event")
	}
	return signed
}

func (l *EventObserver) contextCancelHandler() {
	defer func() { close(l.waitCh) }()
	defer l.log.Info("Stopped")
	<-l.ctx.Done()
}
