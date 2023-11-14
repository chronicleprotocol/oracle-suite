//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package dataprovider

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/origin"
	utilHCL "github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"

	"github.com/hashicorp/hcl/v2"
)

type configOrigin struct {
	// Name of the origin.
	Name string `hcl:"name,label"`

	// Type is the type of the origin.
	Type string `hcl:"type"`

	// OriginConfig is the configuration of the origin.
	// Handled by PostDecodeBlock method.
	OriginConfig any

	// HCL fields:
	Content hcl.BodyContent `hcl:",content"`
	Remain  hcl.Body        `hcl:",remain"`
	Range   hcl.Range       `hcl:",range"`
}

// configOriginStatic is a configuration for the static origin.
type configOriginStatic struct{}

// configOriginTickGenericJQ is a configuration for the TickGenericJQ origin.
type configOriginTickGenericJQ struct {
	URL string `hcl:"url"` // Do not use config.URL because it encodes $ sign
	JQ  string `hcl:"jq"`
}

type configOriginIShares struct {
	URL string `hcl:"url"`
}

type configBalancerContracts struct {
	EthereumClient    string                   `hcl:"client,label"`
	ContractAddresses origin.ContractAddresses `hcl:"addresses"`
	// `references` are optional, the key should be matched with `addresses` as the additional address.
	ReferenceAddresses origin.ContractAddresses `hcl:"references,optional"`
}

type configOriginBalancer struct {
	// `addresses` are the pool addresses of WeightedPool2Tokens or MetaStablePool
	// `references` are used if the pool is MetaStablePool, key of mapping is the same key to `addresses` and
	// value should be the token0 address of pool.
	// If the pool is WeightedPool2Tokens, `references` should not contain the reference key of that pool.
	Contracts configBalancerContracts `hcl:"contracts,block"`
}

type configCurveContracts struct {
	EthereumClient string `hcl:"client,label"`
	// `addresses` are the pool addresses that are using `int256` (stableswap)
	StableSwapContractAddresses origin.ContractAddresses `hcl:"addresses"`
	// `addresses2` are the pool address that are using `uint256` (cryptoswap)
	CryptoSwapContractAddresses origin.ContractAddresses `hcl:"addresses2"`
}

type configOriginCurve struct {
	Contracts configCurveContracts `hcl:"contracts,block"`
}

type configContracts struct {
	EthereumClient    string                   `hcl:"client,label"`
	ContractAddresses origin.ContractAddresses `hcl:"addresses"`
}

type configOriginDSR struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginRocketPool struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginSDAI struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginSushiswap struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginUniswapV2 struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginUniswapV3 struct {
	Contracts configContracts `hcl:"contracts,block"`
}

type configOriginWrappedStakedETH struct {
	Contracts configContracts `hcl:"contracts,block"`
}

// averageFromBlocks is a list of blocks distances from the latest blocks from
// which prices will be averaged.
var averageFromBlocks = []int64{0, 10, 20}

func (c *configOrigin) PostDecodeBlock(
	ctx *hcl.EvalContext,
	_ *hcl.BodySchema,
	_ *hcl.Block,
	_ *hcl.BodyContent) hcl.Diagnostics {

	var config any
	switch c.Type {
	case "static":
		config = &configOriginStatic{}
	case "tick_generic_jq":
		config = &configOriginTickGenericJQ{}
	case "balancerV2":
		config = &configOriginBalancer{}
	case "curve":
		config = &configOriginCurve{}
	case "dsr":
		config = &configOriginDSR{}
	case "ishares":
		config = &configOriginIShares{}
	case "rocketpool":
		config = &configOriginRocketPool{}
	case "sdai":
		config = &configOriginSDAI{}
	case "sushiswap":
		config = &configOriginSushiswap{}
	case "uniswapV2":
		config = &configOriginUniswapV2{}
	case "uniswapV3":
		config = &configOriginUniswapV3{}
	case "wsteth":
		config = &configOriginWrappedStakedETH{}
	default:
		return hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Unknown origin: %s", c.Type),
			Subject:  c.Range.Ptr(),
		}}
	}
	if diags := utilHCL.Decode(ctx, c.Remain, config); diags.HasErrors() {
		return diags
	}
	c.OriginConfig = config
	return nil
}

func (c configOrigin) OnEncodeBlock(body *utilHCL.Block) hcl.Diagnostics {
	return utilHCL.Encode(c.OriginConfig, body)
}

func (c *configOrigin) configureOrigin(d Dependencies) (origin.Origin, error) {
	switch o := c.OriginConfig.(type) {
	case *configOriginStatic:
		return origin.NewStatic(), nil
	case *configOriginTickGenericJQ:
		origin, err := origin.NewTickGenericJQ(origin.TickGenericJQConfig{
			URL:     o.URL,
			Query:   o.JQ,
			Headers: nil,
			Client:  d.HTTPClient,
			Logger:  d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create jq origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginBalancer:
		origin, err := origin.NewBalancerV2(origin.BalancerV2Config{
			Client:             d.Clients[o.Contracts.EthereumClient],
			ContractAddresses:  o.Contracts.ContractAddresses,
			ReferenceAddresses: o.Contracts.ReferenceAddresses,
			Blocks:             averageFromBlocks,
			Logger:             d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create balancer origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginCurve:
		origin, err := origin.NewCurve(origin.CurveConfig{
			Client:                      d.Clients[o.Contracts.EthereumClient],
			StableSwapContractAddresses: o.Contracts.StableSwapContractAddresses,
			CryptoSwapContractAddresses: o.Contracts.CryptoSwapContractAddresses,
			Blocks:                      averageFromBlocks,
			Logger:                      d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create curve origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginDSR:
		origin, err := origin.NewDSR(origin.DSRConfig{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create dsr origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginIShares:
		origin, err := origin.NewIShares(origin.ISharesConfig{
			URL:     o.URL,
			Headers: nil,
			Client:  d.HTTPClient,
			Logger:  d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create ishares origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginRocketPool:
		origin, err := origin.NewRocketPool(origin.RocketPoolConfig{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create rocketpool origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginSDAI:
		origin, err := origin.NewSDAI(origin.SDAIConfig{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create sdai origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginSushiswap:
		origin, err := origin.NewSushiswap(origin.SushiswapConfig{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create sushiswap origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginUniswapV2:
		origin, err := origin.NewUniswapV2(origin.UniswapV2Config{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create uniswap v2 origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginUniswapV3:
		origin, err := origin.NewUniswapV3(origin.UniswapV3Config{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create uniswap v3 origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	case *configOriginWrappedStakedETH:
		origin, err := origin.NewWrappedStakedETH(origin.WrappedStakedETHConfig{
			Client:            d.Clients[o.Contracts.EthereumClient],
			ContractAddresses: o.Contracts.ContractAddresses,
			Blocks:            averageFromBlocks,
			Logger:            d.Logger,
		})
		if err != nil {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Runtime error",
				Detail:   fmt.Sprintf("Failed to create wsteth v3 origin: %s", err),
				Subject:  c.Range.Ptr(),
			}
		}
		return origin, nil
	}
	return nil, fmt.Errorf("unknown origin %s", c.Type)
}
