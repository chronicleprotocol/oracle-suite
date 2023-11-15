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

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Morph struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	waitCh    chan error

	morphFile      string
	configRegistry *chronicle.ConfigRegistry
	interval       *timeutil.Ticker
	base           config.HasDefaults
	log            log.Logger

	lastIPFS string
}

type Config struct {
	MorphFile             string
	Interval              *timeutil.Ticker
	Client                rpc.RPC
	ConfigRegistryAddress types.Address
	Base                  config.HasDefaults
	Logger                log.Logger
}

const LoggerTag = "MORPH"

// NewMorphService creates Morph, which proceeds the following steps:
// - Periodically pull the config from on-chain.
// - Compares with previous one, if found difference, exit app.
func NewMorphService(cfg Config) (*Morph, error) {
	configRegistry := chronicle.NewConfigRegistry(cfg.Client, cfg.ConfigRegistryAddress)

	m := &Morph{
		waitCh:         make(chan error),
		log:            cfg.Logger.WithField("tag", LoggerTag),
		morphFile:      cfg.MorphFile,
		configRegistry: configRegistry,
		interval:       cfg.Interval,
		base:           cfg.Base,
	}
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	return m, nil
}

func (m *Morph) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.ctx, m.ctxCancel = context.WithCancel(ctx)
	m.log.
		WithFields(log.Fields{
			"interval": m.interval.Duration(),
		}).
		Info("Starting")
	m.interval.Start(m.ctx)
	go m.reloadRoutine()
	go m.contextCancelHandler()
	return nil
}

func (m *Morph) Wait() <-chan error {
	return m.waitCh
}

func (m *Morph) Monitor() error {
	m.log.Debug("Fetching latest on-chain config")

	// Fetch latest IPFS from ConfigRegistry contract
	latest, err := m.configRegistry.Latest().Call(m.ctx, types.LatestBlockNumber)
	if err != nil {
		m.log.WithError(err).Error("Failed fetching latest ipfs for on-chain config")
		return err
	}

	m.log.WithField("ipfs", latest).Info("Found latest on-chain config")

	if latest == m.lastIPFS { // Do not fetch the ipfs content if nothing changed
		m.log.Info("There are no updated configuration")
		return nil
	}

	// Pull down the content from IPFS
	req, err := http.NewRequestWithContext(m.ctx, "GET", latest, nil)
	if err != nil {
		m.log.WithError(err).Error("Failed creating http request for ipfs")
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		m.log.WithError(err).Error("Failed fetching ipfs content")
		return err
	}
	onChainConfig, err := io.ReadAll(res.Body)
	if err != nil {
		m.log.WithError(err).Error("Failed reading http for fetching ipfs content")
		return err
	}
	res.Body.Close()

	m.lastIPFS = latest

	// Read morphFile
	cacheConfig, _ := os.ReadFile(m.morphFile)

	if strings.Compare(string(onChainConfig), string(cacheConfig)) == 0 {
		m.log.Info("There are no updated configuration")
		return nil
	}

	file, err := os.Create(m.morphFile)
	if err != nil {
		m.log.WithError(err).Error("Failed creating in-memory config to the file:" + m.morphFile)
		return err
	}
	defer file.Close()
	_, err = file.Write(onChainConfig)
	if err != nil {
		m.log.WithError(err).Error("Failed writing in-memory config to the file:" + m.morphFile)
		return err
	}

	m.log.Info("Found that on-chain configuration has been updated")

	return nil
}

func (m *Morph) reloadRoutine() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.interval.TickCh():
			err := m.Monitor()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (m *Morph) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}
