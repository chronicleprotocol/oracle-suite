package priceprovider

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/graph"
	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

const (
	defaultFreshnessThreshold = time.Minute
	defaultExpiryThreshold    = time.Minute * 5
)

type configPriceModel struct {
	// Name of the price model.
	Name string `hcl:"name,label"`

	configNode
}

// configDynamicNode is an interface that is implemented by node types that
// can be used in a price model.
type configDynamicNode interface {
	buildGraph(roots map[string]graph.Node) ([]graph.Node, error)
	hclRange() hcl.Range
}

type configNode struct {
	// Pair is the pair of the source in the form of "BASE/QOUTE".
	Pair provider.Pair `hcl:"pair,label"`

	Nodes []configDynamicNode // Handled by PostDecodeBlock method.

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"`
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type configNodeOrigin struct {
	Origin string `hcl:"origin,label"`

	configNode

	FetchPair          provider.Pair `hcl:"fetch_pair,optional"`
	FreshnessThreshold int           `hcl:"freshness_threshold,optional"`
	ExpiryThreshold    int           `hcl:"expiry_threshold,optional"`
}

type configNodeReference struct {
	configNode

	PriceModel string `hcl:"price_model"`
}

type configNodeInvert struct {
	configNode
}

type configNodeIndirect struct {
	configNode
}

type configNodeMedian struct {
	configNode

	MinSources int `hcl:"min_sources"`
}

type DeviationCircuitBreaker struct {
	configNode

	Threshold float64 `hcl:"threshold"`
}

func (c *configPriceModel) ConfigurePriceModel(roots map[string]graph.Node) (graph.Node, error) {
	nodes, err := c.buildGraph(roots)
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Price model must have exactly one root node",
			Subject:  c.Range.Ptr(),
		}
	}
	return nodes[0], nil
}

var nodeSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "origin", LabelNames: []string{"origin", "pair"}},
		{Type: "reference", LabelNames: []string{"pair"}},
		{Type: "invert", LabelNames: []string{"pair"}},
		{Type: "indirect", LabelNames: []string{"pair"}},
		{Type: "median", LabelNames: []string{"pair"}},
		{Type: "deviation_circuit_breaker", LabelNames: []string{"pair"}},
	},
}

func (c *configNode) PostDecodeBlock(
	ctx *hcl.EvalContext,
	_ *hcl.BodySchema,
	_ *hcl.Block,
	_ *hcl.BodyContent) hcl.Diagnostics {

	content, diags := c.Remain.Content(nodeSchema)
	if diags.HasErrors() {
		return diags
	}
	var node configDynamicNode
	for _, block := range content.Blocks {
		switch block.Type {
		case "origin":
			node = &configNodeOrigin{}
		case "reference":
			node = &configNodeReference{}
		case "invert":
			node = &configNodeInvert{}
		case "indirect":
			node = &configNodeIndirect{}
		case "median":
			node = &configNodeMedian{}
		case "deviation_circuit_breaker":
			node = &DeviationCircuitBreaker{}
		}
		if diags := utilHCL.DecodeBlock(ctx, block, node); diags.HasErrors() {
			return diags
		}
		c.Nodes = append(c.Nodes, node)
	}
	return nil
}

func (c *configNode) hclRange() hcl.Range {
	return c.Range
}

func (c *configNode) buildGraph(roots map[string]graph.Node) ([]graph.Node, error) {
	nodes := make([]graph.Node, len(c.Nodes))
	for i, node := range c.Nodes {
		var err error
		nodes[i], err = buildNode(node, roots)
		if err != nil {
			return nil, err
		}
		branches, err := node.buildGraph(roots)
		if err != nil {
			return nil, err
		}
		if err := nodes[i].AddBranch(branches...); err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   err.Error(),
				Subject:  node.hclRange().Ptr(),
			}
		}
	}
	return nodes, nil
}

func buildNode(node configDynamicNode, roots map[string]graph.Node) (graph.Node, error) {
	switch node := node.(type) {
	case *configNodeOrigin:
		return buildOriginNode(node)
	case *configNodeReference:
		return buildReferenceNode(node, roots)
	case *configNodeInvert:
		return graph.NewInvertNode(node.Pair), nil
	case *configNodeIndirect:
		return graph.NewIndirectNode(node.Pair), nil
	case *configNodeMedian:
		return graph.NewMedianNode(node.Pair, node.MinSources), nil
	case *DeviationCircuitBreaker:
		return graph.NewDevCircuitBreakerNode(node.Pair, node.Threshold), nil
	default:
		return nil, fmt.Errorf("unsupported node type")
	}
}

func buildOriginNode(node *configNodeOrigin) (graph.Node, error) {
	freshnessThreshold := time.Duration(node.FreshnessThreshold)
	expiryThreshold := time.Duration(node.ExpiryThreshold)
	if freshnessThreshold == 0 {
		freshnessThreshold = defaultFreshnessThreshold
	}
	if expiryThreshold == 0 {
		expiryThreshold = defaultExpiryThreshold
	}
	if freshnessThreshold <= 0 {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Freshness threshold must be greater than zero",
			Subject:  node.hclRange().Ptr(),
		}
	}
	if expiryThreshold <= 0 {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Expiry threshold must be greater than zero",
			Subject:  node.hclRange().Ptr(),
		}
	}
	if freshnessThreshold > expiryThreshold {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Freshness threshold must be less than expiry threshold",
			Subject:  node.hclRange().Ptr(),
		}
	}
	fetchPair := node.FetchPair
	if node.FetchPair.Empty() {
		fetchPair = node.Pair
	}
	return graph.NewOriginNode(
		node.Origin,
		node.Pair,
		fetchPair,
		freshnessThreshold,
		expiryThreshold,
	), nil
}

func buildReferenceNode(node *configNodeReference, roots map[string]graph.Node) (graph.Node, error) {
	priceModel, ok := roots[node.PriceModel]
	if !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Unknown price model: %s", node.PriceModel),
			Subject:  node.hclRange().Ptr(),
		}
	}
	return priceModel, nil
}
