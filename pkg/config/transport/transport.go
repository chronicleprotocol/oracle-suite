package transport

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/net/proxy"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/recoverer"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Dependencies struct {
	Keys     ethereum.KeyRegistry
	Clients  ethereum.ClientRegistry
	Messages map[string]transport.Message
	Feeds    []types.Address
	Logger   log.Logger
}

type BootstrapDependencies struct {
	Logger log.Logger
}

type ConfigTransport struct {
	LibP2P *libP2PConfig `hcl:"libp2p,block"`
	WebAPI *webAPIConfig `hcl:"webapi,block"`

	// Configured transport:
	transport transport.Transport
}

type libP2PConfig struct {
	// ListenAddrs is the list of listening addresses for libp2p node encoded
	// using the multiaddress format.
	ListenAddrs []string `hcl:"listen_addrs"`

	// PrivKeySeed is the random hex-encoded 32 bytes. It is used to generate
	// a unique identity on the libp2p network. The value may be empty to
	// generate a random seed.
	PrivKeySeed string `hcl:"priv_key_seed,optional"`

	// BootstrapAddrs is the list of bootstrap addresses for libp2p node
	// encoded using the multiaddress format.
	BootstrapAddrs []string `hcl:"bootstrap_addrs,optional"`

	// DirectPeersAddrs is the list of direct peer addresses to which messages
	// will be sent directly. Addresses are encoded using the format the
	// multiaddress format. This option must be configured symmetrically on
	// both ends.
	DirectPeersAddrs []string `hcl:"direct_peers_addrs,optional"`

	// BlockedAddrs is the list of blocked addresses encoded using the
	// multiaddress format.
	BlockedAddrs []string `hcl:"blocked_addrs,optional"`

	// DisableDiscovery disables node discovery. If enabled, the IP address of
	// a node will not be broadcast to other peers. This option must be used
	// together with `directPeersAddrs`.
	DisableDiscovery bool `hcl:"disable_discovery,optional"`

	// EthereumKey is the name of the Ethereum key to use for signing messages.
	// Required if the transport is used for sending messages.
	EthereumKey string `hcl:"ethereum_key,optional"`
}

type webAPIConfig struct {
	// ListenAddr is the address on which the WebAPI server will listen for
	// incoming connections. The address must be in the format `host:port`.
	// When used with a TOR hidden service, the server should listen on
	// localhost.
	ListenAddr string `hcl:"listen_addr"`

	// Socks5ProxyAddr is the address of the SOCKS5 proxy server. The address
	// must be in the format `host:port`.
	Socks5ProxyAddr string `hcl:"socks5_proxy_addr"`

	// EthereumKey is the name of the Ethereum key to use for signing messages.
	// Required if the transport is used for sending messages.
	EthereumKey string `hcl:"ethereum_key"`

	// AddressBook configuration. Address book provides a list of addresses
	// to which messages will be sent.

	// EthereumAddressBook is the configuration for the Ethereum address book.
	EthereumAddressBook *webAPIEthereumAddressBook `hcl:"ethereum_address_book,block"`

	// StaticAddressBook is the configuration for the static address book.
	StaticAddressBook *webAPIStaticAddressBook `hcl:"static_address_book,block"`
}

type webAPIEthereumAddressBook struct {
	// ContractAddr is the Ethereum address of the address book contract.
	ContractAddr string `hcl:"contract_addr"`

	// EthereumClient is the name of the Ethereum client to use for reading
	// the address book.
	EthereumClient string `hcl:"ethereum_client"`
}

type webAPIStaticAddressBook struct {
	// Addresses is the list of static addresses to which messages will be
	// sent.
	Addresses []string `hcl:"addresses"`
}

func (c *ConfigTransport) Transport(d Dependencies) (transport.Transport, error) {
	if c.transport != nil {
		return c.transport, nil
	}
	var transports []transport.Transport
	switch {
	case c.LibP2P != nil:
		t, err := c.configureLibP2P(d)
		if err != nil {
			return nil, err
		}
		transports = append(transports, t)
	case c.WebAPI != nil:
		t, err := c.configureWebAPI(d)
		if err != nil {
			return nil, err
		}
		transports = append(transports, t)
	}
	switch {
	case len(transports) == 0:
		return nil, errors.New("no transports configured")
	case len(transports) == 1:
		c.transport = transports[0]
	default:
		c.transport = chain.New(transports...)
	}
	return c.transport, nil
}

func (c *ConfigTransport) LibP2PBootstrap(d BootstrapDependencies) (transport.Transport, error) {
	if c.LibP2P == nil {
		return nil, errors.New("libp2p transport not configured")
	}
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	cfg := libp2p.Config{
		Mode:             libp2p.BootstrapMode,
		PeerPrivKey:      peerPrivKey,
		ListenAddrs:      c.LibP2P.ListenAddrs,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		Logger:           d.Logger,
		AppName:          "bootstrap",
		AppVersion:       suite.Version,
	}
	p, err := libp2p.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("transport config error: %w", err)
	}
	return p, nil
}

func (c *ConfigTransport) configureWebAPI(d Dependencies) (transport.Transport, error) {
	// Configure HTTP client:
	httpClient := http.DefaultClient
	if len(c.WebAPI.Socks5ProxyAddr) != 0 {
		dialer, err := proxy.SOCKS5("tcp", c.WebAPI.Socks5ProxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to the proxy: %w", err)
		}
		httpClient.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.Dial(network, address)
			},
		}
	}

	// Configure address book:
	var (
		addressBook  webapi.AddressBook
		addressBooks []webapi.AddressBook
	)
	switch {
	case c.WebAPI.EthereumAddressBook != nil:
		rpcClient := d.Clients[c.WebAPI.EthereumAddressBook.EthereumClient]
		if rpcClient == nil {
			return nil, fmt.Errorf("WebAPI config: ethereum client %q not found", c.WebAPI.EthereumAddressBook.EthereumClient)
		}
		contractAddr, err := types.AddressFromHex(c.WebAPI.EthereumAddressBook.ContractAddr)
		if err != nil || contractAddr == types.ZeroAddress {
			return nil, fmt.Errorf("WebAPI config: invalid contract address %q", c.WebAPI.EthereumAddressBook.ContractAddr)
		}
		addressBooks = append(addressBooks, webapi.NewEthereumAddressBook(
			geth.NewClient(rpcClient),
			contractAddr,
			time.Hour,
		))
	case c.WebAPI.StaticAddressBook != nil:
		addressBooks = append(
			addressBooks,
			webapi.NewStaticAddressBook(c.WebAPI.StaticAddressBook.Addresses),
		)
	}
	switch {
	case len(addressBooks) == 0:
		return nil, errors.New("WebAPI config: no address book configured")
	case len(addressBooks) == 1:
		addressBook = addressBooks[0]
	default:
		addressBook = webapi.NewMultiAddressBook(addressBooks...)
	}

	// Configure signer:
	key := d.Keys[c.WebAPI.EthereumKey]
	if key == nil {
		return nil, fmt.Errorf("WebAPI config: key %q not found", c.WebAPI.EthereumKey)
	}

	// Configure transport:
	webapiTransport, err := webapi.New(webapi.Config{
		ListenAddr:      c.WebAPI.ListenAddr,
		AddressBook:     addressBook,
		Topics:          d.Messages,
		AuthorAllowlist: d.Feeds,
		FlushTicker:     timeutil.NewTicker(time.Minute),
		Signer:          key,
		Client:          httpClient,
		Logger:          d.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("WebAPI config: %w", err)
	}
	return recoverer.New(webapiTransport, d.Logger), nil
}

func (c *ConfigTransport) configureLibP2P(d Dependencies) (transport.Transport, error) {
	// Configure LibP2P private keys:
	ethereumKey := d.Keys[c.LibP2P.EthereumKey] // Key is optional.
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, fmt.Errorf("LibP2P config: %w", err)
	}
	var messagePrivKey crypto.PrivKey
	if ethereumKey != nil {
		messagePrivKey = ethkey.NewPrivKey(ethereumKey)
	}

	// Configure LibP2P transport:
	cfg := libp2p.Config{
		Mode:             libp2p.ClientMode,
		PeerPrivKey:      peerPrivKey,
		Topics:           d.Messages,
		MessagePrivKey:   messagePrivKey,
		ListenAddrs:      c.LibP2P.ListenAddrs,
		BootstrapAddrs:   c.LibP2P.BootstrapAddrs,
		DirectPeersAddrs: c.LibP2P.DirectPeersAddrs,
		BlockedAddrs:     c.LibP2P.BlockedAddrs,
		AuthorAllowlist:  d.Feeds,
		Discovery:        !c.LibP2P.DisableDiscovery,
		Signer:           ethereumKey,
		Logger:           d.Logger,
		AppName:          "spire",
		AppVersion:       suite.Version,
	}
	libP2PTransport, err := libp2p.New(cfg)
	if err != nil {
		return nil, err
	}
	return recoverer.New(libP2PTransport, d.Logger), nil
}

func (c *ConfigTransport) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.LibP2P.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.LibP2P.PrivKeySeed)
		if err != nil {
			return nil, fmt.Errorf("invalid privKeySeed value, failed to decode hex data: %w", err)
		}
		if len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("invalid privKeySeed value, 32 bytes expected")
		}
		seedReader = bytes.NewReader(seed)
	}
	privKey, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		return nil, fmt.Errorf("invalid privKeySeed value, failed to generate key: %w", err)
	}
	return privKey, nil
}
