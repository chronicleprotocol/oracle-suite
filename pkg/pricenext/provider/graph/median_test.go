package graph

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestMedianNode(t *testing.T) {
	tests := []struct {
		name          string
		ticks         []provider.Tick
		pair          provider.Pair
		min           int
		expectedPrice float64
		wantErr       bool
	}{
		{
			name: "two ticks",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "B"},
			min:           2,
			expectedPrice: 1.5,
			wantErr:       false,
		},
		{
			name: "three ticks",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(3), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
			},
			pair:          provider.Pair{Base: "A", Quote: "B"},
			min:           3,
			expectedPrice: 2,
			wantErr:       false,
		},
		{
			name: "not enough ticks",
			ticks: []provider.Tick{
				{Price: bn.Float(1), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(2), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now()},
				{Price: bn.Float(3), Pair: provider.Pair{Base: "A", Quote: "B"}, Time: time.Now(), Error: errors.New("err")},
			},
			pair:          provider.Pair{Base: "A", Quote: "B"},
			min:           3,
			expectedPrice: 2,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create median node
			node := NewMedianNode(tt.pair, tt.min)

			for _, tick := range tt.ticks {
				n := new(mockNode)
				n.On("Tick").Return(tick)
				require.NoError(t, node.AddBranch(n))
			}

			// Test
			tick := node.Tick()
			if tt.wantErr {
				assert.Error(t, tick.Validate())
			} else {
				assert.Equal(t, tt.expectedPrice, tick.Price.Float64())
				require.NoError(t, tick.Validate())
			}
		})
	}
}
