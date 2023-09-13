package graph

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestTickMedianNode(t *testing.T) {
	tests := []struct {
		name          string
		points        []datapoint.Point
		minValues     int
		expectedValue *bn.DecFloatPointNumber
		wantErr       bool
	}{
		{
			name: "one value",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 1),
					Time:  time.Now(),
				},
			},
			minValues:     1,
			expectedValue: bn.DecFloatPoint(1),
			wantErr:       false,
		},
		{
			name: "two values",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 1),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 2, 2),
					Time:  time.Now(),
				},
			},
			minValues:     2,
			expectedValue: bn.DecFloatPoint(1.5),
			wantErr:       false,
		},
		{
			name: "three values",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 1),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 2, 2),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 3, 3),
					Time:  time.Now(),
				},
			},
			minValues:     3,
			expectedValue: bn.DecFloatPoint(2),
			wantErr:       false,
		},
		{
			name: "not enough values",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 1),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 2, 2),
					Time:  time.Now(),
				},
				{
					Time:  time.Now(),
					Error: errors.New("error"),
				},
			},
			minValues: 3,
			wantErr:   true,
		},
		{
			name: "different pairs",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 1),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "A"}, 2, 2),
					Time:  time.Now(),
				},
			},
			minValues: 2,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create indirect node
			node := NewTickMedianNode(tt.minValues)

			for _, dataPoint := range tt.points {
				n := new(mockNode)
				n.On("DataPoint").Return(dataPoint)
				require.NoError(t, node.AddNodes(n))
			}

			// Test
			point := node.DataPoint()
			if tt.wantErr {
				assert.Error(t, point.Validate())
			} else {
				expValue, _ := tt.expectedValue.BigFloat().Float64()
				value, _ := point.Value.(value.NumericValue).Number().BigFloat().Float64()
				assert.Equal(t, expValue, value)
				require.NoError(t, point.Validate())
			}
		})
	}
}
