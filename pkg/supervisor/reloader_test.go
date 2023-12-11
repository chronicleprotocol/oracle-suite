//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReloader(t *testing.T) {
	t.Run("start service", func(t *testing.T) {
		s := &service{waitCh: make(chan error)}
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				return nil
			},
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, r.Start(ctx))
		assert.Eventually(t, func() bool {
			return s.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("stop service", func(t *testing.T) {
		s := &service{waitCh: make(chan error)}
		c := make(chan struct{})
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				c <- struct{}{}
				close(serviceCh)
				return nil
			},
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, r.Start(ctx))
		assert.Eventually(t, func() bool {
			return s.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
		<-c
		assert.Eventually(t, func() bool {
			return !s.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("reload service", func(t *testing.T) {
		s1 := &service{waitCh: make(chan error)}
		s2 := &service{waitCh: make(chan error)}
		c := make(chan struct{})
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s1
				c <- struct{}{}
				serviceCh <- s2
				c <- struct{}{}
				close(serviceCh)
				return nil
			},
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, r.Start(ctx))
		assert.Eventually(t, func() bool {
			return s1.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
		<-c
		assert.Eventually(t, func() bool {
			return !s1.Started() && s2.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
		<-c
		assert.Eventually(t, func() bool {
			return !s1.Started() && !s2.Started()
		}, 100*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("service failed to start", func(t *testing.T) {
		s := &service{waitCh: make(chan error), failOnStart: true}
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				return nil
			},
		})
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		require.NoError(t, r.Start(ctx))
		require.Error(t, <-r.Wait())
	})

	t.Run("service stopped", func(t *testing.T) {
		s := &service{waitCh: make(chan error)}
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				s.waitCh <- nil
				return nil
			},
		})
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		require.NoError(t, r.Start(ctx))
		require.Nil(t, <-r.Wait())
	})

	t.Run("service failed", func(t *testing.T) {
		s := &service{waitCh: make(chan error)}
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				s.waitCh <- fmt.Errorf("service failed")
				return nil
			},
		})
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		require.NoError(t, r.Start(ctx))
		require.Error(t, <-r.Wait())
	})

	t.Run("factory failed", func(t *testing.T) {
		s := &service{waitCh: make(chan error)}
		r := NewReloader(ReloaderConfig{
			Factory: func(ctx context.Context, serviceCh chan Service) error {
				serviceCh <- s
				return fmt.Errorf("factory failed")
			},
		})
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		require.NoError(t, r.Start(ctx))
		require.Error(t, <-r.Wait())
	})
}
