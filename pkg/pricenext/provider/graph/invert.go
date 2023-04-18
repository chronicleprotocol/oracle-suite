package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type InvertMeta struct {
	// Tick is a tick used to calculate inverted price.
	Tick []provider.Tick
}

func (m InvertMeta) Meta() map[string]any {
	return map[string]any{
		"aggregator": "invert",
		"tick":       m.Tick,
	}
}

// InvertNode is a node that inverts the pair and price. E.g. if the pair is
// BTC/USD and the price is 1000, then the pair will be USD/BTC and the price
// will be 0.001.
type InvertNode struct {
	pair   provider.Pair
	branch Node
}

func NewInvertNode(pair provider.Pair) *InvertNode {
	return &InvertNode{pair: pair}
}

// AddBranch implements the Node interface.
func (i *InvertNode) AddBranch(branch ...Node) error {
	if i.branch != nil {
		return fmt.Errorf("branch already exists")
	}
	if len(branch) != 1 {
		return fmt.Errorf("expected 1 branch, got %d", len(branch))
	}
	if branch[0].Pair().Equal(i.pair.Invert()) {
		return fmt.Errorf("expected pair %s, got %s", i.pair, branch[0].Pair())
	}
	i.branch = branch[0]
	return nil
}

// Branches implements the Node interface.
func (i *InvertNode) Branches() []Node {
	return []Node{i.branch}
}

// Pair implements the Node interface.
func (i *InvertNode) Pair() provider.Pair {
	pair := i.branch.Pair()
	return provider.Pair{
		Base:  pair.Quote,
		Quote: pair.Base,
	}
}

// Tick implements the Node interface.
func (i *InvertNode) Tick() provider.Tick {
	tick := i.branch.Tick()
	return provider.Tick{
		Pair:      tick.Pair.Invert(),
		Price:     tick.Price.Inv(),
		Volume24h: tick.Volume24h.Div(tick.Price),
		Time:      tick.Time,
		Meta:      &InvertMeta{Tick: []provider.Tick{tick}},
		Warning:   tick.Warning,
		Error:     tick.Error,
	}
}
