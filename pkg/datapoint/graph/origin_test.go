package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestOriginNode(t *testing.T) {
	tests := []struct {
		name               string
		dataPoint          datapoint.Point
		query              any
		freshnessThreshold time.Duration
		expiryThreshold    time.Duration
		expectedFresh      bool
		expectedExpired    bool
		wantErr            bool
	}{
		{
			name:               "valid",
			dataPoint:          datapoint.Point{Value: numericValue{x: bn.Float(1)}, Time: time.Now()},
			query:              "query",
			freshnessThreshold: time.Minute,
			expiryThreshold:    time.Minute,
			expectedFresh:      true,
			expectedExpired:    false,
			wantErr:            false,
		},
		{
			name:               "not fresh",
			dataPoint:          datapoint.Point{Value: numericValue{x: bn.Float(1)}, Time: time.Now().Add(-30 * time.Second)},
			query:              "query",
			freshnessThreshold: time.Second * 20,
			expiryThreshold:    time.Second * 40,
			expectedFresh:      false,
			expectedExpired:    false,
			wantErr:            false,
		},
		{
			name:               "expired",
			dataPoint:          datapoint.Point{Value: numericValue{x: bn.Float(1)}, Time: time.Now().Add(-60 * time.Second)},
			query:              "query",
			freshnessThreshold: time.Second * 20,
			expiryThreshold:    time.Second * 40,
			expectedFresh:      false,
			expectedExpired:    true,
			wantErr:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create origin node
			node := NewOriginNode("origin", tt.query, tt.freshnessThreshold, tt.expiryThreshold)

			// Test
			assert.Equal(t, tt.query, node.Query())
			if tt.wantErr {
				assert.Error(t, node.SetDataPoint(tt.dataPoint))
			} else {
				require.NoError(t, node.SetDataPoint(tt.dataPoint))
				assert.Equal(t, tt.expectedFresh, node.IsFresh())
				assert.Equal(t, tt.expectedExpired, node.IsExpired())
				if tt.expectedExpired {
					dataPoint := node.DataPoint()
					require.Error(t, dataPoint.Validate())
				}
			}
		})
	}
}
