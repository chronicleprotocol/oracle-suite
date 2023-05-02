package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestIndirectNode(t *testing.T) {
	tests := []struct {
		name          string
		ticks         []provider.Tick
		pair          provider.Pair
		expectedPrice float64
		wantErr       bool
	}{
		{
			name: "three nodes",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "B", Quote: "C"}, Time: time.Now()},
				{Price: bn.Float(3), Pair: provider.Pair{Base: "C", Quote: "D"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "D"},
			expectedPrice: 6,
			wantErr:       false,
		},
		{
			name: "A/B->B/C",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "B", Quote: "C"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "C"},
			expectedPrice: 2,
			wantErr:       false,
		},
		{
			name: "B/A->B/C",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "B", Quote: "A"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "B", Quote: "C"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "C"},
			expectedPrice: 2,
			wantErr:       false,
		},
		{
			name: "A/B->C/B",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "C", Quote: "B"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "C"},
			expectedPrice: 0.5,
			wantErr:       false,
		},
		{
			name: "B/A->C/B",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "B", Quote: "A"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "C", Quote: "B"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "C"},
			expectedPrice: 0.5,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create indirect node
			node := NewIndirectNode(tt.pair)

			for _, tick := range tt.ticks {
				n := new(mockNode)
				n.On("Tick").Return(tick)
				require.NoError(t, node.AddBranch(n))
			}

			// Test
			tick := node.Tick()
			assert.Equal(t, tt.expectedPrice, tick.Price.Float64())
			if tt.wantErr {
				assert.Error(t, tick.Validate())
			} else {
				require.NoError(t, tick.Validate())
			}
		})
	}
}
