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
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/middleware"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/twitter"
)

const (
	LibP2P  = "libp2p"
	LibSSB  = "libssb"
	Twitter = "twitter"
)

var p2pTransportFactory = func(cfg libp2p.Config) (transport.Transport, error) {
	return libp2p.New(cfg)
}

type Transport struct {
	Transport string            `yaml:"transport"`
	P2P       LibP2PConfig      `yaml:"libp2p"`
	SSB       ScuttlebuttConfig `yaml:"ssb"`
	Twitter   TwitterConfig     `yaml:"twitter"`
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

type TwitterConfig struct {
	Accounts       []string `yaml:"accounts"`
	ConsumerKey    string   `yaml:"consumerKey"`
	ConsumerSecret string   `yaml:"consumerSecret"`
	AccessToken    string   `yaml:"accessToken"`
	AccessSecret   string   `yaml:"accessSecret"`
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
	case Twitter:
		cfg := twitter.Config{
			Accounts:             c.Twitter.Accounts,
			ConsumerKey:          c.Twitter.ConsumerKey,
			ConsumerSecret:       c.Twitter.ConsumerSecret,
			AccessToken:          c.Twitter.AccessToken,
			AccessSecret:         c.Twitter.AccessSecret,
			Topics:               t,
			PostTweetsInterval:   time.Minute,
			FetchTweetsInterval:  time.Second * 15,
			QueueSize:            1024,
			MaximumDataSize:      100000,
			MaximumTweetLength:   250,
			MosaicType:           twitter.ImageTypeJPEG,
			MosaicBitsPerChannel: 2,
			MosaicBlockSize:      16,
			Logger:               d.Logger,
		}
		tw, err := twitter.New(cfg)
		if err != nil {
			return nil, err
		}
		return priceLimiterMiddleware(twitterMiddleware(tw)), nil
	case LibSSB:
		return nil, errors.New("ssb not yet implemented")
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

type TweeterPrice struct {
	*messages.Price
}

func (t *TweeterPrice) Tweet() string {
	f, _ := new(big.Float).SetInt(t.Price.Price.Val).Float64()
	return fmt.Sprintf("%s: %f", t.Price.Price.Wat, f/oracle.PriceMultiplier)
}

func twitterMiddleware(t transport.Transport) transport.Transport {
	m := middleware.New(t)
	m.Use(middleware.BroadcastMiddlewareFunc(func(_ context.Context, next middleware.BroadcastFunc) middleware.BroadcastFunc {
		return func(topic string, msg transport.Message) error {
			switch mt := msg.(type) {
			case *messages.Price:
				msg = &TweeterPrice{Price: mt}
			}
			return next(topic, msg)
		}
	}))
	return m
}

func priceLimiterMiddleware(t transport.Transport) transport.Transport {
	m := middleware.New(t)
	m.Use(middleware.BroadcastMiddlewareFunc(func(_ context.Context, next middleware.BroadcastFunc) middleware.BroadcastFunc {
		prices := make(map[string]*messages.Price)
		return func(topic string, msg transport.Message) error {
			if topic == messages.PriceV0MessageName {
				return nil
			}
			if price, ok := msg.(*messages.Price); ok {
				prev, ok := prices[price.Price.Wat]
				if !ok {
					prices[price.Price.Wat] = price
					return next(topic, msg)
				}
				if time.Since(prev.Price.Age) > time.Minute*10 {
					prices[price.Price.Wat] = price
					return next(topic, msg)
				}
				if spread(price.Price.Val, prev.Price.Val) > 1 {
					prices[price.Price.Wat] = price
					return next(topic, msg)
				}
				return nil
			}
			return next(topic, msg)
		}
	}))
	return m
}

func spread(a, b *big.Int) float64 {
	if a.Sign() == 0 || b.Sign() == 0 {
		return math.Inf(1)
	}

	oldPriceF := new(big.Float).SetInt(a)
	newPriceF := new(big.Float).SetInt(b)

	x := new(big.Float).Sub(newPriceF, oldPriceF)
	x = new(big.Float).Quo(x, oldPriceF)
	x = new(big.Float).Mul(x, big.NewFloat(100))
	xf, _ := x.Float64()

	return math.Abs(xf)
}
