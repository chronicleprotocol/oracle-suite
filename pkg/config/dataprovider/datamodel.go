package dataprovider

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

const (
	defaultFreshnessThreshold = time.Minute
	defaultExpiryThreshold    = time.Minute * 5
)

type configDataModel struct {
	// Name of the data model.
	Name string `hcl:"name,label"`

	configNode
}

// configDynamicNode is an interface that is implemented by node types that
// can be used in a price model.
type configDynamicNode interface {
	buildGraph(origins map[string]origin.Origin, roots map[string]graph.Node) ([]graph.Node, error)
	hclRange() hcl.Range
}

type configNode struct {
	Nodes []configDynamicNode // Handled by PostDecodeBlock method.

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"`
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

// configNodeOrigin is a configuration for an Origin node.
type configNodeOrigin struct {
	Origin string `hcl:"origin,label"`

	configNode

	Query              cty.Value `hcl:"query"`
	FreshnessThreshold int       `hcl:"freshness_threshold,optional"`
	ExpiryThreshold    int       `hcl:"expiry_threshold,optional"`
}

// configNodeReference is a configuration for a Reference node.
type configNodeReference struct {
	configNode

	DataModel string `hcl:"data_model"`
}

// configNodeInvert is a configuration for an Invert node.
type configNodeInvert struct {
	configNode
}

// configNodeAlias is a configuration for an Alias node.
type configNodeAlias struct {
	Pair value.Pair `hcl:"pair,label"`

	configNode
}

// configNodeIndirect is a configuration for an Indirect node.
type configNodeIndirect struct {
	configNode
}

// configNodeMedian is a configuration for a Median node.
type configNodeMedian struct {
	configNode

	MinValues int `hcl:"min_values"`
}

// DeviationCircuitBreaker is a configuration for a DeviationCircuitBreaker node.
type DeviationCircuitBreaker struct {
	configNode

	Threshold float64 `hcl:"threshold"`
}

func (c *configDataModel) configureDataModel(
	origins map[string]origin.Origin,
	roots map[string]graph.Node,
) (graph.Node, error) {

	nodes, err := c.buildGraph(origins, roots)
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   "Data model must have exactly one root node",
			Subject:  c.Range.Ptr(),
		}
	}
	return nodes[0], nil
}

var nodeSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "origin", LabelNames: []string{"origin"}},
		{Type: "reference", LabelNames: []string{}},
		{Type: "invert", LabelNames: []string{}},
		{Type: "alias", LabelNames: []string{"pair"}},
		{Type: "indirect", LabelNames: []string{}},
		{Type: "median", LabelNames: []string{}},
		{Type: "deviation_circuit_breaker", LabelNames: []string{}},
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
		case "alias":
			node = &configNodeAlias{}
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

func (c configNode) OnEncodeBlock(block *utilHCL.Block) hcl.Diagnostics {
	for _, node := range c.Nodes {
		var blockType string
		switch nodeType := node.(type) {
		case *configNodeOrigin:
			blockType = "origin"
		case *configNodeReference:
			blockType = "reference"
		case *configNodeInvert:
			blockType = "invert"
		case *configNodeAlias:
			blockType = "alias"
		case *configNodeIndirect:
			blockType = "indirect"
		case *configNodeMedian:
			blockType = "median"
		case *DeviationCircuitBreaker:
			blockType = "deviation_circuit_breaker"
		default:
			return hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Encode error",
				Detail:   fmt.Sprintf("Unsupported node type: %T", nodeType),
			}}
		}
		err := utilHCL.EncodeBlock(node, block, blockType, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configNode) hclRange() hcl.Range {
	return c.Range
}

func (c *configNode) buildGraph(origins map[string]origin.Origin, roots map[string]graph.Node) ([]graph.Node, error) {
	nodes := make([]graph.Node, len(c.Nodes))
	for i, node := range c.Nodes {
		var err error
		nodes[i], err = buildNode(node, origins, roots)
		if err != nil {
			return nil, err
		}
		childNodes, err := node.buildGraph(origins, roots)
		if err != nil {
			return nil, err
		}
		if err := nodes[i].AddNodes(childNodes...); err != nil {
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

// buildNode returns a graph node based on the given configuration.
func buildNode(
	node configDynamicNode,
	origins map[string]origin.Origin,
	roots map[string]graph.Node,
) (graph.Node, error) {

	switch node := node.(type) {
	case *configNodeOrigin:
		return buildOriginNode(node, origins)
	case *configNodeReference:
		return buildReferenceNode(node, roots)
	case *configNodeInvert:
		return graph.NewTickInvertNode(), nil
	case *configNodeAlias:
		return graph.NewTickAliasNode(node.Pair), nil
	case *configNodeIndirect:
		return graph.NewTickIndirectNode(), nil
	case *configNodeMedian:
		return graph.NewTickMedianNode(node.MinValues), nil
	case *DeviationCircuitBreaker:
		return graph.NewDevCircuitBreakerNode(), nil
	default:
		return nil, fmt.Errorf("unsupported node type")
	}
}

// buildOriginNode returns an Origin node based on the given configuration.
func buildOriginNode(node *configNodeOrigin, origins map[string]origin.Origin) (graph.Node, error) {
	// Validate the threshold values.
	freshnessThreshold := time.Duration(node.FreshnessThreshold) * time.Second
	expiryThreshold := time.Duration(node.ExpiryThreshold) * time.Second
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

	// Check if the origin is known.
	if _, ok := origins[node.Origin]; !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Unknown origin: %s", node.Origin),
			Subject:  node.hclRange().Ptr(),
		}
	}

	// Parse the query value.
	var query any
	switch origins[node.Origin].(type) {
	case *origin.TickGenericJQ:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.BalancerV2:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.ComposableBalancerV2:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.WeightedBalancerV2:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.Curve:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.DSR:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.IShares:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.RocketPool:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.SDAI:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.Sushiswap:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.UniswapV2:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.UniswapV3:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.WrappedStakedETH:
		pair, err := value.PairFromString(node.Query.AsString())
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Invalid query: %s", err),
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = pair
	case *origin.Static:
		if node.Query.Type() != cty.Number {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   "Query must be a number",
				Subject:  node.hclRange().Ptr(),
			}
		}
		query = node.Query.AsBigFloat()
	}

	// Create the node.
	return graph.NewOriginNode(
		node.Origin,
		query,
		freshnessThreshold,
		expiryThreshold,
	), nil
}

// buildReferenceNode returns a Reference node based on the given configuration.
func buildReferenceNode(node *configNodeReference, roots map[string]graph.Node) (graph.Node, error) {
	model, ok := roots[node.DataModel]
	if !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Unknown data model: %s", node.DataModel),
			Subject:  node.hclRange().Ptr(),
		}
	}
	return model, nil
}
