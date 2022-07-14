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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
)

const LibP2P = "libp2p"
const LibSSB = "ssb"
const DefaultTransport = LibP2P

var p2pTransportFactory = func(cfg libp2p.Config) (transport.Transport, error) {
	return libp2p.New(cfg)
}

type Transport struct {
	Transport string      `json:"transport"`
	P2P       P2P         `json:"libp2p"`
	SSB       Scuttlebutt `json:"ssb"`
}

type P2P struct {
	PrivKeySeed      string   `json:"privKeySeed"`
	ListenAddrs      []string `json:"listenAddrs"`
	BootstrapAddrs   []string `json:"bootstrapAddrs"`
	DirectPeersAddrs []string `json:"directPeersAddrs"`
	BlockedAddrs     []string `json:"blockedAddrs"`
	DisableDiscovery bool     `json:"disableDiscovery"`
}

type Scuttlebutt struct {
	Caps string `json:"caps"`
}

type Caps struct {
	Shs    string `json:"shs"`
	Sign   string `json:"sign"`
	Invite string `json:"invite,omitempty"`
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
	case LibP2P:
		fallthrough
	default:
		peerPrivKey, err := c.generatePrivKey()
		if err != nil {
			return nil, err
		}
		cfg := libp2p.Config{
			Mode:             libp2p.ClientMode,
			PeerPrivKey:      peerPrivKey,
			Topics:           t,
			MessagePrivKey:   ethkey.NewPrivKey(d.Signer),
			ListenAddrs:      c.P2P.ListenAddrs,
			BootstrapAddrs:   c.P2P.BootstrapAddrs,
			DirectPeersAddrs: c.P2P.DirectPeersAddrs,
			BlockedAddrs:     c.P2P.BlockedAddrs,
			FeedersAddrs:     d.Feeds,
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
