package timeutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtx()

	start := time.Now()

	ticker := NewTicker(10 * time.Millisecond)
	ticker.Start(ctx)

	n := 0
	for n < 10 {
		<-ticker.TickCh()
		n++
	}

	cancelCtx()

	elapsed := time.Since(start)
	assert.True(t, elapsed >= 100*time.Millisecond)
}
