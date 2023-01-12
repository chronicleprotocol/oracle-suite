package recoverer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
)

type panicTransport struct{}

func (p *panicTransport) Start(ctx context.Context) error {
	return nil
}

func (p *panicTransport) Wait() <-chan error {
	return nil
}

func (p *panicTransport) Broadcast(topic string, message transport.Message) error {
	panic("test panic")
}

func (p *panicTransport) Messages(topic string) <-chan transport.ReceivedMessage {
	panic("test panic")
}

func TestRecoverer_Broadcast(t *testing.T) {
	r := New(&panicTransport{}, nil)
	err := r.Broadcast("test", nil)
	assert.Error(t, err)
}

func TestRecoverer_Messages(t *testing.T) {
	r := New(&panicTransport{}, nil)
	_ = r.Messages("test")
	// Should not panic.
}
