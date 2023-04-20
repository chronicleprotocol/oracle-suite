package graph

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
)

// ErrMissingTick is returned when an origin did not return a tick for a pair.
var ErrMissingTick = errors.New("origin did not return a tick")

// maxConcurrentUpdates represents the maximum number of concurrent tick
// fetches from origins.
const maxConcurrentUpdates = 10

// Updater updates the origin nodes using ticks from the origins.
type Updater struct {
	origins map[string]origin.Origin
	limiter chan struct{}
}

// NewUpdater returns a new Updater instance.
func NewUpdater(origins map[string]origin.Origin) *Updater {
	return &Updater{
		origins: origins,
		limiter: make(chan struct{}, maxConcurrentUpdates),
	}
}

// Update updates the origin nodes in the given graphs.
//
// Only origin nodes that are not fresh will be updated.
func (u *Updater) Update(ctx context.Context, graphs []Node) error {
	nodes, pairs := u.identifyNodesAndPairsToUpdate(graphs)
	ticks, err := u.fetchTicksForPairs(ctx, pairs)
	if err != nil {
		return err
	}
	u.updateNodesWithTicks(nodes, ticks)
	return nil
}

// identifyNodesAndPairsToUpdate returns the nodes that need to be updated along
// with the pairs needed to fetch the ticks for those nodes.
func (u *Updater) identifyNodesAndPairsToUpdate(graphs []Node) (nodesMap, pairsMap) {
	nodes := make(nodesMap)
	pairs := make(pairsMap)
	Walk(func(n Node) {
		if originNode, ok := n.(*OriginNode); ok {
			if originNode.IsFresh() {
				return
			}
			nodes.add(originNode)
			pairs.add(originNode)
		}
	}, graphs...)
	return nodes, pairs
}

// fetchTicksForPairs fetches the ticks for the given pairs from the origins.
//
// Ticks are fetched asynchronously, number of concurrent fetches is limited by
// the maxConcurrentUpdates constant.
func (u *Updater) fetchTicksForPairs(ctx context.Context, pairs pairsMap) (ticksMap, error) {
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(pairs))

	ticks := make(ticksMap)
	for originName, pairs := range pairs {
		go func(originName string, pairs []provider.Pair) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					for _, pair := range pairs {
						ticks.add(originName, provider.Tick{
							Pair:  pair,
							Error: fmt.Errorf("PANIC: %v", r),
						})
					}
					mu.Unlock()
				}
			}()

			origin := u.origins[originName]
			if origin == nil {
				return
			}

			// Limit the number of concurrent updates.
			u.limiter <- struct{}{}
			defer func() { <-u.limiter }()

			// Fetch ticks from the origin and store them in the map.
			for _, tick := range origin.FetchTicks(ctx, pairs) {
				mu.Lock()
				ticks.add(originName, tick)
				mu.Unlock()
			}
		}(originName, pairs)
	}

	wg.Wait()

	return ticks, nil
}

// updateNodesWithTicks updates the nodes with the given ticks.
//
// If a tick is missing for a node, the ErrMissingTick error will be set on the
// node as a warning.
func (u *Updater) updateNodesWithTicks(nodes nodesMap, ticks ticksMap) {
	for op, nodes := range nodes {
		tick, ok := ticks[op]
		for _, node := range nodes {
			if !ok {
				node.SetWarning(ErrMissingTick)
				continue
			}
			if err := node.SetTick(tick); err != nil {
				node.SetWarning(err)
			}
		}
	}
}

type (
	pairsMap map[string][]provider.Pair      // pairs grouped by origin
	nodesMap map[originPairKey][]*OriginNode // nodes grouped by origin and pair
	ticksMap map[originPairKey]provider.Tick // ticks grouped by origin and pair
)

type originPairKey struct {
	origin string
	pair   provider.Pair
}

func (m pairsMap) add(node *OriginNode) {
	m[node.Origin()] = appendIfUnique(m[node.Origin()], node.FetchPair())
}

func (m nodesMap) add(node *OriginNode) {
	originPair := originPairKey{
		origin: node.Origin(),
		pair:   node.FetchPair(),
	}
	m[originPair] = appendIfUnique(m[originPair], node)
}

func (m ticksMap) add(origin string, tick provider.Tick) {
	originPair := originPairKey{
		origin: origin,
		pair:   tick.Pair,
	}
	m[originPair] = tick
}

func appendIfUnique[T comparable](slice []T, item T) []T {
	for _, i := range slice {
		if i == item {
			return slice
		}
	}
	return append(slice, item)
}
