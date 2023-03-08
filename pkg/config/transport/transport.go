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
	"strings"
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

const (
	LibP2P = "libp2p"
	LibSSB = "ssb"
	WebAPI = "webapi"
)

const (
	EthereumAddressBook = "ethereum"
	StaticAddressBook   = "static"
)

var p2pTransportFactory = func(cfg libp2p.Config) (transport.Transport, error) {
	return libp2p.New(cfg)
}

type Transport struct {
	Transport any               `yaml:"transport"`
	P2P       LibP2PConfig      `yaml:"libp2p"`
	SSB       ScuttlebuttConfig `yaml:"ssb"`
	WebAPI    WebAPIConfig      `yaml:"webapi"`
}

type LibP2PConfig struct {
	PrivKeySeed      string   `yaml:"privKeySeed"`
	ListenAddrs      []string `yaml:"listenAddrs"`
	BootstrapAddrs   []string `yaml:"bootstrapAddrs"`
	DirectPeersAddrs []string `yaml:"directPeersAddrs"`
	BlockedAddrs     []string `yaml:"blockedAddrs"`
	DisableDiscovery bool     `yaml:"disableDiscovery"`
}

type ScuttlebuttConfig struct {
	Caps string `yaml:"caps"`
}

type ScuttlebuttCapsConfig struct {
	Shs    string `yaml:"shs"`
	Sign   string `yaml:"sign"`
	Invite string `yaml:"invite,omitempty"`
}

type WebAPIConfig struct {
	ListenAddr          string                           `yaml:"listenAddr"`
	Socks5ProxyAddr     string                           `yaml:"socks5ProxyAddr"`
	AddressBookType     any                              `yaml:"addressBookType"`
	EthereumAddressBook *WebAPIEthereumAddressBookConfig `yaml:"ethereumAddressBook"`
	StaticAddressBook   *WebAPIStaticAddressBookConfig   `yaml:"staticAddressBook"`
}

type WebAPIStaticAddressBookConfig struct {
	RemoteAddrs []string `yaml:"remoteAddrs"`
}

type WebAPIEthereumAddressBookConfig struct {
	AddressBookAddr ethereum.Address `yaml:"addressBookAddr"`
	Ethereum        ethereumConfig.Ethereum
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
	var types []string
	switch varType := c.Transport.(type) {
	case string:
		types = []string{varType}
	case []any:
		for _, t := range varType {
			if s, ok := t.(string); ok {
				types = append(types, s)
				continue
			}
			return nil, fmt.Errorf("transport config error: invalid transport type: %v", t)
		}
	case []string:
		types = varType
	case nil:
		types = []string{LibP2P}
	default:
		return nil, fmt.Errorf("transport config error: invalid transport type: %v", varType)
	}
	switch len(types) {
	case 1:
		return c.configureTransport(d, types[0], t)
	default:
		var ts []transport.Transport
		for _, typ := range types {
			t, err := c.configureTransport(d, typ, t)
			if err != nil {
				return nil, err
			}
			ts = append(ts, t)
		}
		return chain.New(ts...), nil
	}
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
	p, err := p2pTransportFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("transport config error: %w", err)
	}
	return p, nil
}

//nolint:funlen,gocyclo
func (c *Transport) configureTransport(
	d Dependencies,
	typ string,
	t map[string]transport.Message) (transport.Transport, error) {

	switch strings.ToLower(typ) {
	case LibSSB:
		return nil, errors.New("ssb not yet implemented")
	case WebAPI:
		if c.WebAPI.ListenAddr == "" {
			return nil, errors.New("webapi listen addr not set")
		}
		var types []string
		switch varType := c.WebAPI.AddressBookType.(type) {
		case string:
			types = []string{varType}
		case []any:
			for _, t := range varType {
				if s, ok := t.(string); ok {
					types = append(types, s)
					continue
				}
				return nil, fmt.Errorf("transport config error: invalid address book type: %v", t)
			}
		}
		var addressBooks []webapi.AddressBook
		for _, typ := range types {
			switch typ {
			case EthereumAddressBook:
				if c.WebAPI.EthereumAddressBook == nil {
					return nil, errors.New("ethereum address book config not set")
				}
				cli, err := c.WebAPI.EthereumAddressBook.Ethereum.ConfigureEthereumClient(d.Signer, d.Logger)
				if err != nil {
					return nil, fmt.Errorf("cannot configure ethereum client: %w", err)
				}
				addressBooks = append(
					addressBooks,
					webapi.NewEthereumAddressBook(cli, c.WebAPI.EthereumAddressBook.AddressBookAddr, time.Hour),
				)
			case StaticAddressBook:
				if c.WebAPI.StaticAddressBook == nil {
					return nil, errors.New("static address book config not set")
				}
				addressBooks = append(
					addressBooks,
					webapi.NewStaticAddressBook(c.WebAPI.StaticAddressBook.RemoteAddrs),
				)
			default:
				if c.WebAPI.AddressBookType == "" {
					return nil, errors.New("address book type not set")
				}
				return nil, fmt.Errorf("invalid address book type: %s", c.WebAPI.AddressBookType)
			}
		}
		var addressBook webapi.AddressBook
		switch len(addressBooks) {
		case 0:
			return nil, errors.New("no address book configured")
		case 1:
			addressBook = addressBooks[0]
		default:
			addressBook = webapi.NewMultiAddressBook(addressBooks...)
		}
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
		tra, err := webapi.New(webapi.Config{
			ListenAddr:      c.WebAPI.ListenAddr,
			AddressBook:     addressBook,
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
	case LibP2P:
		fallthrough
	default:
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
		tra, err := p2pTransportFactory(cfg)
		if err != nil {
			return nil, err
		}
		return recoverer.New(tra, d.Logger), nil
	}
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
