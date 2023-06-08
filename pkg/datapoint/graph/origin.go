package graph

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

// OriginNode is a node that provides a data point for a specific origin.
type OriginNode struct {
	mu sync.RWMutex

	origin    string
	query     any
	dataPoint datapoint.Point

	// freshnessThreshold describes the duration within which the price is
	// considered fresh, and an update can be skipped.
	freshnessThreshold time.Duration

	// expiryThreshold describes the duration after which the price is
	// considered expired, and an update is required.
	expiryThreshold time.Duration
}

// NewOriginNode creates a new OriginNode instance.
//
// The query argument is used to fetch a data point from the origin.
//
// The freshnessThreshold and expiryThreshold arguments are used to determine
// whether a data point is fresh or expired.
//
// A data point is considered fresh if it was updated within the freshnessThreshold
// duration. In this case, the data point update is not required.
//
// A data point is considered expired if it was updated more than the expiryThreshold
// duration ago. In this case, the data point is considered invalid and an update is
// required.
//
// There must be a gap between the freshnessThreshold and expiryThreshold so that
// a data point will be updated before it is considered expired.
func NewOriginNode(
	origin string,
	query any,
	freshnessThreshold time.Duration,
	expiryThreshold time.Duration,
) *OriginNode {

	return &OriginNode{
		origin:             origin,
		query:              query,
		dataPoint:          datapoint.Point{Error: errors.New("data point is not set")},
		freshnessThreshold: freshnessThreshold,
		expiryThreshold:    expiryThreshold,
	}
}

// SetDataPoint sets the data point.
//
// If the provided data point is older than the current data point, it will be
// ignored and an error will be returned.
//
// If the provided data point is invalid, it will be ignored as long as the
// current data point is still valid. This is done to prevent from updating
// valid data points with invalid ones. In this case, the error will be
// returned.
func (n *OriginNode) SetDataPoint(point datapoint.Point) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if point.Time.After(point.Time) {
		return fmt.Errorf("unable to set data point: data point is older than the current data point")
	}
	if n.validate() {
		if err := point.Validate(); err != nil {
			return fmt.Errorf("unable to set data point: %w", err)
		}
	}
	point.Meta = maputil.Merge(point.Meta, n.Meta())
	n.dataPoint = point
	return nil
}

// AddNodes implements the Node interface.
func (n *OriginNode) AddNodes(nodes ...Node) error {
	if len(nodes) > 0 {
		return fmt.Errorf("origin node cannot have connected nodes")
	}
	return nil
}

// Nodes implements the Node interface.
func (n *OriginNode) Nodes() []Node {
	return nil
}

// Origin returns the origin name.
func (n *OriginNode) Origin() string {
	return n.origin
}

// Query returns the query used to fetch the data point.
func (n *OriginNode) Query() any {
	return n.query
}

// DataPoint implements the Node interface.
func (n *OriginNode) DataPoint() datapoint.Point {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.validate()
	return n.dataPoint
}

// IsFresh returns true if the price is considered fresh, that is, the price
// update is not required.
//
// Note, that the price that is not fresh is not necessarily expired.
func (n *OriginNode) IsFresh() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isFresh()
}

// IsExpired returns true if the price is considered expired.
func (n *OriginNode) IsExpired() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isExpired()
}

// Meta implements the Node interface.
func (n *OriginNode) Meta() map[string]any {
	return map[string]any{
		"type":                "origin",
		"origin":              n.origin,
		"query":               n.query,
		"freshness_threshold": n.freshnessThreshold,
		"expiry_threshold":    n.expiryThreshold,
	}
}

func (n *OriginNode) isFresh() bool {
	return n.dataPoint.Time.Add(n.freshnessThreshold).After(time.Now())
}

func (n *OriginNode) isExpired() bool {
	return n.dataPoint.Time.Add(n.expiryThreshold).Before(time.Now())
}

func (n *OriginNode) validate() bool {
	if n.dataPoint.Error == nil && n.isExpired() {
		n.dataPoint.Time = time.Now()
		n.dataPoint.Meta = n.Meta()
		n.dataPoint.Error = fmt.Errorf("data point is expired")
	}
	return n.dataPoint.Validate() == nil
}
