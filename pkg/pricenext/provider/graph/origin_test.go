package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestOriginNode(t *testing.T) {
	tests := []struct {
		name               string
		tick               provider.Tick
		pair               provider.Pair
		fetchPair          provider.Pair
		freshnessThreshold time.Duration
		expiryThreshold    time.Duration
		expectedFresh      bool
		expectedExpired    bool
		wantErr            bool
	}{
		{
			name:               "valid",
			tick:               provider.Tick{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
			pair:               provider.Pair{Base: "A", Quote: "B"},
			fetchPair:          provider.Pair{Base: "A", Quote: "B"},
			freshnessThreshold: time.Minute,
			expiryThreshold:    time.Minute,
			expectedFresh:      true,
			expectedExpired:    false,
			wantErr:            false,
		},
		{
			name:               "not fresh",
			tick:               provider.Tick{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now().Add(-30 * time.Second)},
			pair:               provider.Pair{Base: "A", Quote: "B"},
			fetchPair:          provider.Pair{Base: "A", Quote: "B"},
			freshnessThreshold: time.Second * 20,
			expiryThreshold:    time.Second * 40,
			expectedFresh:      false,
			expectedExpired:    false,
			wantErr:            false,
		},
		{
			name:               "expired",
			tick:               provider.Tick{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now().Add(-60 * time.Second)},
			pair:               provider.Pair{Base: "A", Quote: "B"},
			fetchPair:          provider.Pair{Base: "A", Quote: "B"},
			freshnessThreshold: time.Second * 20,
			expiryThreshold:    time.Second * 40,
			expectedFresh:      false,
			expectedExpired:    true,
			wantErr:            false,
		},
		{
			name:               "wrong pair",
			tick:               provider.Tick{Price: bn.Float(1), Pair: provider.Pair{Base: "C", Quote: "D"}, Time: time.Now()},
			pair:               provider.Pair{Base: "A", Quote: "B"},
			fetchPair:          provider.Pair{Base: "A", Quote: "B"},
			freshnessThreshold: time.Minute,
			expiryThreshold:    time.Minute,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create origin node
			node := NewOriginNode("origin", tt.pair, tt.fetchPair, tt.freshnessThreshold, tt.expiryThreshold)

			// Test
			assert.Equal(t, tt.pair, node.Pair())
			assert.Equal(t, tt.fetchPair, node.FetchPair())
			if tt.wantErr {
				assert.Error(t, node.SetTick(tt.tick))
			} else {
				require.NoError(t, node.SetTick(tt.tick))
				assert.Equal(t, tt.expectedFresh, node.IsFresh())
				assert.Equal(t, tt.expectedExpired, node.IsExpired())
				if tt.expectedExpired {
					tick := node.Tick()
					require.Error(t, tick.Validate())
				}
			}
		})
	}
}
