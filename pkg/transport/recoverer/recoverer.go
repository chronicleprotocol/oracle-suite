package recoverer

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

// Recoverer is a transport wrapper that handles panics that occur in the
// underlying transport.
type Recoverer struct {
	t transport.Transport
	l log.Logger
}

// New creates a new Recoverer transport.
func New(t transport.Transport, l log.Logger) *Recoverer {
	if t == nil {
		panic("t cannot be nil")
	}
	if l == nil {
		l = null.New()
	}
	return &Recoverer{t: t, l: l}
}

// Start implements the transport.Transport interface.
func (r *Recoverer) Start(ctx context.Context) error {
	return r.t.Start(ctx)
}

// Wait implements the transport.Transport interface.
func (r *Recoverer) Wait() <-chan error {
	return r.t.Wait()
}

// Broadcast implements the transport.Transport interface.
func (r *Recoverer) Broadcast(topic string, message transport.Message) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			r.l.WithField("panic", rec).Error("Recovered from panic")
			err = fmt.Errorf("recovered from panic: %v", rec)
		}
	}()
	return r.t.Broadcast(topic, message)
}

// Messages implements the transport.Transport interface.
func (r *Recoverer) Messages(topic string) <-chan transport.ReceivedMessage {
	defer func() {
		if rec := recover(); rec != nil {
			r.l.WithField("panic", rec).Error("Recovered from panic")
		}
	}()
	return r.t.Messages(topic)
}
