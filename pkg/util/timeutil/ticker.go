package timeutil

import (
	"context"
	"time"
)

// Ticker is a wrapper around time.Ticker that allows to manually invoke
// a tick and can be stopped via context.
type Ticker struct {
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
	if t.d > 0 {
		if ctx == nil {
			panic("timeutil.PokeTicker: context is nil")
		}
		t.ctx = ctx
		if t.t == nil {
			t.t = time.NewTicker(t.d)
		} else {
			t.t.Reset(t.d)
		}
		go t.ticker()
	}
}

// Duration returns the ticker duration.
func (t *Ticker) Duration() time.Duration {
	return t.d
}

// Tick invokes a tick.
func (t *Ticker) Tick() {
	t.c <- time.Now()
}

// TickCh returns the ticker channel.
func (t *Ticker) TickCh() <-chan time.Time {
	return t.c
}

func (t *Ticker) ticker() {
	for {
		select {
		case <-t.ctx.Done():
			t.t.Stop()
			t.t = nil
			t.ctx = nil
			return
		case tm := <-t.t.C:
			t.c <- tm
		}
	}
}
