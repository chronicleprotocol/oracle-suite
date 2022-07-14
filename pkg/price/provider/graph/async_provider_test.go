package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
)

func Test_gcdTTL(t *testing.T) {
	p := provider.Pair{Base: "A", Quote: "B"}
	root := nodes.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix()+10)
	on1 := nodes.NewOriginNode(nodes.OriginPair{Origin: "a", Pair: p}, 12*time.Second, ttl)
	on2 := nodes.NewOriginNode(nodes.OriginPair{Origin: "b", Pair: p}, 6*time.Second, ttl)
	on3 := nodes.NewOriginNode(nodes.OriginPair{Origin: "b", Pair: p}, 10*time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 2*time.Second, gcdTTL([]nodes.Node{root}))
}
