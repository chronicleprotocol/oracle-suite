package graph

import (
	"fmt"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
)

type ErrOriginTickExpired struct{ Tick provider.Tick }

func (e ErrOriginTickExpired) Error() string {
	return fmt.Sprintf("tick expired: %s", e.Tick)
}

type ErrOriginInvalidTick struct{ Tick provider.Tick }

func (e ErrOriginInvalidTick) Error() string {
	return fmt.Sprintf("unable to update origin node: invalid tick %s", e.Tick)
}

type ErrOriginIncompatiblePair struct{ Pair provider.Pair }

func (e ErrOriginIncompatiblePair) Error() string {
	return fmt.Sprintf("unable to update origin node: incompatible pair %s", e.Pair)
}

type ErrOriginTickTooOld struct{ Tick provider.Tick }

func (e ErrOriginTickTooOld) Error() string {
	return fmt.Sprintf("unable to update origin node: tick too old %s", e.Tick)
}

type OriginMeta struct {
	// Origin is an origin name.
	Origin string
}

func (m OriginMeta) Meta() map[string]any {
	return map[string]any{
		"origin": m.Origin,
	}
}

// OriginNode is a node that provides a tick for a given asset pair from a
// specific origin.
type OriginNode struct {
	mu sync.RWMutex

	origin string
	pair   provider.Pair
	tick   provider.Tick
	warn   error

	// freshnessThreshold describes the duration within which the price is
	// considered fresh, and an update can be skipped.
	freshnessThreshold time.Duration

	// expiryThreshold describes the duration after which the price is
	// considered expired, and an update is required.
	expiryThreshold time.Duration
}

// NewOriginNode creates a new OriginNode instance.
//
// The freshnessThreshold and expiryThreshold arguments are used to determine
// whether the price is fresh or expired.
//
// The price is considered fresh if it was updated within the freshnessThreshold
// duration. In this case, the price update is not required.
//
// The price is considered expired if it was updated more than the expiryThreshold
// duration ago. In this case, the price is considered invalid and an update is
// required.
//
// There must be a gap between the freshnessThreshold and expiryThreshold so that
// the price will be updated before it is considered expired.
//
// Note that price that is considered not fresh may not be considered expired.
func NewOriginNode(origin string, pair provider.Pair, freshnessThreshold, expiryThreshold time.Duration) *OriginNode {
	return &OriginNode{
		origin:             origin,
		pair:               pair,
		tick:               provider.Tick{Pair: pair},
		freshnessThreshold: freshnessThreshold,
		expiryThreshold:    expiryThreshold,
	}
}

// AddBranch implements the Node interface.
func (n *OriginNode) AddBranch(branch ...Node) error {
	if len(branch) > 0 {
		return fmt.Errorf("origin node cannot have branches")
	}
	return nil
}

// Branches implements the Node interface.
func (n *OriginNode) Branches() []Node {
	return nil
}

// Origin returns the origin name.
func (n *OriginNode) Origin() string {
	return n.origin
}

// Pair implements the Node interface.
func (n *OriginNode) Pair() provider.Pair {
	return n.pair
}

// Tick implements the Node interface.
func (n *OriginNode) Tick() provider.Tick {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.tick.Error != nil {
		return n.tick
	}
	if n.IsExpired() {
		n.tick.Error = ErrOriginTickExpired{Tick: n.tick}
	}
	return n.tick
}

// Warning returns the warning associated with the node.
func (n *OriginNode) Warning() error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.warn
}

// SetTick sets the node tick.
//
// Tick is updated only if the new tick is valid, is not older than the current
// tick, and has the same pair as the node.
//
// Meta field of the given tick is ignored and replaced with the origin name.
// It returns an error if the given price is incompatible with the node.
func (n *OriginNode) SetTick(tick provider.Tick) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.Pair().Equal(tick.Pair) {
		return ErrOriginIncompatiblePair{Pair: tick.Pair}
	}
	if err := tick.Validate(); err != nil {
		return ErrOriginInvalidTick{Tick: tick}
	}
	if n.tick.Time.After(tick.Time) {
		return ErrOriginTickTooOld{Tick: tick}
	}
	tick.Meta = OriginMeta{Origin: n.origin}
	n.tick = tick
	return nil
}

// SetWarning sets the warning associated with the node.
func (n *OriginNode) SetWarning(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.warn = err
}

// IsFresh returns true if the price is considered fresh, that is, the price
// update is not required.
//
// Note, that the price that is not fresh is not necessarily expired.
func (n *OriginNode) IsFresh() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.tick.Time.Add(n.freshnessThreshold).After(time.Now())
}

// IsExpired returns true if the price is considered expired.
func (n *OriginNode) IsExpired() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.tick.Time.Add(-n.expiryThreshold).After(time.Now())
}
