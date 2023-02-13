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
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereumv2/rpcsplitter"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const splitterVirtualHost = "makerdao-splitter"
const defaultTotalTimeout = 10
const defaultGracefulTimeout = 1

type Ethereum struct {
	RPCURLs         []string `hcl:"rpc_urls"`
	From            string   `hcl:"from,optional"`
	Keystore        string   `hcl:"keystore,optional"`
	Password        string   `hcl:"password,optional"`
	Timeout         int      `hcl:"timeout,optional"`
	GracefulTimeout int      `hcl:"graceful_timeout,optional"`
	MaxBlocksBehind int      `hcl:"max_blocks_behind,optional"`
}

func (c *Ethereum) ConfigureSigner() (ethereum.Signer, error) {
	account, err := c.configureAccount()
	if err != nil {
		return nil, err
	}
	return geth.NewSigner(account), nil
}

func (c *Ethereum) ConfigureRPCClient(logger log.Logger) (*rpc.Client, error) {
	if len(c.RPCURLs) == 0 {
		return nil, errors.New("ethereum config: value of the RPCURLs key must be string or array of strings")
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
	// In theory, we don't need to use RPCURLs-Splitter for a single endpoint, but
	// to make the application behavior consistent we use it.
	switch len(c.RPCURLs) {
	case 0:
		return nil, errors.New("missing address to a RPCURLs client in the configuration file")
	default:
		splitter, err := rpcsplitter.NewTransport(
			splitterVirtualHost,
			nil,
			rpcsplitter.WithEndpoints(c.RPCURLs),
			rpcsplitter.WithTotalTimeout(time.Second*time.Duration(timeout)),
			rpcsplitter.WithGracefulTimeout(time.Second*time.Duration(gracefulTimeout)),
			rpcsplitter.WithRequirements(minimumRequiredResponses(len(c.RPCURLs)), maxBlocksBehind),
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
		return rpcClient, nil
	}
}

func (c *Ethereum) ConfigureEthereumClient(signer ethereum.Signer, logger log.Logger) (*geth.Client, error) {
	client, err := c.ConfigureRPCClient(logger)
	if err != nil {
		return nil, err
	}
	return geth.NewClient(ethclient.NewClient(client), signer), nil
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
