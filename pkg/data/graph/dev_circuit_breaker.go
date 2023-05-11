package graph

import (
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// DevCircuitBreakerNode is a circuit breaker that tips if the value deviation
// between two nodes is greater than the breaker value.
//
// It must have three nodes. First node is the value, second is the
// reference value and third is the threshold value.
//
// The value deviation is calculated as:
// abs(1.0 - (reference / value))
//
// All nodes must return a value that implements the data.NumericValue
// interface.
type DevCircuitBreakerNode struct {
	valueNode     Node
	referenceNode Node
	thresholdNode Node
}

// NewDevCircuitBreakerNode creates a new DevCircuitBreakerNode instance.
func NewDevCircuitBreakerNode() *DevCircuitBreakerNode {
	return &DevCircuitBreakerNode{}
}

// Nodes implements the Node interface.
func (n *DevCircuitBreakerNode) Nodes() []Node {
	if n.valueNode == nil || n.referenceNode == nil || n.thresholdNode == nil {
		return nil
	}
	return []Node{n.valueNode, n.referenceNode, n.thresholdNode}
}

// AddNodes implements the Node interface.
//
// If more than three nodes are added, an error is returned.
func (n *DevCircuitBreakerNode) AddNodes(nodes ...Node) error {
	if len(nodes) > 0 && n.valueNode == nil {
		n.valueNode = NewWrapperNode(nodes[0], map[string]any{"type": "value"})
		nodes = nodes[1:]
	}
	if len(nodes) > 0 && n.referenceNode == nil {
		n.referenceNode = NewWrapperNode(nodes[0], map[string]any{"type": "reference_value"})
		nodes = nodes[1:]
	}
	if len(nodes) > 0 && n.thresholdNode == nil {
		n.thresholdNode = NewWrapperNode(nodes[0], map[string]any{"type": "threshold"})
		nodes = nodes[1:]
	}
	if len(nodes) > 0 {
		return fmt.Errorf("only three nodes are allowed")
	}
	return nil
}

// DataPoint implements the Node interface.
func (n *DevCircuitBreakerNode) DataPoint() data.Point {
	// Validate nodes.
	if n.valueNode == nil || n.referenceNode == nil || n.thresholdNode == nil {
		return data.Point{
			Error: fmt.Errorf("three nodes are required: value, reference value and threshold"),
			Time:  time.Now(),
		}
	}
	valuePoint := n.valueNode.DataPoint()
	refPoint := n.referenceNode.DataPoint()
	thresholdPoint := n.thresholdNode.DataPoint()
	if err := valuePoint.Validate(); err != nil {
		return data.Point{
			Error: fmt.Errorf("invalid value data point: %w", err),
			Time:  time.Now(),
		}
	}
	if err := refPoint.Validate(); err != nil {
		return data.Point{
			Error: fmt.Errorf("invalid reference data point: %w", err),
			Time:  time.Now(),
		}
	}
	if err := thresholdPoint.Validate(); err != nil {
		return data.Point{
			Error: fmt.Errorf("invalid threshold data point: %w", err),
			Time:  time.Now(),
		}
	}
	valueValue, ok := valuePoint.Value.(data.NumericValue)
	if !ok {
		return data.Point{
			Error: fmt.Errorf("invalid value data point, expected numeric value"),
			Time:  time.Now(),
		}
	}
	refValue, ok := refPoint.Value.(data.NumericValue)
	if !ok {
		return data.Point{
			Error: fmt.Errorf("invalid reference data point, expected numeric value"),
			Time:  time.Now(),
		}
	}
	thresholdValue, ok := thresholdPoint.Value.(data.NumericValue)
	if !ok {
		return data.Point{
			Error: fmt.Errorf("invalid threshold data point, expected numeric value"),
			Time:  time.Now(),
		}
	}

	meta := n.Meta()

	// Calculate deviation.
	deviation := bn.Float(1.0).Sub(refValue.Number().Div(valueValue.Number())).Abs().Float64()
	meta["deviation"] = deviation
	meta["threshold"] = thresholdValue.Number().Float64()

	// Return tick, if deviation is greater than threshold, add error.
	point := valuePoint
	point.SubPoints = []data.Point{n.valueNode.DataPoint(), n.referenceNode.DataPoint()}
	point.Meta = meta
	if deviation > thresholdValue.Number().Float64() {
		point.Error = fmt.Errorf("deviation %f is greater than threshold %s", deviation, thresholdValue.Number())
	}
	return point
}

// Meta implements the Node interface.
func (n *DevCircuitBreakerNode) Meta() map[string]any {
	return map[string]any{
		"type": "deviation_circuit_breaker",
	}
}
