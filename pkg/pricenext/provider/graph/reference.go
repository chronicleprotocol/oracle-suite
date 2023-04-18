package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type ReferenceMeta struct {
	Tick provider.Tick
}

func (m ReferenceMeta) Meta() map[string]any {
	return map[string]any{
		"aggregator": "reference",
		"tick":       m.Tick,
	}
}

// ReferenceNode is a node that references another node.
type ReferenceNode struct {
	pair   provider.Pair
	branch Node
}

func NewReferenceNode(pair provider.Pair) *ReferenceNode {
	return &ReferenceNode{pair: pair}
}

// AddBranch implements the Node interface.
func (i *ReferenceNode) AddBranch(branch ...Node) error {
	if len(branch) == 0 {
		return nil
	}
	if i.branch != nil {
		return fmt.Errorf("branch already exists")
	}
	if len(branch) != 1 {
		return fmt.Errorf("expected 1 branch, got %d", len(branch))
	}
	if !branch[0].Pair().Equal(i.pair) {
		return fmt.Errorf("expected pair %s, got %s", i.pair, branch[0].Pair())
	}
	i.branch = branch[0]
	return nil
}

// Branches implements the Node interface.
func (i *ReferenceNode) Branches() []Node {
	if i.branch == nil {
		return nil
	}
	return []Node{i.branch}
}

// Pair implements the Node interface.
func (i *ReferenceNode) Pair() provider.Pair {
	return i.pair
}

// Tick implements the Node interface.
func (i *ReferenceNode) Tick() provider.Tick {
	if i.branch == nil {
		return provider.Tick{
			Pair:  i.pair,
			Error: fmt.Errorf("branch is not set (this is likely a bug)"),
		}
	}
	tick := i.branch.Tick()
	return provider.Tick{
		Pair:      tick.Pair,
		Price:     tick.Price,
		Volume24h: tick.Volume24h,
		Time:      tick.Time,
		Meta:      &ReferenceMeta{Tick: tick},
		Warning:   tick.Warning,
		Error:     tick.Error,
	}
}
