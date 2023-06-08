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
		expectedValue *bn.FloatNumber
		wantErr       bool
	}{
		{
			name: "one value",
			points: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(1),
						Volume24h: bn.Float(1),
					},
					Time: time.Now(),
				},
			},
			minValues:     1,
			expectedValue: bn.Float(1),
			wantErr:       false,
		},
		{
			name: "two values",
			points: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(1),
						Volume24h: bn.Float(1),
					},
					Time: time.Now(),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(2),
						Volume24h: bn.Float(2),
					},
					Time: time.Now(),
				},
			},
			minValues:     2,
			expectedValue: bn.Float(1.5),
			wantErr:       false,
		},
		{
			name: "three values",
			points: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(1),
						Volume24h: bn.Float(1),
					},
					Time: time.Now(),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(2),
						Volume24h: bn.Float(2),
					},
					Time: time.Now(),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(3),
						Volume24h: bn.Float(3),
					},
					Time: time.Now(),
				},
			},
			minValues:     3,
			expectedValue: bn.Float(2),
			wantErr:       false,
		},
		{
			name: "not enough values",
			points: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(1),
						Volume24h: bn.Float(1),
					},
					Time: time.Now(),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(2),
						Volume24h: bn.Float(2),
					},
					Time: time.Now(),
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
					Value: value.Tick{
						Pair:      value.Pair{Base: "A", Quote: "B"},
						Price:     bn.Float(1),
						Volume24h: bn.Float(1),
					},
					Time: time.Now(),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "B", Quote: "A"},
						Price:     bn.Float(2),
						Volume24h: bn.Float(2),
					},
					Time: time.Now(),
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
				assert.Equal(t, tt.expectedValue.Float64(), point.Value.(value.NumericValue).Number().Float64())
				require.NoError(t, point.Validate())
			}
		})
	}
}
