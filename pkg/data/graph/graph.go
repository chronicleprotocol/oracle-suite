package graph

import "github.com/chronicleprotocol/oracle-suite/pkg/data"

// Node represents a node in a graph.
//
// A node can use data from connected nodes to calculate its own data or
// it can return data from a data source.
type Node interface {
	// Nodes returns a list of nodes that are connected to the current node.
	Nodes() []Node

	// AddNodes adds nodes to the current node.
	AddNodes(...Node) error

	// DataPoint returns a data point for the current node. Leaf nodes return
	// a data point from a data source, while other nodes return a data point
	// calculated from the data points of connected nodes.
	DataPoint() data.Point

	// Meta returns a map that contains meta information about the node used
	// to debug price models. It should not contain data that is accessible
	// from node's methods.
	Meta() map[string]any
}

// MapMeta is a map that contains meta information as a key-value pairs.
type MapMeta map[string]any

// Map implements the data.Meta interface.
func (m MapMeta) Map() map[string]any {
	return m
}

// Walk walks through the graph recursively and calls the given function
// for each node.
func Walk(fn func(Node), nodes ...Node) {
	visited := map[Node]struct{}{}

	for _, node := range nodes {
		var walkNodes func(Node)
		walkNodes = func(node Node) {
			// Skip already visited nodes.
			if _, ok := visited[node]; ok {
				return
			}

			// Mark the node as visited.
			visited[node] = struct{}{}

			// Recursively walk through the graph.
			for _, n := range node.Nodes() {
				walkNodes(n)
			}
		}
		walkNodes(node)
	}

	// Call the given callback function for each node.
	for n := range visited {
		fn(n)
	}
}

// DetectCycle returns a cycle path in the given graph if a cycle is detected,
// otherwise returns an empty slice.
func DetectCycle(node Node) []Node {
	visited := map[Node]struct{}{}

	// checkCycle recursively checks for cycles in the graph.
	var checkCycle func(Node, []Node) []Node
	checkCycle = func(currentNode Node, path []Node) []Node {
		// If currentNode is already in the path, a cycle is detected.
		for _, parent := range path {
			if parent == currentNode {
				return path
			}
		}

		// Skip checking already visited nodes.
		if _, ok := visited[currentNode]; ok {
			return nil
		}
		visited[currentNode] = struct{}{}

		// Add the current node to the path.
		path = append(path, currentNode)

		// Check for cycles in each node connected to the current node.
		for _, nextNode := range currentNode.Nodes() {
			// Create a copy of the path for each node.
			pathCopy := make([]Node, len(path))
			copy(pathCopy, path)

			// If a cycle is detected, return the path.
			if cyclePath := checkCycle(nextNode, pathCopy); cyclePath != nil {
				return cyclePath
			}
		}

		return nil
	}

	return checkCycle(node, nil)
}
