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

package morph

import (
	"fmt"
	"os"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	morphService "github.com/chronicleprotocol/oracle-suite/pkg/morph"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const defaultInterval = 60 * 60

type Config struct {
	Ethereum ethereumConfig.Config `hcl:"ethereum,block"`
	Morph    morphConfig           `hcl:"morph,block"`
	Logger   *loggerConfig.Config  `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (Config) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Ethereum,
		config.Contracts,
		config.Morph,
	}
}

func (Config) DefaultPaths() []string {
	if cache := os.Getenv("CFG_CONFIG_CACHE"); cache != "" {
		return []string{cache, "config/config-morph.hcl"}
	}
	return []string{config.ConfigCacheFile, "config/config-morph.hcl"}
}

type morphConfig struct {
	// MorphFile is a file path to cache the latest config
	MorphFile string `hcl:"cache_path"`

	// EthereumClient is a name of an Ethereum client to use
	EthereumClient string `hcl:"ethereum_client"`

	// ConfigRegistryAddress is an address of ConfigRegistry contract.
	ConfigRegistryAddress types.Address `hcl:"config_registry"`

	// Interval is an interval of pulling on-chain config in seconds
	Interval uint32 `hcl:"interval"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

func (c *Config) Configure(base pkgSupervisor.Config, baseLogger log.Logger, appName string, appVersion string) (*morphService.Morph, error) {
	logger, err := c.Logger.Logger(loggerConfig.Dependencies{
		AppName:    appName,
		AppVersion: appVersion,
		BaseLogger: baseLogger,
	})
	if err != nil {
		return nil, err
	}

	clients, err := c.Ethereum.ClientRegistry(ethereumConfig.Dependencies{Logger: logger})
	if err != nil {
		return nil, err
	}

	interval := c.Morph.Interval
	if interval == 0 {
		interval = defaultInterval
	}

	cfg := morphService.Config{
		MorphFile:             c.Morph.MorphFile,
		Client:                clients[c.Morph.EthereumClient],
		ConfigRegistryAddress: c.Morph.ConfigRegistryAddress,
		Interval:              timeutil.NewTicker(time.Second * time.Duration(interval)),
		Base:                  base,
		Logger:                logger,
	}
	morph, err := morphService.NewMorphService(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Morph service: %v", err),
		}
	}
	return morph, nil
}
