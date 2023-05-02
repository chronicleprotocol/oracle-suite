package graph

import (
	"context"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
)

const UpdaterLoggerTag = "GRAPH_UPDATER"

// maxConcurrentUpdates represents the maximum number of concurrent tick
// fetches from origins.
const maxConcurrentUpdates = 10

// Updater updates the origin nodes using ticks from the origins.
type Updater struct {
	origins map[string]origin.Origin
	limiter chan struct{}
	logger  log.Logger
}

// NewUpdater returns a new Updater instance.
func NewUpdater(origins map[string]origin.Origin, logger log.Logger) *Updater {
	if logger == nil {
		logger = null.New()
	}
	return &Updater{
		origins: origins,
		limiter: make(chan struct{}, maxConcurrentUpdates),
		logger:  logger.WithField("tag", UpdaterLoggerTag),
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

			origin := u.origins[originName]
			if origin == nil {
				return
			}

			// Recover from panics that may occur during fetching ticks.
			defer func() {
				if r := recover(); r != nil {
					u.logger.
						WithFields(log.Fields{
							"origin": originName,
							"panic":  r,
						}).
						Error("Panic while fetching ticks")
				}
			}()

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
				u.logger.
					WithFields(log.Fields{
						"origin": op.origin,
						"pair":   op.pair,
					}).
					Warn("Origin did not return a tick for pair")
				continue
			}
			if err := node.SetTick(tick); err != nil {
				u.logger.
					WithFields(log.Fields{
						"origin": op.origin,
						"pair":   op.pair,
					}).
					WithError(err).
					Warn("Failed to set tick on origin node")
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
