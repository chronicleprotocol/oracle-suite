package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// DevCircuitBreakerNode is a circuit breaker that tips if the price deviation
// between two branches is greater than the breaker value.
//
// It must have two branches. First branch is the price branch, second branch
// is the reference branch. Deviation is calculated as follows:
// abs(1.0 - (reference_price / price))
type DevCircuitBreakerNode struct {
	pair            provider.Pair
	priceBranch     Node
	referenceBranch Node
	threshold       float64
}

// NewDevCircuitBreakerNode creates a new DevCircuitBreakerNode instance.
func NewDevCircuitBreakerNode(pair provider.Pair, threshold float64) *DevCircuitBreakerNode {
	return &DevCircuitBreakerNode{
		pair:      pair,
		threshold: threshold,
	}
}

// Branches implements the Node interface.
func (n *DevCircuitBreakerNode) Branches() []Node {
	if n.priceBranch == nil || n.referenceBranch == nil {
		return nil
	}
	return []Node{n.priceBranch, n.referenceBranch}
}

// AddBranch implements the Node interface.
//
// Node requires two branches, first branch is the price branch, second branch
// is the reference branch.
//
// If more than two branches are added, an error is returned.
func (n *DevCircuitBreakerNode) AddBranch(nodes ...Node) error {
	for _, node := range nodes {
		if !node.Pair().Equal(n.pair) {
			return fmt.Errorf("expected pair %s, got %s", n.pair, node.Pair())
		}
	}
	if len(nodes) > 0 && n.priceBranch == nil {
		n.priceBranch = NewWrapperNode(nodes[0], MapMeta{"type": "price"})
		nodes = nodes[1:]
	}
	if len(nodes) > 0 && n.referenceBranch == nil {
		n.referenceBranch = NewWrapperNode(nodes[0], MapMeta{"type": "reference_price"})
		nodes = nodes[1:]
	}
	if len(nodes) > 0 {
		return fmt.Errorf("only two branches are allowed")
	}
	return nil
}

// Pair implements the Node interface.
func (n *DevCircuitBreakerNode) Pair() provider.Pair {
	return n.pair
}

// Tick implements the Node interface.
func (n *DevCircuitBreakerNode) Tick() provider.Tick {
	// Validate branches.
	if n.priceBranch == nil || n.referenceBranch == nil {
		return provider.Tick{
			Pair:  n.pair,
			Error: fmt.Errorf("two branches are required"),
		}
	}
	meta := n.Meta().(MapMeta)
	if err := n.priceBranch.Tick().Validate(); err != nil {
		return provider.Tick{
			Pair:  n.pair,
			Error: fmt.Errorf("invalid price tick: %w", err),
			Meta:  meta,
		}
	}
	if err := n.referenceBranch.Tick().Validate(); err != nil {
		return provider.Tick{
			Pair:  n.pair,
			Error: fmt.Errorf("invalid reference tick: %w", err),
			Meta:  meta,
		}
	}

	// Calculate deviation.
	price := n.priceBranch.Tick()
	reference := n.referenceBranch.Tick()
	deviation := bn.Float(1.0).Sub(reference.Price.Div(price.Price)).Abs().Float64()
	meta["deviation"] = deviation

	// Return tick, if deviation is greater than threshold, add error.
	tick := n.priceBranch.Tick()
	tick.SubTicks = []provider.Tick{price, reference}
	tick.Meta = meta
	if deviation > n.threshold {
		tick.Error = fmt.Errorf("deviation %f is greater than threshold %f", deviation, n.threshold)
	}
	return tick
}

// Meta implements the Node interface.
func (n *DevCircuitBreakerNode) Meta() provider.Meta {
	return MapMeta{"type": "deviation_circuit_breaker", "threshold": n.threshold}
}
