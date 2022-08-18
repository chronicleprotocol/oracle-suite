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

package chanutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	ch3 := make(chan int)

	go func() {
		for i := 0; i < 2; i++ {
			ch1 <- i
			ch2 <- i
			ch3 <- i
		}
		close(ch1)
		close(ch2)
		close(ch3)
	}()

	ch := Merge(ch1, ch2, ch3)
	n := 0
	for range ch {
		n++
	}

	time.Sleep(time.Millisecond * 100)
	_, ok := <-ch
	assert.False(t, ok) // Channel should be closed.
	assert.Equal(t, n, 6)
}
