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

package store

import (
	"context"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
)

func TestMemoryStorage_Add(t *testing.T) {
	ctx := context.Background()
	addr := types.Address{}
	model := "model1"
	point := datapoint.Point{Time: time.Now()}
	oldPoint := datapoint.Point{Time: time.Now().Add(-time.Hour)}

	t.Run("adding first point", func(t *testing.T) {
		storage := NewMemoryStorage()

		err := storage.Add(ctx, addr, model, point)
		require.NoError(t, err)

		_, exists := storage.ds[dataPointKey{feed: addr, model: model}]
		require.True(t, exists)
	})
	t.Run("adding older point", func(t *testing.T) {
		storage := NewMemoryStorage()
		err := storage.Add(ctx, addr, model, point)
		require.NoError(t, err)

		err = storage.Add(ctx, addr, model, oldPoint) // should be ignored
		require.NoError(t, err)

		storedPoint, _ := storage.ds[dataPointKey{feed: addr, model: model}]
		require.Equal(t, point, storedPoint)
	})
}

func TestMemoryStorage_LatestFrom(t *testing.T) {
	ctx := context.Background()
	addr := types.Address{}
	model := "model1"
	point := datapoint.Point{Time: time.Now()}

	t.Run("point exists", func(t *testing.T) {
		storage := NewMemoryStorage()
		err := storage.Add(ctx, addr, model, point)
		require.NoError(t, err)

		retPoint, ok, err := storage.LatestFrom(ctx, addr, model)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, point, retPoint)
	})
	t.Run("point does not exist", func(t *testing.T) {
		storage := NewMemoryStorage()

		_, ok, err := storage.LatestFrom(ctx, addr, model)
		require.NoError(t, err)
		require.False(t, ok)
	})
}

func TestMemoryStorage_Latest(t *testing.T) {
	ctx := context.Background()
	addr := types.Address{}
	model := "model"
	point := datapoint.Point{Time: time.Now()}

	t.Run("model exists", func(t *testing.T) {
		storage := NewMemoryStorage()
		err := storage.Add(ctx, addr, model, point)
		require.NoError(t, err)

		points, err := storage.Latest(ctx, model)
		require.NoError(t, err)
		require.Len(t, points, 1)
		require.Equal(t, point, points[addr])
	})
	t.Run("model does not exist", func(t *testing.T) {
		storage := NewMemoryStorage()

		points, err := storage.Latest(ctx, model)
		require.NoError(t, err)
		require.Empty(t, points)
	})
}
