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
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/webapi"
)

const (
	LibP2P = "libp2p"
	LibSSB = "libssb"
	WebAPI = "webapi"
)

var p2pTransportFactory = func(cfg libp2p.Config) (transport.Transport, error) {
	return libp2p.New(cfg)
}

type Transport struct {
	Transport string            `yaml:"transport"`
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
	Ethereum              ethereumConfig.Ethereum `yaml:"ethereum"`
	ListenAddr            string                  `yaml:"listenAddr"`
	Socks5ProxyAddr       string                  `yaml:"socks5ProxyAddr"`
	ConsumersContractAddr ethereum.Address        `yaml:"consumersContractAddr"`
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
	switch strings.ToLower(c.Transport) {
	case LibSSB:
		return nil, errors.New("ssb not yet implemented")
	case WebAPI:
		httpClient := http.DefaultClient
		if len(c.WebAPI.Socks5ProxyAddr) != 0 {
			dialSocksProxy, err := proxy.SOCKS5("tcp", c.WebAPI.Socks5ProxyAddr, nil, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("transport config error: cannot connect to the proxy: %w", err)
			}
			httpClient.Transport = &http.Transport{
				DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
					return dialSocksProxy.Dial(network, address)
				},
			}
		}
		cli, err := c.WebAPI.Ethereum.ConfigureEthereumClient(d.Signer, d.Logger)
		if err != nil {
			return nil, fmt.Errorf("transport config error: cannot configure ethereum client: %w", err)
		}
		cc := webapi.NewEthereumAddressBook(cli, c.WebAPI.ConsumersContractAddr, time.Hour)
		return webapi.New(webapi.Config{
			ListenAddr:      c.WebAPI.ListenAddr,
			AddressBook:     cc,
			Topics:          t,
			AuthorAllowlist: d.Feeds,
			Signer:          d.Signer,
			Client:          httpClient,
			Logger:          d.Logger,
		})
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
		p, err := p2pTransportFactory(cfg)
		if err != nil {
			return nil, err
		}
		return p, nil
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
		return nil, err
	}
	return p, nil
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
