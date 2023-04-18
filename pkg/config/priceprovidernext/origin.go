package priceprovider

import (
	"fmt"
	"net/http"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider/origin"

	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

type OriginDependencies struct {
	HTTPClient *http.Client
	Clients    ethereum.ClientRegistry
}

type configOrigin struct {
	// Name of the origin.
	Name string `hcl:"name,label"`

	// Origin is the name of the origin provider.
	Origin string `hcl:"origin"`

	OriginConfig any // Handled by PostDecodeBlock method.

	// HCL fields:
	Content hcl.BodyContent `hcl:",content"`
	Remain  hcl.Body        `hcl:",remain"`
	Range   hcl.Range       `hcl:",range"`
}

type configOriginGenericJQ struct {
	URL string `hcl:"url"` // Do not use config.URL because it encode $ sign
	JQ  string `hcl:"jq"`
}

type configOriginGenericEVM struct {
	EthereumClient string                       `hcl:"ethereum_client"`
	Pairs          []configOriginGenericEVMPair `hcl:"pair,block"`

	// HCL fields:
	Content hcl.BodyContent `hcl:",content"`
}

type configOriginGenericEVMPair struct {
	Pair      provider.Pair `hcl:"pair,label"`
	Contract  types.Address `hcl:"contract"`
	ABI       string        `hcl:"abi"`
	Arguments []any         `hcl:"arguments"`
	Decimals  uint8         `hcl:"decimals"`

	// HCL fields:
	Content hcl.BodyContent `hcl:",content"`
}

func (c *configOrigin) PostDecodeBlock(
	ctx *hcl.EvalContext,
	_ *hcl.BodySchema,
	_ *hcl.Block,
	_ *hcl.BodyContent) hcl.Diagnostics {
	var config any
	switch c.Origin {
	case "generic_jq":
		config = &configOriginGenericJQ{}
	case "generic_evm":
		config = &configOriginGenericEVM{}
	default:
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Unknown origin: %s", c.Origin),
			Subject:  c.Range.Ptr(),
		}}
	}
	if diags := utilHCL.Decode(ctx, c.Remain, config); diags.HasErrors() {
		return diags
	}
	c.OriginConfig = config
	return nil
}

func (c *configOrigin) ConfigureOrigin(d OriginDependencies) (origin.Origin, error) {
	switch o := c.OriginConfig.(type) {
	case *configOriginGenericJQ:
		origin, err := origin.NewGenericJQ(d.HTTPClient, nil, o.URL, o.JQ)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginGenericEVM:
		pairs := make(map[provider.Pair]origin.GenericETHContract)
		client, ok := d.Clients[o.EthereumClient]
		if !ok {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", o.EthereumClient),
				Subject:  o.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}
		for _, p := range o.Pairs {
			method, err := abi.ParseMethod(p.ABI)
			if err != nil {
				return nil, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Validation error",
					Detail:   fmt.Sprintf("Failed to parse ABI: %s", err),
					Subject:  p.Content.Attributes["abi"].Range.Ptr(),
				}
			}
			pairs[p.Pair] = origin.GenericETHContract{
				Method:    method,
				Contract:  p.Contract,
				Arguments: p.Arguments,
				Decimals:  p.Decimals,
			}
		}
		origin, err := origin.NewGenericEVM(client, pairs)
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	}
	return nil, fmt.Errorf("unknown origin %s", c.Origin)
}
