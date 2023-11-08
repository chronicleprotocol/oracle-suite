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

package chanutil

import (
	"errors"
	"sync"
)

var ErrOutputClosed = errors.New("output channel closed")

// FanIn is a fan-in channel multiplexer. It takes multiple input channels
// and merges them into a single output channel. The implementation is
// thread-safe.
type FanIn[T any] struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	once     sync.Once
	closed   bool
	acquired bool
	out      chan T
}

// NewFanIn creates a new FanIn instance.
func NewFanIn[T any](chs ...<-chan T) *FanIn[T] {
	fi := &FanIn[T]{out: make(chan T)}
	_ = fi.Add(chs...)
	return fi
}

// Add adds a new input channel. If the output channel is already closed,
// the error is returned.
func (fi *FanIn[T]) Add(chs ...<-chan T) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.closed {
		return ErrOutputClosed
	}
	for _, ch := range chs {
		fi.wg.Add(1)
		go func(ch <-chan T) {
			for v := range ch {
				fi.out <- v
			}
			fi.wg.Done()
		}(ch)
	}
	return nil
}

// Wait blocks until all the input channels are closed.
func (fi *FanIn[T]) Wait() {
	fi.wg.Wait()
}

// Close closes the output channel. If the output channel is already closed,
// the error is returned. This method must be called only after all input
// channels are closed. Otherwise, the code may panic due to sending to a
// closed channel. To make sure that all input channels are closed, a call
// to this method can be preceded by a call to the Wait method. Alternatively,
// the AutoClose method can be used.
func (fi *FanIn[T]) Close() error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.closed {
		return ErrOutputClosed
	}
	close(fi.out)
	fi.closed = true
	return nil
}

// AutoClose will automatically close the output channel when all input
// channels are closed. This method must be called only after at least one
// input channel has been added. Otherwise, it will immediately close the
// output channel. This method is idempotent and non-blocking.
func (fi *FanIn[T]) AutoClose() {
	fi.once.Do(func() {
		go func() {
			fi.Wait()
			_ = fi.Close()
		}()
	})
}

// Chan returns the output channel.
//
// Because there is only one output channel, this method can be called only
// once to prevent using the same channel in multiple places which may lead
// to unexpected results. Calling this method multiple times will panic.
func (fi *FanIn[T]) Chan() <-chan T {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.acquired {
		panic("output channel already acquired")
	}
	fi.acquired = true
	return fi.out
}

// FanOut is a fan-out channel demultiplexer. It takes a single input channel
// and distributes its values to multiple output channels. The input channel
// is emptied even if there are no output channels. Output channels are closed
// when the input channel is closed. The implementation is thread-safe.
type FanOut[T any] struct {
	mu  sync.Mutex
	in  <-chan T
	out []chan T
}

// NewFanOut creates a new FanOut instance.
func NewFanOut[T any](ch <-chan T) *FanOut[T] {
	fo := &FanOut[T]{in: ch}
	go fo.worker()
	return fo
}

// Chan returns a new output channel.
func (fo *FanOut[T]) Chan() <-chan T {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	ch := make(chan T)
	if fo.in == nil {
		// If the input channel is already closed, the output channel is closed
		// immediately.
		close(ch)
		return ch
	}
	fo.out = append(fo.out, ch)
	return ch
}

func (fo *FanOut[T]) worker() {
	for v := range fo.in {
		for _, ch := range fo.chs() {
			ch <- v
		}
	}
	fo.mu.Lock()
	defer fo.mu.Unlock()
	for _, ch := range fo.out {
		close(ch)
	}
	// Remove references to the output channels to help the garbage collector
	// to free the memory. These channels are inaccessible at this point.
	fo.in = nil
	fo.out = nil
}

func (fo *FanOut[T]) chs() []chan T {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	return fo.out
}
