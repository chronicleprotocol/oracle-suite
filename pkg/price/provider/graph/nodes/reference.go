package nodes

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

// ReferenceNode is a node that points to another node.
type ReferenceNode struct {
	reference Node
}

func NewReferenceNode() *ReferenceNode {
	return &ReferenceNode{}
}

func (n *ReferenceNode) SetReference(reference Node) {
	n.reference = reference
}

// Children implements the Node interface.
func (n *ReferenceNode) Children() []Node {
	return []Node{n.reference}
}

func (n *ReferenceNode) Pair() provider.Pair {
	switch typedNode := n.reference.(type) {
	case Origin:
		return typedNode.OriginPair().Pair
	case Aggregator:
		return typedNode.Pair()
	}
	panic("unsupported node")
}

func (n *ReferenceNode) Price() AggregatorPrice {
	switch typedNode := n.reference.(type) {
	case Origin:
		price := typedNode.Price()
		return AggregatorPrice{
			PairPrice: PairPrice{
				Pair:      price.Pair,
				Price:     price.Price,
				Bid:       price.Bid,
				Ask:       price.Ask,
				Volume24h: price.Volume24h,
				Time:      price.Time,
			},
			OriginPrices:     []OriginPrice{price},
			AggregatorPrices: nil,
			Parameters:       nil,
			Error:            price.Error,
		}
	case Aggregator:
		price := typedNode.Price()
		return AggregatorPrice{
			PairPrice: PairPrice{
				Pair:      price.Pair,
				Price:     price.Price,
				Bid:       price.Bid,
				Ask:       price.Ask,
				Volume24h: price.Volume24h,
				Time:      price.Time,
			},
			OriginPrices:     nil,
			AggregatorPrices: []AggregatorPrice{price},
			Parameters:       nil,
			Error:            price.Error,
		}
	}
	return AggregatorPrice{Error: fmt.Errorf("unsupported node")}
}
