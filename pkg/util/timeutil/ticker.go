package timeutil

import (
	"context"
	"sync"
	"time"
)

// Ticker is a wrapper around time.Ticker that allows to manually invoke
// a tick and can be stopped via context.
type Ticker struct {
	mu  sync.RWMutex
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
		panic("timeutil.Ticker: ticker is already started")
	}
	if ctx == nil {
		panic("timeutil.Ticker: context is nil")
	}
	t.ctx = ctx
	go t.ticker(t.d)
}

// Duration returns the ticker duration.
func (t *Ticker) Duration() time.Duration {
	return t.d
}

// Tick sends a tick to the ticker channel.
// Ticker must be started before calling this method.
func (t *Ticker) Tick() {
	t.TickAt(time.Now())
}

// TickAt sends a tick to the ticker channel with the given time.
// Ticker must be started before calling this method.
func (t *Ticker) TickAt(at time.Time) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.ctx == nil || t.ctx.Err() != nil {
		panic("timeutil.Ticker: ticker is not started")
	}
	t.c <- at
}

// TickCh returns the ticker channel.
func (t *Ticker) TickCh() <-chan time.Time {
	return t.c
}

func (t *Ticker) ticker(d time.Duration) {
	if d == 0 {
		return
	}
	t.t = time.NewTicker(d)
	for {
		select {
		case <-t.ctx.Done():
			t.mu.Lock()
			t.ctx = nil
			t.mu.Unlock()
			t.t.Stop()
			return
		case tm := <-t.t.C:
			t.c <- tm
		}
	}
}

// VarTicker is a wrapper around time.Ticker that allows to manually invoke
// a tick and can be stopped via context.
//
// It allows to specify multiple durations. The ticker will tick at the
// specified durations and then repeat the last duration.
type VarTicker struct {
	mu  sync.RWMutex
	ctx context.Context

	d []time.Duration
	c chan time.Time
}

// NewVarTicker returns a new VarTicker instance.
func NewVarTicker(ds ...time.Duration) *VarTicker {
	return &VarTicker{d: ds, c: make(chan time.Time)}
}

// Start starts the ticker.
func (t *VarTicker) Start(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx != nil {
		panic("timeutil.VarTicker: ticker is already started")
	}
	if ctx == nil {
		panic("timeutil.VarTicker: context is nil")
	}
	t.ctx = ctx
	go t.ticker(t.d)
}

// Durations returns the ticker durations.
func (t *VarTicker) Durations() []time.Duration {
	return t.d
}

// Tick sends a tick to the ticker channel.
// Ticker must be started before calling this method.
func (t *VarTicker) Tick() {
	t.TickAt(time.Now())
}

// TickAt sends a tick to the ticker channel with the given time.
// Ticker must be started before calling this method.
func (t *VarTicker) TickAt(at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx == nil || t.ctx.Err() != nil {
		panic("timeutil.VarTicker: ticker is not started")
	}
	t.c <- at
}

// TickCh returns the ticker channel.
func (t *VarTicker) TickCh() <-chan time.Time {
	return t.c
}

func (t *VarTicker) ticker(d []time.Duration) {
	if len(d) == 0 {
		return
	}

	// If there are more than one duration, use timer until the last one
	// duration.
	if len(d) >= 2 {
		timer := time.NewTimer(d[0])
		for {
			select {
			case <-t.ctx.Done():
				t.mu.Lock()
				t.ctx = nil
				t.mu.Unlock()
				timer.Stop()
				return
			case tm := <-timer.C:
				t.c <- tm
			}
			d = d[1:]
			if len(d) == 1 {
				break
			}
			timer.Reset(d[0])
		}
	}

	// Use ticker for the last duration.
	ticker := time.NewTicker(d[0])
	for {
		select {
		case <-t.ctx.Done():
			t.mu.Lock()
			t.ctx = nil
			t.mu.Unlock()
			ticker.Stop()
			return
		case tm := <-ticker.C:
			t.c <- tm
		}
	}
}
