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

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/makerdao/oracle-suite/internal/rpcsplitter"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
	"github.com/makerdao/oracle-suite/pkg/ethereum/geth"
	"github.com/makerdao/oracle-suite/pkg/log/null"
)

const splitterVirtualHost = "makerdao-splitter"

var ethClientFactory = func(endpoints []string) (geth.EthClient, error) {
	switch len(endpoints) {
	case 0:
		return nil, errors.New("missing address to a RPC client in the configuration file")
	case 1:
		return ethclient.Dial(endpoints[0])
	default:
		// TODO: pass logger
		splitter, err := rpcsplitter.NewTransport(endpoints, splitterVirtualHost, nil, null.New())
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
	From     string      `json:"from"`
	Keystore string      `json:"keystore"`
	Password string      `json:"password"`
	RPC      interface{} `json:"rpc"`
}

func (c *Ethereum) ConfigureSigner() (ethereum.Signer, error) {
	account, err := c.configureAccount()
	if err != nil {
		return nil, err
	}
	return geth.NewSigner(account), nil
}

func (c *Ethereum) ConfigureRPCClient() (geth.EthClient, error) {
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
		return nil, errors.New("value of the RPC key must be string or array of strings")
	}
	return ethClientFactory(endpoints)
}

func (c *Ethereum) ConfigureEthereumClient(signer ethereum.Signer) (*geth.Client, error) {
	client, err := c.ConfigureRPCClient()
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
