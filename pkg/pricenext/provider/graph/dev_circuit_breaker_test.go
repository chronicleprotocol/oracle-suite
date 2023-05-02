package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestDevCircuitBreakerNode(t *testing.T) {
	tests := []struct {
		name           string
		priceValue     float64
		referenceValue float64
		wantErr        bool
	}{
		{
			name:           "below threshold",
			priceValue:     10,
			referenceValue: 10.5,
			wantErr:        false,
		},
		{
			name:           "above threshold",
			priceValue:     10,
			referenceValue: 12,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock nodes
			pair := provider.Pair{Base: "BTC", Quote: "USD"}
			priceNode := new(mockNode)
			referenceNode := new(mockNode)
			priceNode.On("Pair").Return(pair)
			referenceNode.On("Pair").Return(pair)
			priceNode.On("Tick").Return(provider.Tick{
				Pair:  pair,
				Price: bn.Float(tt.priceValue),
				Time:  time.Now(),
			})
			referenceNode.On("Tick").Return(provider.Tick{
				Pair:  pair,
				Price: bn.Float(tt.referenceValue),
				Time:  time.Now(),
			})

			// Create dev circuit breaker node
			node := NewDevCircuitBreakerNode(provider.Pair{Base: "BTC", Quote: "USD"}, 0.1)
			require.NoError(t, node.AddBranch(priceNode, referenceNode))

			// Test
			tick := node.Tick()
			assert.Equal(t, tick.Price.Float64(), tt.priceValue)
			if tt.wantErr {
				assert.Error(t, tick.Validate())
			} else {
				require.NoError(t, tick.Validate())
			}
		})
	}
}

func TestDevCircuitBreakerNode_AddBranch(t *testing.T) {
	mockNode := new(mockNode)
	mockNode.On("Pair").Return(provider.Pair{Base: "BTC", Quote: "USD"})

	tests := []struct {
		name    string
		input   []Node
		wantErr bool
	}{
		{
			name:    "add one branch",
			input:   []Node{mockNode},
			wantErr: false,
		},
		{
			name:    "add two branches",
			input:   []Node{mockNode, mockNode},
			wantErr: false,
		},
		{
			name:    "add three branches",
			input:   []Node{mockNode, mockNode, mockNode},
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
