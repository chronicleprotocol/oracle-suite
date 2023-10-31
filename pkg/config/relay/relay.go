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

package relay

import (
	"fmt"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/signer"
	datapointStore "github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	musigStore "github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/relay"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Services struct {
	Relay      *relay.Relay
	PriceStore *datapointStore.Store
	MuSigStore *musigStore.Store
}

type Dependencies struct {
	Clients   ethereumConfig.ClientRegistry
	Transport transport.Service
	Logger    log.Logger
}

type Config struct {
	// Median is a list of Median contracts to watch.
	Median []configMedian `hcl:"median,block"`

	// Scribe is a list of Scribe contracts to watch.
	Scribe []configScribe `hcl:"scribe,block"`

	// OptimisticScribe is a list of OptimisticScribe contracts to watch.
	OptimisticScribe []configOptimisticScribe `hcl:"optimistic_scribe,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured services:
	services *Services
}

type configCommon struct {
	// EthereumClient is a name of an Ethereum client to use.
	EthereumClient string `hcl:"ethereum_client"`

	// ContractAddr is an address of a Median contract.
	ContractAddr types.Address `hcl:"contract_addr"`

	// Pairs is a list of pairs to store in the price store.
	Feeds []types.Address `hcl:"feeds"`

	// DataModel is a data model to use for the Median contract.
	DataModel string `hcl:"data_model"`

	// Spread is a minimum spread between the current price to trigger an
	// update. A spread is represented as a percentage point, e.g. 1 means
	// 1%.
	Spread float64 `hcl:"spread"`

	// Expiration is a time in seconds after which the price is considered
	// expired which triggers an update.
	Expiration uint32 `hcl:"expiration"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type configMedian struct {
	configCommon

	// Pairs is a list of pairs to store in the price store.
	Feeds []types.Address `hcl:"feeds"`
}

type configScribe struct {
	configCommon
}

type configOptimisticScribe struct {
	configCommon

	// OptimisticSpread is a minimum spread between the current price to
	// trigger an optimistic update. A spread is represented as a percentage
	// point, e.g. 1 means 1%.
	OptimisticSpread float64 `hcl:"optimistic_spread"`

	// OptimisticExpiration is a time in seconds after which the price is
	// considered expired which triggers an optimistic update.
	OptimisticExpiration uint32 `hcl:"optimistic_expiration"`
}

func configCommonFields(c configCommon) log.Fields {
	return log.Fields{
		"ethereumClient": c.EthereumClient,
		"contractAddr":   c.ContractAddr,
		"dataModel":      c.DataModel,
		"spread":         c.Spread,
		"expiration":     c.Expiration,
	}
}

const LoggerTag = "CONFIG_" + relay.LoggerTag

func (c *Config) Relay(d Dependencies) (*Services, error) {
	logger := d.Logger.
		WithField("tag", LoggerTag)

	if c.services != nil {
		return c.services, nil
	}

	// Find data models required by all median contracts.
	var (
		medianDataModels []string
		scribeDataModels []string
	)
	for _, cfg := range c.Median {
		medianDataModels = append(medianDataModels, cfg.DataModel)
	}
	for _, cfg := range c.Scribe {
		scribeDataModels = append(scribeDataModels, cfg.DataModel)
	}
	for _, cfg := range c.OptimisticScribe {
		scribeDataModels = append(scribeDataModels, cfg.DataModel)
	}
	dataModels := append(append(medianDataModels, scribeDataModels...), scribeDataModels...)

	logger.
		WithFields(log.Fields{
			"dataModels": dataModels,
		}).
		Debug("Data models")

	// Create a data point store service for all median contracts.
	priceStoreSrv, err := datapointStore.New(datapointStore.Config{
		Storage:    datapointStore.NewMemoryStorage(),
		Transport:  d.Transport,
		Models:     dataModels,
		Recoverers: []datapoint.Recoverer{signer.NewTickRecoverer(crypto.ECRecoverer)},
		Logger:     d.Logger,
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Store error",
			Detail:   fmt.Sprintf("Failed to create the data point store service: %v", err),
			Subject:  &c.Range,
		}
	}

	// Create Store service.
	musigStoreSrv := musigStore.New(musigStore.Config{
		Transport:  d.Transport,
		DataModels: scribeDataModels,
		Logger:     d.Logger,
	})

	var (
		medianCfgs   []relay.ConfigMedian
		scribeCfgs   []relay.ConfigScribe
		opScribeCfgs []relay.ConfigOptimisticScribe
	)

	for _, cfg := range c.Median {
		client, ok := d.Clients[cfg.EthereumClient]
		if !ok {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", cfg.EthereumClient),
				Subject:  cfg.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}

		logger.
			WithField("contract", "Median").
			WithFields(configCommonFields(cfg.configCommon)).
			Info("Contract")

		medianCfgs = append(medianCfgs, relay.ConfigMedian{
			DataModel:       cfg.DataModel,
			ContractAddress: cfg.ContractAddr,
			FeedAddresses:   cfg.Feeds,
			Client:          client,
			DataPointStore:  priceStoreSrv,
			Spread:          cfg.Spread,
			Expiration:      time.Second * time.Duration(cfg.Expiration),
		})
	}
	for _, cfg := range c.Scribe {
		client, ok := d.Clients[cfg.EthereumClient]
		if !ok {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", cfg.EthereumClient),
				Subject:  cfg.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}

		logger.
			WithField("contract", "Scribe").
			WithFields(configCommonFields(cfg.configCommon)).
			Info("Contract")

		scribeCfgs = append(scribeCfgs, relay.ConfigScribe{
			DataModel:       cfg.DataModel,
			ContractAddress: cfg.ContractAddr,
			Client:          client,
			MuSigStore:      musigStoreSrv,
			Spread:          cfg.Spread,
			Expiration:      time.Second * time.Duration(cfg.Expiration),
		})
	}
	for _, cfg := range c.OptimisticScribe {
		client, ok := d.Clients[cfg.EthereumClient]
		if !ok {
			return nil, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation error",
				Detail:   fmt.Sprintf("Ethereum client %q is not configured", cfg.EthereumClient),
				Subject:  cfg.Content.Attributes["ethereum_client"].Range.Ptr(),
			}
		}

		logger.
			WithField("contract", "OptimisticScribe").
			WithFields(configCommonFields(cfg.configCommon)).
			Info("Contract")

		opScribeCfgs = append(opScribeCfgs, relay.ConfigOptimisticScribe{
			DataModel:            cfg.DataModel,
			ContractAddress:      cfg.ContractAddr,
			Client:               client,
			MuSigStore:           musigStoreSrv,
			Spread:               cfg.Spread,
			Expiration:           time.Second * time.Duration(cfg.Expiration),
			OptimisticSpread:     cfg.OptimisticSpread,
			OptimisticExpiration: time.Second * time.Duration(cfg.OptimisticExpiration),
		})
	}

	relaySrv, err := relay.New(relay.Config{
		Medians:           medianCfgs,
		Scribes:           scribeCfgs,
		OptimisticScribes: opScribeCfgs,
		Logger:            d.Logger,
		Ticker:            timeutil.NewTicker(time.Minute * 2),
	})
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Relay error",
			Detail:   fmt.Sprintf("Failed to create the relay service: %v", err),
			Subject:  &c.Range,
		}
	}

	c.services = &Services{
		Relay:      relaySrv,
		PriceStore: priceStoreSrv,
		MuSigStore: musigStoreSrv,
	}
	return c.services, nil
}
