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

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry_noerror(t *testing.T) {
	n := time.Now()

	require.NoError(t, Retry(context.Background(), func() error {
		return nil
	}, 3, time.Millisecond*100))

	require.Less(t, time.Since(n), time.Millisecond*100)
}

func TestRetry_error(t *testing.T) {
	n := time.Now()
	c := 0

	require.Error(t, Retry(context.Background(), func() error {
		c++
		return errors.New("error")
	}, 3, time.Millisecond*100))

	require.Greater(t, time.Since(n), time.Millisecond*200)
	require.Equal(t, 3, c)
}

func TestRetry_ctxCancel(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	n := time.Now()
	time.AfterFunc(time.Millisecond*100, ctxCancel)

	require.Error(t, Retry(ctx, func() error {
		return errors.New("error")
	}, 3, time.Second*1))

	require.Less(t, time.Since(n), time.Millisecond*200)
}
