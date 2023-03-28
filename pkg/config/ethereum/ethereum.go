package ethereum

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/rpc/transport"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/rpcsplitter"
)

const (
	splitterVirtualHost    = "rpc-splitter"
	defaultTotalTimeout    = 10
	defaultGracefulTimeout = 1
)

type (
	KeyRegistry    map[string]wallet.Key
	ClientRegistry map[string]rpc.RPC
)

type Dependencies struct {
	// Logger is the logger that is used by RPC-Splitter.
	Logger log.Logger
}

// ConfigEthereum contains the configuration for Ethereum clients and keys.
type ConfigEthereum struct {
	// Keys is a list of Ethereum keys.
	Keys []ConfigKey `hcl:"key,block"`

	// RandKeys is a list of random keys.
	RandKeys []string `hcl:"rand_keys,optional"`

	// Clients is a list of Ethereum clients.
	Clients []ConfigClient `hcl:"client,block"`

	// Configured services:
	prepared bool
	keys     KeyRegistry
	clients  ClientRegistry
}

// ConfigKey contains the configuration for an Ethereum key.
type ConfigKey struct {
	// Name is the unique name of the key that can be referenced by other
	// services.
	Name string `hcl:"name,label"`

	// Address is the address of the key in hex format.
	Address string `hcl:"address"`

	// KeystorePath is the path to the keystore directory.
	KeystorePath string `hcl:"keystore_path"`

	// Passphrase is the path to the file containing the passphrase for the
	// key. If empty, then the passphrase is not provided.
	PassphraseFile string `hcl:"passphrase_file,optional"`

	// Configured key:
	key wallet.Key
}

// ConfigClient contains the configuration for an Ethereum client.
type ConfigClient struct {
	// Name is the unique name of the client that can be referenced by other
	// services.
	Name string `hcl:"name,label"`

	// RPCURLs is a list of RPC URLs to use for the client. If multiple URLs
	// are provided, then RPC-Splitter will be used.
	RPCURLs []string `hcl:"rpc_urls"`

	// RPC-Splitter timeout settings:

	// Total timeout for the request, in seconds.
	Timeout int `hcl:"timeout,optional"`

	// GracefulTimeout is the time to wait for the response, in seconds, for
	// slower nodes after reaching minimum number of responses required for the
	// request.
	GracefulTimeout int `hcl:"graceful_timeout,optional"`

	// MaxBlocksBehind is the maximum number of blocks behind the node with the
	// highest block number can be. RPC-Splitter will use the lowest block number
	// from all nodes that is not more than MaxBlocksBehind behind the highest
	// block number.
	MaxBlocksBehind int `hcl:"max_blocks_behind,optional"`

	// Key configuration:

	// EthereumKey is the name of the Ethereum key to use for signing
	// transactions.
	EthereumKey string `hcl:"ethereum_key,optional"`

	// ChainID is the chain ID to use for signing transactions.
	ChainID uint64 `hcl:"chain_id,optional"`

	// Configured services:
	client rpc.RPC
}

// KeyRegistry returns the list of configured Ethereum keys.
func (e *ConfigEthereum) KeyRegistry(d Dependencies) (KeyRegistry, error) {
	if err := e.prepare(d); err != nil {
		return nil, err
	}
	return e.keys, nil
}

// ClientRegistry returns the list of configured Ethereum clients.
func (e *ConfigEthereum) ClientRegistry(d Dependencies) (ClientRegistry, error) {
	if err := e.prepare(d); err != nil {
		return nil, err
	}
	return e.clients, nil
}

func (e *ConfigEthereum) prepare(d Dependencies) error {
	if e.prepared {
		return nil
	}
	if err := e.prepareKeys(); err != nil {
		return err
	}
	if err := e.prepareClients(d.Logger); err != nil {
		return err
	}
	e.prepared = true
	return nil
}

func (e *ConfigEthereum) prepareKeys() error {
	e.keys = make(map[string]wallet.Key)
	for _, k := range e.Keys {
		if k.Name == "" {
			return errors.New("ethereum config: key name is required")
		}
		if _, ok := e.keys[k.Name]; ok {
			return fmt.Errorf("ethereum config: key with name %q already exists", k.Name)
		}
		key, err := k.Key()
		if err != nil {
			return err
		}
		e.keys[k.Name] = key
	}
	for _, k := range e.RandKeys {
		if k == "" {
			return errors.New("ethereum config: random key name is required")
		}
		if _, ok := e.keys[k]; ok {
			return fmt.Errorf("ethereum config: key with name %q already exists", k)
		}
		e.keys[k] = wallet.NewRandomKey()
	}
	return nil
}

func (e *ConfigEthereum) prepareClients(logger log.Logger) error {
	e.clients = make(map[string]rpc.RPC)
	for _, c := range e.Clients {
		if c.Name == "" {
			return errors.New("ethereum config: client name is required")
		}
		if _, ok := e.clients[c.Name]; ok {
			return fmt.Errorf("ethereum config: client with name %q already exists", c.Name)
		}
		client, err := c.Client(logger, e.keys)
		if err != nil {
			return err
		}
		e.clients[c.Name] = client
	}
	return nil
}

// Key returns the configured Ethereum key.
func (k *ConfigKey) Key() (wallet.Key, error) {
	if k.key != nil {
		return k.key, nil
	}
	addr, err := types.AddressFromHex(k.Address)
	if err != nil || addr.IsZero() {
		return nil, fmt.Errorf("ethereum config: invalid key address: %w", err)
	}
	passphrase, err := k.readAccountPassphrase(k.PassphraseFile)
	if err != nil {
		return nil, fmt.Errorf("ethereum config: failed to read key passphrase: %w", err)
	}
	key, err := wallet.NewKeyFromDirectory(k.KeystorePath, passphrase, addr)
	if err != nil {
		return nil, fmt.Errorf("ethereum config: failed to load key: %w", err)
	}
	k.key = key
	return key, nil
}

func (k *ConfigKey) readAccountPassphrase(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	passphrase, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(passphrase), "\n"), nil
}

// Client returns the configured RPC client.
func (c *ConfigClient) Client(logger log.Logger, keys KeyRegistry) (rpc.RPC, error) {
	if c.client != nil {
		return c.client, nil
	}
	rpcTransport, err := c.transport(logger)
	if err != nil {
		return nil, err
	}
	opts := []rpc.ClientOptions{rpc.WithTransport(rpcTransport)}
	if c.EthereumKey != "" {
		key, ok := keys[c.EthereumKey]
		if !ok {
			return nil, fmt.Errorf("ethereum config: ethereum_key %q not found", c.EthereumKey)
		}
		opts = append(
			opts,
			rpc.WithKeys(key),
			rpc.WithDefaultAddress(key.Address()),
		)
	}
	if c.ChainID != 0 {
		opts = append(opts, rpc.WithChainID(c.ChainID))
	}
	client, err := rpc.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("ethereum config: failed to create RPC client: %w", err)
	}
	c.client = client
	return client, nil
}

func (c *ConfigClient) transport(logger log.Logger) (transport.Transport, error) {
	if len(c.RPCURLs) == 0 {
		return nil, errors.New("ethereum config: value of the rpc_urls cannot be empty")
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
		return nil, errors.New("ethereum config: graceful_timeout cannot be less than 1 (or 0 to use the default value)")
	}
	maxBlocksBehind := c.MaxBlocksBehind
	if c.MaxBlocksBehind < 0 {
		return nil, errors.New("ethereum config: max_blocks_behind cannot be less than 0")
	}
	// In theory, we don't need to use RPC-Splitter for a single endpoint, but
	// to make the application behavior consistent we use it.
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
	rpcTransport, err := transport.NewHTTP(transport.HTTPOptions{
		URL:        fmt.Sprintf("http://%s", splitterVirtualHost),
		HTTPClient: &http.Client{Transport: splitter},
	})
	if err != nil {
		return nil, fmt.Errorf("ethereum config: failed to create RPC transport: %w", err)
	}
	return rpcTransport, nil
}

func minimumRequiredResponses(endpoints int) int {
	if endpoints < 2 {
		return endpoints
	}
	return endpoints - 1
}
