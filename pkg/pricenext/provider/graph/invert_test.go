package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestInvertNode_AddBranch(t *testing.T) {
	mockNode := new(mockNode)
	mockNode.On("Pair").Return(provider.Pair{Base: "USD", Quote: "BTC"})
	tests := []struct {
		name    string
		input   []Node
		wantErr bool
	}{
		{
			name:    "add single branch",
			input:   []Node{mockNode},
			wantErr: false,
		},
		{
			name:    "add second branch",
			input:   []Node{mockNode, mockNode},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewInvertNode(provider.Pair{Base: "BTC", Quote: "USD"})
			err := node.AddBranch(tt.input...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, node.Branches(), 1)
				assert.Equal(t, tt.input, node.Branches())
			}
		})
	}
}

func TestInvertNode_Branches(t *testing.T) {
	mockNode := new(mockNode)
	mockNode.On("Pair").Return(provider.Pair{Base: "USD", Quote: "BTC"})
	invertNode := NewInvertNode(provider.Pair{Base: "BTC", Quote: "USD"})
	require.NoError(t, invertNode.AddBranch(mockNode))
	branches := invertNode.Branches()
	assert.Equal(t, 1, len(branches))
	assert.Equal(t, mockNode, branches[0])
}

func TestInvertNode_Pair(t *testing.T) {
	node := NewInvertNode(provider.Pair{Base: "BTC", Quote: "USD"})
	pair := node.Pair()
	assert.Equal(t, "BTC", pair.Base)
	assert.Equal(t, "USD", pair.Quote)
}

func TestInvertNode_Tick(t *testing.T) {
	mockNode := new(mockNode)
	mockNode.On("Pair").Return(provider.Pair{Base: "USD", Quote: "BTC"})
	mockNode.On("Tick").Return(provider.Tick{
		Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
		Price:     bn.Float(20000),
		Volume24h: bn.Float(2),
	})
	node := NewInvertNode(provider.Pair{Base: "BTC", Quote: "USD"})
	require.NoError(t, node.AddBranch(mockNode))
	tick := node.Tick()
	assert.Equal(t, "USD", tick.Pair.Base)
	assert.Equal(t, "BTC", tick.Pair.Quote)
	assert.Equal(t, bn.Float(0.00005).Float64(), tick.Price.Float64())
	assert.Equal(t, bn.Float(40000).Float64(), tick.Volume24h.Float64())
}
