package timeutil

import (
	"context"
	"sync"
	"time"
)

// Ticker is a wrapper around time.Ticker that allows to manually invoke
// a tick and can be stopped via context.
type Ticker struct {
	mu  sync.Mutex
	ctx context.Context

	d time.Duration
	t *time.Ticker
	c chan time.Time
}

// NewTicker returns a new Ticker instance.
// If d is 0, the ticker will not be started and only manual ticks will be
// possible.
func NewTicker(d time.Duration) *Ticker {
	return &Ticker{d: d, c: make(chan time.Time)}
}

// Start starts the ticker.
func (t *Ticker) Start(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx != nil {
		panic("timeutil.PokeTicker: ticker is already started")
	}
	if ctx == nil {
		panic("timeutil.PokeTicker: context is nil")
	}
	t.ctx = ctx
	go t.ticker()
}

// Duration returns the ticker duration.
func (t *Ticker) Duration() time.Duration {
	return t.d
}

// Tick sends a tick to the ticker channel.
// Ticker must be started before calling this method.
func (t *Ticker) Tick() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx == nil || t.ctx.Err() != nil {
		panic("timeutil.PokeTicker: ticker is not started")
	}
	t.c <- time.Now()
}

// TickCh returns the ticker channel.
func (t *Ticker) TickCh() <-chan time.Time {
	return t.c
}

func (t *Ticker) ticker() {
	if t.d == 0 {
		return
	}
	t.t = time.NewTicker(t.d)
	for {
		select {
		case <-t.ctx.Done():
			t.mu.Lock()
			t.t.Stop()
			t.t = nil
			t.ctx = nil
			t.mu.Unlock()
			return
		case tm := <-t.t.C:
			t.c <- tm
		}
	}
}
