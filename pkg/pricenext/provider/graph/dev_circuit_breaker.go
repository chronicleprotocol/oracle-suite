package graph

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type DevCircuitBreakerMeta struct {
	Tick          provider.Tick
	ReferenceTick provider.Tick
}

func (m DevCircuitBreakerMeta) Meta() map[string]any {
	return map[string]any{
		"node":           "dev_circuit_breaker",
		"tick":           m.Tick,
		"reference_tick": m.ReferenceTick,
	}
}

// DevCircuitBreaker is a circuit breaker that tips if the price deviation between
// two branches is greater than the breaker value.
//
// First branch is the price branch, second branch is the reference branch.
// Deviation is calculated as abs(1.0 - (reference_price / price))
type DevCircuitBreaker struct {
	pair            provider.Pair
	priceBranch     Node
	referenceBranch Node
	threshold       float64
}

// NewDevCircuitBreakerNode creates a new DevCircuitBreaker instance.
func NewDevCircuitBreakerNode(pair provider.Pair, threshold float64) *DevCircuitBreaker {
	return &DevCircuitBreaker{
		pair:      pair,
		threshold: threshold,
	}
}

// Branches implements the Node interface.
func (d *DevCircuitBreaker) Branches() []Node {
	if d.priceBranch == nil || d.referenceBranch == nil {
		return nil
	}
	return []Node{d.priceBranch, d.referenceBranch}
}

// AddBranch implements the Node interface.
func (d *DevCircuitBreaker) AddBranch(nodes ...Node) error {
	for _, node := range nodes {
		if !node.Pair().Equal(d.pair) {
			return fmt.Errorf("expected pair %s, got %s", d.pair, node.Pair())
		}
	}
	if len(nodes) > 0 && d.priceBranch == nil {
		d.priceBranch = nodes[0]
		nodes = nodes[1:]
	}
	if len(nodes) > 0 && d.referenceBranch == nil {
		d.referenceBranch = nodes[0]
		nodes = nodes[1:]
	}
	if len(nodes) > 0 {
		return fmt.Errorf("only two branches are allowed")
	}
	return nil
}

// Pair implements the Node interface.
func (d *DevCircuitBreaker) Pair() provider.Pair {
	return d.pair
}

// Tick implements the Node interface.
func (d *DevCircuitBreaker) Tick() provider.Tick {
	// Validate branches.
	if d.priceBranch == nil || d.referenceBranch == nil {
		return provider.Tick{
			Pair:  d.pair,
			Error: fmt.Errorf("two branches are required"),
		}
	}
	meta := DevCircuitBreakerMeta{Tick: d.priceBranch.Tick(), ReferenceTick: d.referenceBranch.Tick()}
	if err := d.priceBranch.Tick().Validate(); err != nil {
		return provider.Tick{
			Pair:  d.pair,
			Error: fmt.Errorf("invalid price tick: %w", err),
			Meta:  meta,
		}
	}
	if err := d.referenceBranch.Tick().Validate(); err != nil {
		return provider.Tick{
			Pair:  d.pair,
			Error: fmt.Errorf("invalid reference tick: %w", err),
			Meta:  meta,
		}
	}

	// Calculate deviation.
	price := d.priceBranch.Tick().Price
	reference := d.referenceBranch.Tick().Price
	deviation := bn.Float(1.0).Sub(reference.Div(price)).Abs().Float64()

	// Return tick, if deviation is greater than threshold, add error.
	tick := d.priceBranch.Tick()
	tick.Meta = meta
	if deviation > d.threshold {
		tick.Error = fmt.Errorf("deviation %f is greater than breaker %f", deviation, d.threshold)
	}
	return tick
}
