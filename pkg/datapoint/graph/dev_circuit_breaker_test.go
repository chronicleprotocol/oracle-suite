package graph

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestDevCircuitBreakerNode(t *testing.T) {
	tests := []struct {
		name               string
		priceDataPoint     datapoint.Point
		referenceDataPoint datapoint.Point
		wantErr            bool
	}{
		{
			name: "below threshold",
			priceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(10)},
				Time:  time.Now(),
			},
			referenceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(10.6)},
				Time:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "above threshold (lower price than reference)",
			priceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(10)},
				Time:  time.Now(),
			},
			referenceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(12)},
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "above threshold (higher price than reference)",
			priceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(14)},
				Time:  time.Now(),
			},
			referenceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(12)},
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			priceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(14)},
				Time:  time.Now(),
				Error: errors.New("invalid price"),
			},
			referenceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(12)},
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid reference price",
			priceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(14)},
				Time:  time.Now(),
			},
			referenceDataPoint: datapoint.Point{
				Value: numericValue{bn.Float(14)},
				Time:  time.Now(),
				Error: errors.New("invalid reference price"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock nodes.
			priceNode := new(mockNode)
			referenceNode := new(mockNode)
			thresholdNode := new(mockNode)
			priceNode.On("DataPoint").Return(tt.priceDataPoint)
			referenceNode.On("DataPoint").Return(tt.referenceDataPoint)
			thresholdNode.On("DataPoint").Return(datapoint.Point{
				Value: numericValue{bn.Float(0.1)},
				Time:  time.Now(),
			})

			// Create dev circuit breaker node.
			node := NewDevCircuitBreakerNode()
			require.NoError(t, node.AddNodes(priceNode, referenceNode, thresholdNode))

			// Test.
			point := node.DataPoint()
			if tt.wantErr {
				assert.Error(t, point.Validate())
			} else {
				require.NoError(t, point.Validate())
				assert.Equal(t, point.Value, tt.priceDataPoint.Value)
			}
		})
	}
}

func TestDevCircuitBreakerNode_AddNode(t *testing.T) {
	node := new(mockNode)
	tests := []struct {
		name    string
		input   []Node
		wantErr bool
	}{
		{
			name:    "add one node",
			input:   []Node{node},
			wantErr: false,
		},
		{
			name:    "add two nodes",
			input:   []Node{node, node},
			wantErr: false,
		},
		{
			name:    "add three nodes",
			input:   []Node{node, node, node},
			wantErr: false,
		},
		{
			name:    "add four nodes",
			input:   []Node{node, node, node, node},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewDevCircuitBreakerNode()
			err := node.AddNodes(tt.input...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
