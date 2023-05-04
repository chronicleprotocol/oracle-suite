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

func TestDevCircuitBreakerNode(t *testing.T) {
	pair := provider.Pair{Base: "BTC", Quote: "USD"}
	tests := []struct {
		name          string
		priceTick     provider.Tick
		referenceTick provider.Tick
		wantErr       bool
	}{
		{
			name: "below threshold",
			priceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(10),
				Time:  time.Now(),
			},
			referenceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(10.6),
				Time:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "above threshold (lower price than reference)",
			priceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(10),
				Time:  time.Now(),
			},
			referenceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(12),
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "above threshold (higher price than reference)",
			priceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(14),
				Time:  time.Now(),
			},
			referenceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(12),
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			priceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(14),
				Time:  time.Now(),
				Error: errors.New("invalid price"),
			},
			referenceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(12),
				Time:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid reference price",
			priceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(14),
				Time:  time.Now(),
			},
			referenceTick: provider.Tick{
				Pair:  pair,
				Price: bn.Float(14),
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
			priceNode.On("Pair").Return(pair)
			referenceNode.On("Pair").Return(pair)
			priceNode.On("Tick").Return(tt.priceTick)
			referenceNode.On("Tick").Return(tt.referenceTick)

			// Create dev circuit breaker node.
			node := NewDevCircuitBreakerNode(provider.Pair{Base: "BTC", Quote: "USD"}, 0.1)
			require.NoError(t, node.AddBranch(priceNode, referenceNode))

			// Test.
			tick := node.Tick()
			if tt.wantErr {
				assert.Error(t, tick.Validate())
			} else {
				require.NoError(t, tick.Validate())
				assert.Equal(t, tick.Price, tt.priceTick.Price)
			}
		})
	}
}

func TestDevCircuitBreakerNode_AddBranch(t *testing.T) {
	btcusdNode := new(mockNode)
	ethusdNode := new(mockNode)
	btcusdNode.On("Pair").Return(provider.Pair{Base: "BTC", Quote: "USD"})
	ethusdNode.On("Pair").Return(provider.Pair{Base: "ETH", Quote: "USD"})
	tests := []struct {
		name    string
		input   []Node
		wantErr bool
	}{
		{
			name:    "add one branch",
			input:   []Node{btcusdNode},
			wantErr: false,
		},
		{
			name:    "add two branches",
			input:   []Node{btcusdNode, btcusdNode},
			wantErr: false,
		},
		{
			name:    "add three branches",
			input:   []Node{btcusdNode, btcusdNode, btcusdNode},
			wantErr: true,
		},
		{
			name:    "invalid pair",
			input:   []Node{ethusdNode},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewDevCircuitBreakerNode(provider.Pair{Base: "BTC", Quote: "USD"}, 0.1)
			err := node.AddBranch(tt.input...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDevCircuitBreakerNode_Branches(t *testing.T) {
	// Mock nodes.
	priceNode := new(mockNode)
	referenceNode := new(mockNode)
	priceNode.On("Pair").Return(provider.Pair{Base: "BTC", Quote: "USD"})
	referenceNode.On("Pair").Return(provider.Pair{Base: "BTC", Quote: "USD"})

	// Test.
	node := NewDevCircuitBreakerNode(provider.Pair{Base: "BTC", Quote: "USD"}, 0.1)
	require.NoError(t, node.AddBranch(priceNode, referenceNode))
	branches := node.Branches()
	assert.Equal(t, 2, len(branches))
}

func TestDevCircuitBreakerNode_Pair(t *testing.T) {
	node := NewDevCircuitBreakerNode(provider.Pair{Base: "BTC", Quote: "USD"}, 0.1)
	assert.Equal(t, "BTC/USD", node.Pair().String())
}
