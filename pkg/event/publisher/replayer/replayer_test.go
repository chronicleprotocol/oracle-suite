//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package replayer

import (
	"context"
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
		Interval:      100 * time.Millisecond,
		ReplayAfter:   []time.Duration{300 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond},
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

	// Message should resend immediately and then replayed twice after 100ms, 200ms and 300ms.
	time.Sleep(400 * time.Millisecond)
	assert.Equal(t, int32(4), atomic.LoadInt32(&count))

	// Eventually message should be removed from cache.
	time.Sleep(200 * time.Millisecond)
	rep.mu.Lock()
	assert.Equal(t, 0, rep.eventCache.list.Len())
	rep.mu.Unlock()
}
