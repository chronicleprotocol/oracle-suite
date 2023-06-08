package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestTickAliasNode_DataPoint(t *testing.T) {
	mockNode := new(mockNode)
	mockNode.On("DataPoint").Return(datapoint.Point{
		Value: value.Tick{
			Pair:      value.Pair{Base: "BTC", Quote: "USDC"},
			Price:     bn.Float(20000),
			Volume24h: bn.Float(2),
		},
	})
	node := NewTickAliasNode(value.Pair{Base: "BTC", Quote: "USD"})
	require.NoError(t, node.AddNodes(mockNode))
	tick := node.DataPoint().Value.(value.Tick)
	assert.Equal(t, "BTC", tick.Pair.Base)
	assert.Equal(t, "USD", tick.Pair.Quote)
	assert.Equal(t, bn.Float(20000).Float64(), tick.Price.Float64())
	assert.Equal(t, bn.Float(2).Float64(), tick.Volume24h.Float64())
}

func TestTickAliasNode_AddNodes(t *testing.T) {
	node := new(mockNode)
	tests := []struct {
		name    string
		input   []Node
		wantErr bool
	}{
		{
			name:    "add single node",
			input:   []Node{node},
			wantErr: false,
		},
		{
			name:    "add second node",
			input:   []Node{node, node},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTickAliasNode(value.Pair{Base: "BTC", Quote: "USD"})
			err := node.AddNodes(tt.input...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, node.Nodes(), 1)
				assert.Equal(t, tt.input, node.Nodes())
			}
		})
	}
}
