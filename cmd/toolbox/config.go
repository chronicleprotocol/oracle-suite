//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package main

import (
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	spectreConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/spectre"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

type Config struct {
	Ethereum ethereumConfig.Ethereum `json:"ethereum"`
	Spectre  spectreConfig.Spectre   `json:"spectre"`
}

func (c *Config) Configure() (ethereum.Client, ethereum.Signer, error) {
	sig, err := c.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, nil, err
	}
	cli, err := c.Ethereum.ConfigureEthereumClient(sig, null.New())
	if err != nil {
		return nil, nil, err
	}
	return cli, sig, nil
}

func (c *Config) Medianizers() map[string]spectreConfig.Medianizer {
	return c.Spectre.Medianizers
}

type Services struct {
	Client ethereum.Client
	Signer ethereum.Signer
}

func PrepareServices(opts *options) (*Services, error) {
	// Load config file:
	err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Services:
	cli, sig, err := opts.Config.Configure()
	if err != nil {
		return nil, fmt.Errorf("failed to load Spire configuration: %w", err)
	}

	return &Services{
		Client: cli,
		Signer: sig,
	}, nil
}
