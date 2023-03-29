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
		waitCh:  make(chan error, 1),
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
			err := ctx.Err()
			switch err {
			case context.Canceled:
				d.waitCh <- nil
			case context.DeadlineExceeded:
				d.waitCh <- nil
			default:
				d.waitCh <- err
			}
		case <-t.C:
			err := d.service.Start(ctx)
			if err != nil {
				d.waitCh <- err
				return
			}
			d.waitCh <- <-d.service.Wait()
		}
	}()
	return nil
}

// Wait implements the Service interface.
func (d *Delayed) Wait() <-chan error {
	return d.waitCh
}
