package replayer

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type eventProvider struct {
	eventsCh chan *messages.Event
}

func (e eventProvider) Start(ctx context.Context) error {
	return nil
}

func (e eventProvider) Events() chan *messages.Event {
	return e.eventsCh
}

func Test_Replayer(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	ch := make(chan *messages.Event)
	rep, err := New(Config{
		EventProvider: eventProvider{eventsCh: ch},
		ReplayAfter:   []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
	})

	require.NoError(t, err)
	require.NoError(t, rep.Start(ctx))

	evt := &messages.Event{Type: "test", EventDate: time.Now()}
	ch <- evt

	var count int32
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case recv := <-rep.Events():
				assert.Equal(t, evt, recv)
				atomic.AddInt32(&count, 1)
			}
		}
	}()

	// Message should resend immediately and then replayed twice after 100ms and 200ms.
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(3), atomic.LoadInt32(&count))

	// Eventually message should be removed from cache.
	time.Sleep(300 * time.Millisecond)
	rep.mu.Lock()
	assert.Equal(t, 0, rep.eventCache.list.Len())
	rep.mu.Unlock()
}

func Test_intervalGCD(t *testing.T) {
	tests := []struct {
		s    []time.Duration
		want time.Duration
	}{
		{[]time.Duration{1, 2, 3}, 1},
		{[]time.Duration{2, 4, 6}, 2},
		{[]time.Duration{3}, 3},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, intervalGCD(tt.s))
		})
	}
}
