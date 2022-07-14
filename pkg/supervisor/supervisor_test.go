//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package supervisor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
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
	s := New(ctx, null.New())

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
	s := New(ctx, null.New())

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
	s := New(ctx, null.New())

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
