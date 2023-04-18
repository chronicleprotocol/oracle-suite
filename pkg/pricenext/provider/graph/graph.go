package graph

import "github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"

// Node represents a node in a graph.
//
// A node may be a leaf or a branch. A leaf node is a node that has no
// branches and provides a tick price from a data source. A branch node
// calculates a tick price from the ticks of its branches.
type Node interface {
	// Branches returns a list of nodes that are connected to the current node.
	Branches() []Node

	// AddBranch adds a new branch to the current node.
	AddBranch(...Node) error

	// Pair returns a tick pair for the current node.
	Pair() provider.Pair

	// Tick returns a tick for the current node. Leaf nodes return a tick
	// from a data source, while branch nodes return a tick calculated from
	// the ticks of its branches.
	Tick() provider.Tick
}

// Walk walks through the graph recursively and calls the given function
// for each node.
func Walk(fn func(Node), nodes ...Node) {
	r := map[Node]struct{}{}
	for _, node := range nodes {
		var recur func(Node)
		recur = func(node Node) {
			if _, ok := r[node]; ok {
				return
			}
			r[node] = struct{}{}
			for _, n := range node.Branches() {
				recur(n)
			}
		}
		recur(node)
	}
	for n := range r {
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

		// Check for cycles in each branch.
		for _, nextNode := range currentNode.Branches() {
			// Create a copy of the path for each branch.
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
