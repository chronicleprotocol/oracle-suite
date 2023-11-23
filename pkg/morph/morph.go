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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

// Morph service is working as app initializer, is responsible for
// - keeping the local config up-to-dated
// - running the main application with the latest up-to-dated config
// - quiting the main application when quit itself
type Morph struct {
	ctx    context.Context
	waitCh chan error
	log    log.Logger

	// File path to the local cache config file, will be replaced with downloaded on-chain config
	morphFile string

	// Contract to ConfigRegistry smart contract, where config's IPFS url stores
	configRegistry *chronicle.ConfigRegistry

	// Interval in seconds to check if on-chain config has been updated
	interval *timeutil.Ticker

	// AppManager instance to restart the main application
	am *supervisor.AppManager

	// Temporary variable to cache last ipfs url downloaded
	lastIPFS string
}

type Config struct {
	MorphFile                 string
	Client                    rpc.RPC
	ConfigRegistryAddress     types.Address
	Interval                  *timeutil.Ticker
	WorkDir                   string
	ExecutableBinary          string
	Args                      []string
	WaitDurationForAppQuiting time.Duration
	Logger                    log.Logger
}

const LoggerTag = "MORPH"

// NewMorphService creates Morph, which proceeds the following steps:
// - Periodically pull the config from on-chain.
// - Compares with previous one, if found difference, restart app.
func NewMorphService(cfg Config) (*Morph, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}

	configRegistry := chronicle.NewConfigRegistry(cfg.Client, cfg.ConfigRegistryAddress)

	am, err := supervisor.NewAppManager(supervisor.AppManagerConfig{
		Envs:    []string{},
		WorkDir: cfg.WorkDir,
		Bin:     cfg.ExecutableBinary,
		Arguments: append([]string{
			"--config", cfg.MorphFile,
		}, cfg.Args...),
		WaitDurationForQuiting: cfg.WaitDurationForAppQuiting,
		Logger:                 cfg.Logger,
	})
	if err != nil {
		return nil, err
	}

	m := &Morph{
		waitCh:         make(chan error),
		log:            cfg.Logger.WithField("tag", LoggerTag),
		morphFile:      cfg.MorphFile,
		configRegistry: configRegistry,
		interval:       cfg.Interval,
		am:             am,
	}
	return m, nil
}

// Start starts running main application and monitoring the changes of on-chain config.
//   - download the latest on-chain config first and update local config file if newly updated
//   - execute command to run main application with local config file
//     ./ghost run --config config-cache.hcl
//     After executing command, morph will wait for app to run for a while (CFG_RUN_APP_DURATION in seconds)
func (m *Morph) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.ctx = ctx
	m.log.
		WithFields(log.Fields{
			"interval": m.interval.Duration(),
		}).
		Info("Starting")

	// Download the latest on-chain config if it has been updated
	if _, err := m.Monitor(); err != nil {
		m.log.
			WithError(err).
			Error("Monitoring latest config was failed")
		return err
	}
	// Make sure that the main application is running at the beginning
	if err := m.am.Start(ctx); err != nil {
		m.log.
			WithError(err).
			Error("Running app with latest config was failed")
		return err
	}

	go m.reloadRoutine()
	go m.contextCancelHandler()
	return nil
}

func (m *Morph) Wait() <-chan error {
	return m.waitCh
}

func (m *Morph) Monitor() (bool, error) {
	m.log.Debug("Fetching latest on-chain config")

	// Fetch latest IPFS from ConfigRegistry contract
	latest, err := m.configRegistry.Latest().Call(m.ctx, types.LatestBlockNumber)
	if err != nil {
		m.log.WithError(err).Error("Failed fetching latest ipfs for on-chain config")
		return false, err
	}

	m.log.WithField("ipfs", latest).Info("Found latest on-chain config")

	if latest == m.lastIPFS { // Do not fetch the ipfs content if nothing changed
		m.log.Info("There are no updated configuration")
		return false, nil
	}

	// Pull down the content from IPFS
	req, err := http.NewRequestWithContext(m.ctx, "GET", latest, nil)
	if err != nil {
		m.log.WithError(err).Error("Failed creating http request for ipfs")
		return false, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		m.log.WithError(err).Error("Failed fetching ipfs content")
		return false, err
	}
	onChainConfig, err := io.ReadAll(res.Body)
	if err != nil {
		m.log.WithError(err).Error("Failed reading http for fetching ipfs content")
		return false, err
	}
	res.Body.Close()

	m.lastIPFS = latest

	// Read morphFile
	cacheConfig, _ := os.ReadFile(m.morphFile)

	if strings.Compare(string(onChainConfig), string(cacheConfig)) == 0 {
		m.log.Info("There are no updated configuration")
		return false, nil
	}

	file, err := os.Create(m.morphFile)
	if err != nil {
		m.log.WithError(err).Error("Failed creating in-memory config to the file:" + m.morphFile)
		return false, err
	}
	defer file.Close()
	_, err = file.Write(onChainConfig)
	if err != nil {
		m.log.WithError(err).Error("Failed writing in-memory config to the file:" + m.morphFile)
		return false, err
	}

	m.log.Info("Found that on-chain configuration has been updated")

	return true, nil
}

func (m *Morph) RestartApp() error {
	// Quit app
	if err := m.am.QuitApp(); err != nil {
		return err
	}

	// Run the app
	return m.am.RunApp()
}

func (m *Morph) reloadRoutine() {
	m.interval.Start(m.ctx)
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.interval.TickCh():
			if !m.am.IsAppRunning() { // morph checks if app is running per every interval
				m.waitCh <- fmt.Errorf("service app was exited already for some reason")
				return
			}
			updated, err := m.Monitor()
			if err != nil {
				m.waitCh <- err
				return
			}
			if updated {
				if err = m.RestartApp(); err != nil {
					m.waitCh <- err
					return
				}
			}
		}
	}
}

func (m *Morph) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
	// When context has been cancelled, send interrupt signal to app to quit and wait for app to quit.
	if err := m.am.QuitApp(); err != nil {
		m.waitCh <- err
		return
	}
}
