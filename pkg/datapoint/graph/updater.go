package graph

import (
	"context"
	"sync"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/origin"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

const UpdaterLoggerTag = "GRAPH_UPDATER"

// maxConcurrentUpdates represents the maximum number of concurrent tick
// fetches from origins.
const maxConcurrentUpdates = 10

// Updater updates the origin nodes using points from the origins.
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
func (u *Updater) Update(ctx context.Context, graphs []Node) {
	nodes, queries := u.identifyNodesToUpdate(graphs)
	u.updateNodesWithDataPoints(nodes, u.fetchDataPoints(ctx, queries))
}

// identifyNodesToUpdate returns the nodes that need to be updated along
// with the pairs needed to fetch the points for those nodes.
func (u *Updater) identifyNodesToUpdate(graphs []Node) (nodesMap, queryMap) {
	nodes := make(nodesMap)
	queries := make(queryMap)
	Walk(func(n Node) {
		if originNode, ok := n.(*OriginNode); ok {
			if originNode.IsFresh() {
				return
			}
			nodes.add(originNode)
			queries.add(originNode)
		}
	}, graphs...)
	return nodes, queries
}

// fetchDataPoints fetches the points for the given pairs from the origins.
//
// DataPoints are fetched asynchronously, number of concurrent fetches is limited by
// the maxConcurrentUpdates constant.
func (u *Updater) fetchDataPoints(ctx context.Context, queries queryMap) dataPointsMap {
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(queries))

	pointsMap := make(dataPointsMap)
	for originName, query := range queries {
		go func(originName string, queries []any) {
			defer wg.Done()

			origin := u.origins[originName]
			if origin == nil {
				return
			}

			// Recover from panics that may occur during fetching pointsMap.
			defer func() {
				if r := recover(); r != nil {
					u.logger.
						WithFields(log.Fields{
							"origin": originName,
							"panic":  r,
						}).
						Error("Panic while fetching data points from the origin")
				}
			}()

			// Limit the number of concurrent updates.
			u.limiter <- struct{}{}
			defer func() { <-u.limiter }()

			// Fetch data points from the origin and store them in the map.
			points, err := origin.FetchDataPoints(ctx, queries)
			if err != nil {
				u.logger.
					WithError(err).
					WithFields(log.Fields{
						"origin": originName,
					}).
					Error("Failed to fetch data points from the origin")
			}
			for query, point := range points {
				mu.Lock()
				pointsMap.add(originName, query, point)
				mu.Unlock()
			}
		}(originName, query)
	}

	wg.Wait()

	return pointsMap
}

// updateNodesWithDataPoints updates the nodes with the given points.
func (u *Updater) updateNodesWithDataPoints(nodes nodesMap, points dataPointsMap) {
	for k, nodes := range nodes {
		point, ok := points[k]
		for _, node := range nodes {
			if !ok {
				u.logger.
					WithFields(log.Fields{
						"origin": k.origin,
						"query":  k.query,
					}).
					Warn("The origin did not return a data point for the query")
				continue
			}
			if err := node.SetDataPoint(point); err != nil {
				u.logger.
					WithFields(log.Fields{
						"origin": k.origin,
						"query":  k.query,
					}).
					WithError(err).
					Warn("Failed to set data point on the origin node")
			}
		}
	}
}

type (
	queryMap      map[string][]any                   // pairs grouped by origin
	nodesMap      map[originQueryKey][]*OriginNode   // nodes grouped by origin and query
	dataPointsMap map[originQueryKey]datapoint.Point // points grouped by origin and query
)

type originQueryKey struct {
	origin string
	query  any
}

func (m queryMap) add(node *OriginNode) {
	m[node.Origin()] = appendIfUnique(m[node.Origin()], node.Query())
}

func (m nodesMap) add(node *OriginNode) {
	originPair := originQueryKey{
		origin: node.Origin(),
		query:  node.Query(),
	}
	m[originPair] = appendIfUnique(m[originPair], node)
}

func (m dataPointsMap) add(origin string, query any, point datapoint.Point) {
	originPair := originQueryKey{
		origin: origin,
		query:  query,
	}
	m[originPair] = point
}

func appendIfUnique[T comparable](slice []T, item T) []T {
	for _, i := range slice {
		if i == item {
			return slice
		}
	}
	return append(slice, item)
}
