package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/event/store/memory"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestEventStore(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tra := local.New(ctx, 1, map[string]transport.Message{messages.EventMessageName: (*messages.Event)(nil)})

	mem := memory.New(time.Minute)
	evs, err := New(ctx, Config{
		Storage:   mem,
		Transport: tra,
		Log:       null.New(),
	})
	require.NoError(t, err)

	require.NoError(t, evs.Start())
	require.NoError(t, tra.Start())
	defer func() {
		cancelFunc()
		require.NoError(t, <-evs.Wait())
		require.NoError(t, <-tra.Wait())
	}()

	event := &messages.Event{
		Date:       time.Now(),
		Type:       "test",
		ID:         []byte("test"),
		Index:      []byte("idx"),
		Data:       map[string][]byte{"test": []byte("test")},
		Signatures: map[string][]byte{"test": []byte("test")},
	}
	require.NoError(t, tra.Broadcast(messages.EventMessageName, event))

	time.Sleep(100 * time.Millisecond)

	events, err := evs.Events("test", []byte("idx"))
	require.NoError(t, err)

	require.Len(t, events, 1)
	assert.Equal(t, event.Date.Unix(), events[0].Date.Unix())
	assert.Equal(t, event.Type, events[0].Type)
	assert.Equal(t, event.ID, events[0].ID)
	assert.Equal(t, event.Index, events[0].Index)
	assert.Equal(t, event.Data, events[0].Data)
	assert.Equal(t, event.Signatures, events[0].Signatures)
}
