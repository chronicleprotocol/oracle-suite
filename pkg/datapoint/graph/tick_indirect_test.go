package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

func TestTickIndirectNode(t *testing.T) {
	tests := []struct {
		name          string
		points        []datapoint.Point
		pair          value.Pair
		expectedPrice float64
		wantErr       bool
	}{
		{
			name: "three nodes",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "C"}, 2, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "C", Quote: "D"}, 3, 0),
					Time:  time.Now(),
				},
			},
			pair:          value.Pair{Base: "A", Quote: "D"},
			expectedPrice: 6,
			wantErr:       false,
		},
		{
			name: "A/B->B/C",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "C"}, 2, 0),
					Time:  time.Now(),
				},
			},
			pair:          value.Pair{Base: "A", Quote: "C"},
			expectedPrice: 2,
			wantErr:       false,
		},
		{
			name: "B/A->B/C",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "A"}, 1, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "C"}, 2, 0),
					Time:  time.Now(),
				},
			},
			pair:          value.Pair{Base: "A", Quote: "C"},
			expectedPrice: 2,
			wantErr:       false,
		},
		{
			name: "A/B->C/B",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "A", Quote: "B"}, 1, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "C", Quote: "B"}, 2, 0),
					Time:  time.Now(),
				},
			},
			pair:          value.Pair{Base: "A", Quote: "C"},
			expectedPrice: 0.5,
			wantErr:       false,
		},
		{
			name: "B/A->C/B",
			points: []datapoint.Point{
				{
					Value: value.NewTick(value.Pair{Base: "B", Quote: "A"}, 1, 0),
					Time:  time.Now(),
				},
				{
					Value: value.NewTick(value.Pair{Base: "C", Quote: "B"}, 2, 0),
					Time:  time.Now(),
				},
			},
			pair:          value.Pair{Base: "A", Quote: "C"},
			expectedPrice: 0.5,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create indirect node
			node := NewTickIndirectNode()

			for _, dataPoint := range tt.points {
				n := new(mockNode)
				n.On("DataPoint").Return(dataPoint)
				require.NoError(t, node.AddNodes(n))
			}

			// Test
			point := node.DataPoint()
			price, _ := point.Value.(value.NumericValue).Number().BigFloat().Float64()
			assert.Equal(t, tt.expectedPrice, price)
			if tt.wantErr {
				assert.Error(t, point.Validate())
			} else {
				require.NoError(t, point.Validate())
			}
		})
	}
}
