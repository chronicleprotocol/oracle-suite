package retry

import (
	"context"
	"time"
)

// Retry runs the f function until it returns nil.
func Retry(ctx context.Context, f func() error, attempts int, delay time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = f()
		if err == nil {
			return nil
		}
		if i != attempts-1 {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
			case <-t.C:
			}
			t.Stop()
		}
	}
	return err
}
