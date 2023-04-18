package priceprovider

import (
	"net/http"

	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/graph"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"
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

	priceModels := map[string]graph.Node{}
	for _, pm := range c.PriceModels {
		priceModels[pm.Name], err = pm.ConfigurePriceModel()
		if err != nil {
			return nil, err
		}
	}

	return graph.NewProvider(priceModels, graph.NewUpdater(origins)), nil
}
