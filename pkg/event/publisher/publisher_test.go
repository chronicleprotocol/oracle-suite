package publisher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type testListener struct{ ch chan *messages.Event }
type testSigner struct{}

func (t *testListener) Start(_ context.Context) error {
	return nil
}

func (t *testListener) Events() chan *messages.Event {
	return t.ch
}

func (t testSigner) Sign(event *messages.Event) (bool, error) {
	event.Signatures["test"] = []byte("test")
	return true, nil
}

func TestEventPublisher(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	loc := local.New(ctx, 10, map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)})
	lis := &testListener{ch: make(chan *messages.Event, 10)}
	sig := &testSigner{}

	pub, err := New(ctx, Config{
		Listeners: []Listener{lis},
		Signers:   []Signer{sig},
		Transport: loc,
		Logger:    null.New(),
	})
	require.NoError(t, err)

	require.NoError(t, loc.Start())
	require.NoError(t, pub.Start())
	defer func() {
		cancelFunc()
		require.NoError(t, <-loc.Wait())
		require.NoError(t, <-pub.Wait())
	}()

	msg1 := &messages.Event{
		Type:       "event1",
		ID:         []byte("id1"),
		Index:      []byte("idx1"),
		Date:       time.Unix(1, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string][]byte{"sig_key": []byte("val")},
	}
	msg2 := &messages.Event{
		Type:       "event2",
		ID:         []byte("id2"),
		Index:      []byte("idx2"),
		Date:       time.Unix(2, 0),
		Data:       map[string][]byte{"data_key": []byte("val")},
		Signatures: map[string][]byte{"sig_key": []byte("val")},
	}
	lis.ch <- msg1
	lis.ch <- msg2

	time.Sleep(100 * time.Millisecond)

	rMsg1 := <-loc.Messages(messages.EventMessageName)
	rMsg2 := <-loc.Messages(messages.EventMessageName)

	assert.Equal(t, []byte("test"), rMsg1.Message.(*messages.Event).Signatures["test"])
	assert.Equal(t, []byte("test"), rMsg2.Message.(*messages.Event).Signatures["test"])
	// This test relies on us passing the same message instances, so the values
	// added by the signer will be visible in all objects, but this behavior is
	// not required.
	assert.Equal(t, msg1, rMsg1.Message.(*messages.Event))
	assert.Equal(t, msg2, rMsg2.Message.(*messages.Event))
}
