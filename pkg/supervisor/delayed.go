package supervisor

import (
	"context"
	"time"
)

// Delayed is a service that delays the start of another service.
type Delayed struct {
	service Service
	delay   time.Duration
	waitCh  chan error
}

// NewDelayed returns a new Delayed service.
func NewDelayed(service Service, delay time.Duration) *Delayed {
	return &Delayed{
		service: service,
		delay:   delay,
		waitCh:  make(chan error),
	}
}

// Start implements the Service interface.
func (d *Delayed) Start(ctx context.Context) error {
	go func() {
		t := time.NewTimer(d.delay)
		defer t.Stop()
		defer close(d.waitCh)
		select {
		case <-ctx.Done():
			d.waitCh <- ctx.Err()
		case <-t.C:
			d.waitCh <- d.service.Start(ctx)
		}
	}()
	return nil
}

// Wait implements the Service interface.
func (d *Delayed) Wait() <-chan error {
	err := <-d.waitCh
	if err != nil {
		return nil
	}
	return d.service.Wait()
}
