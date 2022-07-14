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

package ethereum

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/rpcsplitter"
)

const splitterVirtualHost = "makerdao-splitter"
const defaultTotalTimeout = 10
const defaultGracefulTimeout = 1

var ethClientFactory = func(
	endpoints []string,
	timeout,
	gracefulTimeout time.Duration,
	maxBlocksBehind int,
	logger log.Logger,
) (geth.EthClient, error) {

	// In theory, we don't need to use RPC-Splitter for a single endpoint, but
	// to make the application behavior consistent we use it.
	switch len(endpoints) {
	case 0:
		return nil, errors.New("missing address to a RPC client in the configuration file")
	default:
		splitter, err := rpcsplitter.NewTransport(
			splitterVirtualHost,
			nil,
			rpcsplitter.WithEndpoints(endpoints),
			rpcsplitter.WithTotalTimeout(timeout),
			rpcsplitter.WithGracefulTimeout(gracefulTimeout),
			rpcsplitter.WithRequirements(minimumRequiredResponses(len(endpoints)), maxBlocksBehind),
			rpcsplitter.WithLogger(logger),
		)
		if err != nil {
			return nil, err
		}
		rpcClient, err := rpc.DialHTTPWithClient(
			fmt.Sprintf("http://%s", splitterVirtualHost),
			&http.Client{Transport: splitter},
		)
		if err != nil {
			return nil, err
		}
		return ethclient.NewClient(rpcClient), nil
	}
}

type Ethereum struct {
	From            string      `json:"from"`
	Keystore        string      `json:"keystore"`
	Password        string      `json:"password"`
	RPC             interface{} `json:"rpc"`
	Timeout         int         `json:"timeout"`
	GracefulTimeout int         `json:"gracefulTimeout"`
	MaxBlocksBehind int         `json:"maxBlocksBehind"`
}

func (c *Ethereum) ConfigureSigner() (ethereum.Signer, error) {
	account, err := c.configureAccount()
	if err != nil {
		return nil, err
	}
	return geth.NewSigner(account), nil
}

func (c *Ethereum) ConfigureRPCClient(logger log.Logger) (geth.EthClient, error) {
	var endpoints []string
	switch v := c.RPC.(type) {
	case string:
		endpoints = []string{v}
	case []interface{}:
		for _, s := range v {
			if s, ok := s.(string); ok {
				endpoints = append(endpoints, s)
			}
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.New("ethereum config: value of the RPC key must be string or array of strings")
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = defaultTotalTimeout
	}
	if timeout < 1 {
		return nil, errors.New("ethereum config: timeout cannot be less than 1 (or 0 to use the default value)")
	}
	gracefulTimeout := c.GracefulTimeout
	if gracefulTimeout == 0 {
		gracefulTimeout = defaultGracefulTimeout
	}
	if gracefulTimeout < 1 {
		return nil, errors.New("ethereum config: gracefulTimeout cannot be less than 1 (or 0 to use the default value)")
	}
	maxBlocksBehind := c.MaxBlocksBehind
	if c.MaxBlocksBehind < 0 {
		return nil, errors.New("ethereum config: maxBlocksBehind cannot be less than 0")
	}
	return ethClientFactory(
		endpoints,
		time.Second*time.Duration(timeout),
		time.Second*time.Duration(gracefulTimeout),
		maxBlocksBehind,
		logger,
	)
}

func (c *Ethereum) ConfigureEthereumClient(signer ethereum.Signer, logger log.Logger) (*geth.Client, error) {
	client, err := c.ConfigureRPCClient(logger)
	if err != nil {
		return nil, err
	}
	return geth.NewClient(client, signer), nil
}

func (c *Ethereum) configureAccount() (*geth.Account, error) {
	if c.From == "" {
		return nil, nil
	}
	passphrase, err := c.readAccountPassphrase(c.Password)
	if err != nil {
		return nil, err
	}
	account, err := geth.NewAccount(c.Keystore, passphrase, ethereum.HexToAddress(c.From))
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Ethereum) readAccountPassphrase(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	passphrase, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read Ethereum password file: %w", err)
	}
	return strings.TrimSuffix(string(passphrase), "\n"), nil
}

func minimumRequiredResponses(endpoints int) int {
	if endpoints < 2 {
		return endpoints
	}
	return endpoints - 1
}
