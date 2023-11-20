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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	pkgMorph "github.com/chronicleprotocol/oracle-suite/pkg/morph"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const defaultInterval = 60 * 60

type Config struct {
	Morph    ConfigMorph           `hcl:"morph,block"`
	Ethereum ethereumConfig.Config `hcl:"ethereum,block"`
	Logger   *loggerConfig.Config  `hcl:"logger,block,optional"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (Config) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Ethereum,
		config.Morph,
	}
}

type ConfigMorph struct {
	// MorphFile is a file path to cache the latest config
	MorphFile string `hcl:"cache_path"`

	// EthereumClient is a name of an Ethereum client to use
	EthereumClient string `hcl:"ethereum_client"`

	// ConfigRegistryAddress is an address of ConfigRegistry contract.
	ConfigRegistryAddress types.Address `hcl:"config_registry"`

	// Interval is an interval of pulling on-chain config in seconds
	Interval uint32 `hcl:"interval"`

	// Config for running application
	AppConfig configApp `hcl:"app,block"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`
}

type configApp struct {
	// WorkDir is a working directory where run application
	WorkDir string `hcl:"work_dir,optional"`

	// ExecutableBinary is a path to executable binary that morph service initiates the app running.
	ExecutableBinary string `hcl:"bin"`

	// Main arguments that indicates the use case of application
	Use string `hcl:"use"`

	// Concatenated string of arguments, with a format of `--x1 y1 --x2 y2 --x3 y3`
	Args string `hcl:"args,optional"`

	// Time duration waiting for app quiting in second
	WaitDurationForAppQuiting uint32 `hcl:"waiting_quiting"`
}

func (c *Config) Services(baseLogger log.Logger, appName string, appVersion string) (pkgSupervisor.Service, error) {
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
	morphService, err := c.Morph.Configure(logger, clients)
	if err != nil {
		return nil, err
	}

	return &Services{
		Morph:  morphService,
		Logger: logger,
	}, nil
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	Morph  *pkgMorph.Morph
	Logger log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.Morph)
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

func (c *ConfigMorph) Configure(baseLogger log.Logger, clients ethereumConfig.ClientRegistry) (*pkgMorph.Morph, error) {
	interval := c.Interval
	if interval == 0 {
		interval = defaultInterval
	}

	var args []string
	if len(c.AppConfig.Args) > 0 {
		args = strings.Split(c.AppConfig.Args, " ")
	}

	cfg := pkgMorph.Config{
		MorphFile:                 c.MorphFile,
		Client:                    clients[c.EthereumClient],
		ConfigRegistryAddress:     c.ConfigRegistryAddress,
		Interval:                  timeutil.NewTicker(time.Second * time.Duration(interval)),
		WorkDir:                   c.AppConfig.WorkDir,
		ExecutableBinary:          c.AppConfig.ExecutableBinary,
		Use:                       c.AppConfig.Use,
		Args:                      args,
		WaitDurationForAppQuiting: time.Duration(c.AppConfig.WaitDurationForAppQuiting) * time.Second,
		Logger:                    baseLogger,
	}
	morph, err := pkgMorph.NewMorphService(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the Morph service: %v", err),
		}
	}
	return morph, nil
}
