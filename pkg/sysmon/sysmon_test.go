package sysmon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/callback"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

func TestSysmon(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	f := int32(0)
	l := callback.New(log.Debug, func(level log.Level, fields log.Fields, msg string) {
		if msg == "Starting" || msg == "Stopped" {
			return
		}
		assert.Equal(t, "Status", msg)
		assert.Equal(t, log.Debug, level)
		atomic.StoreInt32(&f, 1)
	})

	s := New(time.Second, l)
	require.NoError(t, s.Start(ctx))

	// wait for log
	for i := 0; i < 10 && atomic.LoadInt32(&f) == 0; i++ {
		time.Sleep(100 * time.Millisecond)
	}

	// check if log has been sent
	require.Equal(t, int32(1), atomic.LoadInt32(&f), "log message has not been sent")

	// Wait() channel should return nil after cancelling the context
	ctxCancel()
	require.NoError(t, <-s.Wait())
}

func TestSysmon_RunWithDebugLevelOnly(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	l := null.New()

	s := New(time.Second, l)
	require.NoError(t, s.Start(ctx))

	select {
	case n := <-s.Wait():
		require.Nil(t, n)
	case <-time.NewTimer(time.Second).C:
		require.Fail(t, "sysmon should not start with verbosity other than debug")
	}
}
