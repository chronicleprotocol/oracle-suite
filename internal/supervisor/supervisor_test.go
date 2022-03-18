package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type service struct {
	started     bool
	failOnStart bool
	waitCh      chan error
}

func (s *service) Start(ctx context.Context) error {
	if s.failOnStart {
		return errors.New("err")
	}
	s.started = true
	go func() {
		s.started = false
		<-ctx.Done()
		close(s.waitCh)
	}()
	return nil
}

func (s *service) Wait() chan error {
	return s.waitCh
}

func TestSupervisor_CancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := New(ctx)

	s1 := &service{waitCh: make(chan error)}
	s2 := &service{waitCh: make(chan error)}
	s3 := &service{waitCh: make(chan error)}

	s.Watch(s1, s2, s3)

	require.NoError(t, s.Start())
	require.True(t, s1.started)
	require.True(t, s2.started)
	require.True(t, s3.started)

	cancel()

	time.Sleep(100 * time.Millisecond)

	select {
	case <-s.Wait():
	default:
		require.Fail(t, "Wait() channel should not be blocked")
	}

	require.False(t, s1.started)
	require.False(t, s2.started)
	require.False(t, s3.started)
}

func TestSupervisor_FailToStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := New(ctx)

	s1 := &service{waitCh: make(chan error)}
	s2 := &service{waitCh: make(chan error)}
	s3 := &service{waitCh: make(chan error), failOnStart: true}

	s.Watch(s1, s2, s3)

	require.Error(t, s.Start())

	time.Sleep(100 * time.Millisecond)

	require.False(t, s1.started)
	require.False(t, s2.started)
	require.False(t, s3.started)
}

func TestSupervisor_OneFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := New(ctx)

	s1 := &service{waitCh: make(chan error)}
	s2 := &service{waitCh: make(chan error)}
	s3 := &service{waitCh: make(chan error)}

	s.Watch(s1, s2, s3)

	require.NoError(t, s.Start())
	require.True(t, s1.started)
	require.True(t, s2.started)
	require.True(t, s3.started)

	s2.waitCh <- errors.New("err")

	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-s.Wait():
		require.Equal(t, "err", err.Error())
	default:
		require.Fail(t, "Wait() channel should not be blocked")
	}

	require.False(t, s1.started)
	require.False(t, s2.started)
	require.False(t, s3.started)
}
