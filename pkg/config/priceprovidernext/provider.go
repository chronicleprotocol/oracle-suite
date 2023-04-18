package priceprovider

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
)

type Dependencies struct {
	HTTPClient *http.Client
	Clients    ethereum.ClientRegistry
}

type Config struct {
	Origins     []configOrigin     `hcl:"origin,block"`
	PriceModels []configPriceModel `hcl:"price_model,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) PriceProvider(d Dependencies) (provider.Provider, error) {
	var err error

	// Configure origins.
	origins := map[string]origin.Origin{}
	for _, o := range c.Origins {
		origins[o.Name], err = o.ConfigureOrigin(OriginDependencies{
			HTTPClient: d.HTTPClient,
			Clients:    d.Clients,
		})
		if err != nil {
			return nil, err
		}
	}

	// Configure price models.
	priceModels := map[string]graph.Node{}
	for _, pm := range c.PriceModels {
		priceModels[pm.Name] = graph.NewReferenceNode(pm.Pair)
	}
	for _, pm := range c.PriceModels {
		priceModel, err := pm.ConfigurePriceModel(priceModels)
		if err != nil {
			return nil, err
		}
		if err := priceModels[pm.Name].AddBranch(priceModel); err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to add branch to price model %s %s: %s", pm.Pair, pm.Name, err),
				Subject:  pm.Range.Ptr(),
			}
		}
		if nodes := graph.DetectCycle(priceModels[pm.Name]); len(nodes) > 0 {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail: fmt.Sprintf(
					"Cycle detected in price model %s %s: %s",
					pm.Pair,
					pm.Name,
					strings.Join(sliceutil.Map(nodes, func(n graph.Node) string { return n.Pair().String() }), " -> "),
				),
				Subject: pm.Range.Ptr(),
			}
		}
	}

	return graph.NewProvider(priceModels, graph.NewUpdater(origins)), nil
}
