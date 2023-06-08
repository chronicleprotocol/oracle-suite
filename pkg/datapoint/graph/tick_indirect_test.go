package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
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
					Value: value.Tick{Pair: value.Pair{Base: "A", Quote: "B"}, Price: bn.Float(1)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "B", Quote: "C"}, Price: bn.Float(2)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "C", Quote: "D"}, Price: bn.Float(3)},
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
					Value: value.Tick{Pair: value.Pair{Base: "A", Quote: "B"}, Price: bn.Float(1)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "B", Quote: "C"}, Price: bn.Float(2)},
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
					Value: value.Tick{Pair: value.Pair{Base: "B", Quote: "A"}, Price: bn.Float(1)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "B", Quote: "C"}, Price: bn.Float(2)},
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
					Value: value.Tick{Pair: value.Pair{Base: "A", Quote: "B"}, Price: bn.Float(1)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "C", Quote: "B"}, Price: bn.Float(2)},
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
					Value: value.Tick{Pair: value.Pair{Base: "B", Quote: "A"}, Price: bn.Float(1)},
					Time:  time.Now(),
				},
				{
					Value: value.Tick{Pair: value.Pair{Base: "C", Quote: "B"}, Price: bn.Float(2)},
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
			assert.Equal(t, tt.expectedPrice, point.Value.(value.NumericValue).Number().Float64())
			if tt.wantErr {
				assert.Error(t, point.Validate())
			} else {
				require.NoError(t, point.Validate())
			}
		})
	}
}
