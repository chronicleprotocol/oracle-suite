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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFanIn(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	ch3 := make(chan int)

	fi := NewFanIn(ch1, ch2, ch3)
	ch := fi.Chan()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var v []int

		for i := 0; i < 3; i++ {
			v = append(v, <-ch)
		}
		assert.ElementsMatch(t, []int{1, 2, 3}, v)
		wg.Done()
	}()

	ch1 <- 1
	ch2 <- 2
	ch3 <- 3

	close(ch1)
	close(ch2)
	close(ch3)

	fi.Wait()
	assert.NoError(t, fi.Close())

	// The output channel should be closed after closing all the input
	// channels.
	_, ok := <-ch
	assert.False(t, ok)

	// Wait for the goroutine to finish.
	wg.Wait()
}

func TestFanIn_AutoClose(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	ch3 := make(chan int)

	fi := NewFanIn(ch1, ch2, ch3)

	close(ch1)
	close(ch2)
	close(ch3)

	fi.AutoClose()

	// The output channel should be automatically closed because all input
	// channels are closed.
	_, ok := <-fi.Chan()
	assert.False(t, ok)
	assert.Error(t, fi.Close())
}

func TestFanIn_AutoClose_NoInputs(t *testing.T) {
	fi := NewFanIn[int]()

	fi.AutoClose()

	time.Sleep(10 * time.Millisecond)

	// The output channel should be automatically closed because there are no
	// input channels.
	_, ok := <-fi.Chan()
	assert.False(t, ok)
	assert.Error(t, fi.Close())
}

func TestFanIn_AutoClose_AddAfterClose(t *testing.T) {
	ch1 := make(chan int)
	fi := NewFanIn(ch1)
	close(ch1)

	fi.Wait()
	assert.NoError(t, fi.Close())

	// Adding a new channel after closing the fan-in should return an error.
	assert.Error(t, fi.Add(make(chan int)))
}

func TestFanIn_Chan_Panic(t *testing.T) {
	ch := make(chan int)
	fi := NewFanIn(ch)

	require.NotNil(t, fi.Chan())

	// Second call to Chan should panic.
	assert.Panics(t, func() { fi.Chan() })
}

func TestFanOut(t *testing.T) {
	ch := make(chan int)

	fo := NewFanOut(ch)

	out1 := fo.Chan()
	out2 := fo.Chan()
	out3 := fo.Chan()

	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
	}()

	var mu sync.Mutex
	var wg sync.WaitGroup
	var vs [][]int
	for i, ch := range []<-chan int{out1, out2, out3} {
		wg.Add(1)
		mu.Lock()
		vs = append(vs, []int{})
		mu.Unlock()
		go func(ch <-chan int, i int) {
			for v := range ch {
				mu.Lock()
				vs[i] = append(vs[i], v)
				mu.Unlock()
			}
			wg.Done()
		}(ch, i)
	}

	// Wait for all the output channels to be closed. This should happen after
	// closing the input channel.
	wg.Wait()

	// Any of the output channels should receive copy of all the values sent to
	// the input channel.
	assert.ElementsMatch(t, []int{1, 2, 3}, vs[0])
	assert.ElementsMatch(t, []int{1, 2, 3}, vs[1])
	assert.ElementsMatch(t, []int{1, 2, 3}, vs[2])

	// Any output channel should be closed after closing the input channel.
	_, ok := <-fo.Chan()
	assert.False(t, ok)
}
