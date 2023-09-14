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

package feed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestNewTickPrecisionHook(t *testing.T) {
	hook := NewTickPrecisionHook(2, 3)
	assert.Equal(t, uint8(2), hook.maxPricePrec)
	assert.Equal(t, uint8(3), hook.maxVolumePrec)
}

func TestTickPrecisionHook_BeforeSign(t *testing.T) {
	hook := NewTickPrecisionHook(2, 3)
	dp := &datapoint.Point{
		Value: value.Tick{
			Price:     bn.DecFloatPoint("123.456789"),
			Volume24h: bn.DecFloatPoint("987.654321"),
		},
		SubPoints: []datapoint.Point{
			{
				Value: value.Tick{
					Price:     bn.DecFloatPoint("123.456789"),
					Volume24h: bn.DecFloatPoint("987.654321"),
				},
			},
		},
	}

	err := hook.BeforeSign(context.Background(), dp)
	assert.NoError(t, err)

	assert.Equal(t, "123.46", dp.Value.(value.Tick).Price.String())
	assert.Equal(t, "987.654", dp.Value.(value.Tick).Volume24h.String())
	assert.Equal(t, "123.46", dp.SubPoints[0].Value.(value.Tick).Price.String())
	assert.Equal(t, "987.654", dp.SubPoints[0].Value.(value.Tick).Volume24h.String())
}

func TestNewTickTraceHook(t *testing.T) {
	hook := NewTickTraceHook()
	assert.NotNil(t, hook)
}

func TestTickTraceHook_BeforeBroadcast(t *testing.T) {
	hook := NewTickTraceHook()
	dp := &datapoint.Point{
		Meta: map[string]any{
			"type":   "origin",
			"origin": "source1",
		},
		Value: value.Tick{
			Pair:  value.Pair{Base: "BTC", Quote: "USD"},
			Price: bn.DecFloatPoint("123.456789"),
		},
		SubPoints: []datapoint.Point{
			{
				Meta: map[string]any{
					"type":   "origin",
					"origin": "source2",
				},
				Value: value.Tick{
					Pair:  value.Pair{Base: "BTC", Quote: "USD"},
					Price: bn.DecFloatPoint("987.654321"),
				},
			},
		},
	}

	err := hook.BeforeBroadcast(context.Background(), dp)
	assert.NoError(t, err)

	expectedTrace := map[string]string{
		"BTC/USD@source1": "123.456789",
		"BTC/USD@source2": "987.654321",
	}
	assert.Equal(t, expectedTrace, dp.Meta["trace"])
}
