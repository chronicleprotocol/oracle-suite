package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelayed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &service{waitCh: make(chan error)}
	d := NewDelayed(s, 100*time.Millisecond)

	require.NoError(t, d.Start(ctx))
	assert.False(t, s.Started())
	time.Sleep(150 * time.Millisecond)
	assert.True(t, s.Started())
}
