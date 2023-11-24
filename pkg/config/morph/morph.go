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
	"os"
	"strings"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	pkgSupervisor "github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
	pkgWatchDog "github.com/chronicleprotocol/oracle-suite/pkg/watchdog"
)

const defaultInterval = 60 * 60

// Morph provides the solution to keep updated configurations for underlying applications.
//
// Gofer contains the data models we support.
// Every time someone wants new oracle that we've not done before, we need to update data models in gofer and
// deploy to the new feed release.
// As the feed upgrade is painful, there's a big latency in updating feed clients, ensuring everything is correct,
// getting feeds client updated, waiting their feedbacks and achieving the quorum on chain.
// This is a very slow model to push data models.
// We want to push data models on chain so as the feeds react automatically.
// This will help us scale number of clients quickly and grow faster than competitors.
//
// We are going to put updated configuration to the IPFS and store its hash on `ConfigRegistry` smart contract
// ensuring the decentralization and security, and having compatibility to update another config values
// not limited to data models. We call it on-chain configuration.
// Morph keep monitoring of on-chain configuration and let underlying apps such as ghost, spire, spectre, use it.
// In order to implement it, Morph hires WatchDog service to perform it.

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

	// ExecutableBinary is a path to executable binary that watchdog service initiates the app running.
	ExecutableBinary string `hcl:"bin"`

	// Concatenated string of arguments, with a format of `run --x1 y1 --x2 y2 --x3 y3`
	// It includes command and additional arguments of key and value.
	Args string `hcl:"args"`

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
	watchDogService, err := c.Morph.Configure(logger, clients)
	if err != nil {
		return nil, err
	}

	return &Services{
		WatchDog: watchDogService,
		Logger:   logger,
	}, nil
}

// Services returns the services that are configured from the Config struct.
type Services struct {
	WatchDog *pkgWatchDog.WatchDog
	Logger   log.Logger

	supervisor *pkgSupervisor.Supervisor
}

// Start implements the supervisor.Service interface.
func (s *Services) Start(ctx context.Context) error {
	if s.supervisor != nil {
		return fmt.Errorf("services already started")
	}
	s.supervisor = pkgSupervisor.New(s.Logger)
	s.supervisor.Watch(s.WatchDog)
	if l, ok := s.Logger.(pkgSupervisor.Service); ok {
		s.supervisor.Watch(l)
	}
	return s.supervisor.Start(ctx)
}

// Wait implements the supervisor.Service interface.
func (s *Services) Wait() <-chan error {
	return s.supervisor.Wait()
}

func (c *ConfigMorph) Configure(baseLogger log.Logger, clients ethereumConfig.ClientRegistry) (*pkgWatchDog.WatchDog, error) {
	interval := c.Interval
	if interval == 0 {
		interval = defaultInterval
	}

	var args []string
	if len(c.AppConfig.Args) > 0 {
		args = strings.Split(c.AppConfig.Args, " ")
	}
	if len(args) < 1 {
		// args should include command at least, set run by default
		return nil, fmt.Errorf("args should include command")
	}

	var workDir = c.AppConfig.WorkDir
	if workDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workDir = currentDir
	}

	cfg := pkgWatchDog.Config{
		LocalFile:                 c.MorphFile,
		Client:                    clients[c.EthereumClient],
		ConfigRegistryAddress:     c.ConfigRegistryAddress,
		Interval:                  timeutil.NewTicker(time.Second * time.Duration(interval)),
		WorkDir:                   workDir,
		ExecutableBinary:          c.AppConfig.ExecutableBinary,
		Args:                      args,
		WaitDurationForAppQuiting: time.Duration(c.AppConfig.WaitDurationForAppQuiting) * time.Second,
		Logger:                    baseLogger,
	}
	watchDog, err := pkgWatchDog.NewWatchDogService(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the WatchDog service: %v", err),
		}
	}
	return watchDog, nil
}
