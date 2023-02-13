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

	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/net/proxy"

	suite "github.com/chronicleprotocol/oracle-suite"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/chain"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/recoverer"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Transport struct {
	P2P    *LibP2PConfig `hcl:"libp2p,block"`
	WebAPI *WebAPIConfig `hcl:"webapi,block"`
}

type LibP2PConfig struct {
	ListenAddrs      []string `hcl:"listen_addrs"`
	PrivKeySeed      string   `hcl:"priv_key_seed,optional"`
	BootstrapAddrs   []string `hcl:"bootstrap_addrs,optional"`
	DirectPeersAddrs []string `hcl:"direct_peers_addrs,optional"`
	BlockedAddrs     []string `hcl:"blocked_addrs,optional"`
	DisableDiscovery bool     `hcl:"disable_discovery,optional"`
}

type WebAPIConfig struct {
	ListenAddr      string                  `hcl:"listen_addr"`
	Socks5ProxyAddr string                  `hcl:"socks5_proxy_addr"`
	AddressBookAddr ethereum.Address        `hcl:"address_book_addr"`
	Ethereum        ethereumConfig.Ethereum `hcl:"ethereum,block"`
}

type Dependencies struct {
	Signer ethereum.Signer
	Feeds  []ethereum.Address
	Logger log.Logger
}

type BootstrapDependencies struct {
	Logger log.Logger
}

func (c *Transport) Configure(d Dependencies, t map[string]transport.Message) (transport.Transport, error) {
	var ts []transport.Transport
	if c.P2P != nil {
		t, err := c.configureLibP2P(d, t)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if c.WebAPI != nil {
		t, err := c.configureWebAPI(d, t)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		return nil, errors.New("no transports configured")
	}
	if len(ts) == 1 {
		return ts[0], nil
	}
	return chain.New(ts...), nil
}

func (c *Transport) ConfigureP2PBoostrap(d BootstrapDependencies) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	cfg := libp2p.Config{
		Mode:             libp2p.BootstrapMode,
		PeerPrivKey:      peerPrivKey,
		ListenAddrs:      c.P2P.ListenAddrs,
		BootstrapAddrs:   c.P2P.BootstrapAddrs,
		DirectPeersAddrs: c.P2P.DirectPeersAddrs,
		BlockedAddrs:     c.P2P.BlockedAddrs,
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

func (c *Transport) configureWebAPI(d Dependencies, t map[string]transport.Message) (transport.Transport, error) {
	httpClient := http.DefaultClient
	if len(c.WebAPI.Socks5ProxyAddr) != 0 {
		dialSocksProxy, err := proxy.SOCKS5("tcp", c.WebAPI.Socks5ProxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to the proxy: %w", err)
		}
		httpClient.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialSocksProxy.Dial(network, address)
			},
		}
	}
	cli, err := c.WebAPI.Ethereum.ConfigureEthereumClient(d.Signer, d.Logger)
	if err != nil {
		return nil, fmt.Errorf("cannot configure ethereum client: %w", err)
	}
	ab := webapi.NewEthereumAddressBook(cli, c.WebAPI.AddressBookAddr, time.Hour)
	tra, err := webapi.New(webapi.Config{
		ListenAddr:      c.WebAPI.ListenAddr,
		AddressBook:     ab,
		Topics:          t,
		AuthorAllowlist: d.Feeds,
		FlushTicker:     timeutil.NewTicker(time.Minute),
		Signer:          d.Signer,
		Client:          httpClient,
		Logger:          d.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot configure webapi transport: %w", err)
	}
	return recoverer.New(tra, d.Logger), nil
}

func (c *Transport) configureLibP2P(d Dependencies, t map[string]transport.Message) (transport.Transport, error) {
	peerPrivKey, err := c.generatePrivKey()
	if err != nil {
		return nil, err
	}
	var mPK crypto.PrivKey
	if d.Signer != nil && d.Signer.Address() != ethereum.EmptyAddress {
		mPK = ethkey.NewPrivKey(d.Signer)
	}
	cfg := libp2p.Config{
		Mode:             libp2p.ClientMode,
		PeerPrivKey:      peerPrivKey,
		Topics:           t,
		MessagePrivKey:   mPK,
		ListenAddrs:      c.P2P.ListenAddrs,
		BootstrapAddrs:   c.P2P.BootstrapAddrs,
		DirectPeersAddrs: c.P2P.DirectPeersAddrs,
		BlockedAddrs:     c.P2P.BlockedAddrs,
		AuthorAllowlist:  d.Feeds,
		Discovery:        !c.P2P.DisableDiscovery,
		Signer:           d.Signer,
		Logger:           d.Logger,
		AppName:          "spire",
		AppVersion:       suite.Version,
	}
	tra, err := libp2p.New(cfg)
	if err != nil {
		return nil, err
	}
	return recoverer.New(tra, d.Logger), nil
}

func (c *Transport) generatePrivKey() (crypto.PrivKey, error) {
	seedReader := rand.Reader
	if len(c.P2P.PrivKeySeed) != 0 {
		seed, err := hex.DecodeString(c.P2P.PrivKeySeed)
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
