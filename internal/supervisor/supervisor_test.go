package supervisor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type service struct {
	mu sync.Mutex

	started     bool
	failOnStart bool
	waitCh      chan error
}

func (s *service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failOnStart {
		return errors.New("err")
	}
	s.started = true
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()
		close(s.waitCh)
	}()
	return nil
}

func (s *service) Wait() chan error {
	return s.waitCh
}

func (s *service) Started() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

func TestSupervisor_CancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := New(ctx)

	s1 := &service{waitCh: make(chan error)}
	s2 := &service{waitCh: make(chan error)}
	s3 := &service{waitCh: make(chan error)}

	s.Watch(s1, s2, s3)

	require.NoError(t, s.Start())
	time.Sleep(100 * time.Millisecond)

	assert.True(t, s1.Started())
	assert.True(t, s2.Started())
	assert.True(t, s3.Started())

	cancel()
	time.Sleep(100 * time.Millisecond)

	select {
	case <-s.Wait():
	default:
		require.Fail(t, "Wait() channel should not be blocked")
	}

	assert.False(t, s1.Started())
	assert.False(t, s2.Started())
	assert.False(t, s3.Started())
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

	assert.False(t, s1.Started())
	assert.False(t, s2.Started())
	assert.False(t, s3.Started())
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
	time.Sleep(100 * time.Millisecond)

	assert.True(t, s1.Started())
	assert.True(t, s2.Started())
	assert.True(t, s3.Started())

	s2.waitCh <- errors.New("err")
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-s.Wait():
		require.Equal(t, "err", err.Error())
	default:
		require.Fail(t, "Wait() channel should not be blocked")
	}

	assert.False(t, s1.Started())
	assert.False(t, s2.Started())
	assert.False(t, s3.Started())
}
