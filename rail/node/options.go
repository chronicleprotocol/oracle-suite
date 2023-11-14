//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package node

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"

	"github.com/chronicleprotocol/oracle-suite/rail/env"
)

var defaultListens = []string{
	"/ip4/0.0.0.0/tcp/8000",
	"/ip4/0.0.0.0/udp/8000/quic-v1",
	"/ip4/0.0.0.0/udp/8000/quic-v1/webtransport",
	"/ip6/::/tcp/8000",
	"/ip6/::/udp/8000/quic-v1",
	"/ip6/::/udp/8000/quic-v1/webtransport",
}

func initialOptions(ctx context.Context, boots []string) []libp2p.Option {
	return []libp2p.Option{
		libp2p.Ping(false),
		libp2p.ListenAddrStrings(defaultListens...),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
		libp2p.EnableRelayService(),
		// libp2p.EnableAutoRelayWithStaticRelays(),
		// libp2p.EnableAutoRelayWithPeerSource(),
		Bootstraps(ctx, boots),
		Seed(),
	}
}

func Bootstraps(ctx context.Context, addrs []string) libp2p.Option {
	if len(addrs) == 0 {
		addrs = env.Strings("CFG_LIBP2P_BOOTSTRAP_ADDRS", defaultBoots)
	}
	return bootstrap(ctx, addrInfos(addrs)...)
}

func bootstrap(ctx context.Context, boots ...peer.AddrInfo) libp2p.Option {
	return libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		if len(boots) == 0 {
			return nil, nil
		}
		log.Infow("creating DHT router", "boots", boots)
		return dual.New(
			ctx, host,
			dual.DHTOption(
				dht.Mode(dht.ModeAutoServer),
			),
			dual.WanDHTOption(
				dht.BootstrapPeers(boots...),
				dht.Mode(dht.ModeAutoServer),
			),
		)
	})
}

func Seed() libp2p.Option {
	seedReader := rand.Reader
	if seed := env.HexBytes("CFG_LIBP2P_PK_SEED", nil); seed != nil {
		if len(seed) != ed25519.SeedSize {
			log.Fatalf("invalid seed size - want: %d, got: %d", ed25519.SeedSize, len(seed))
		}
		seedReader = bytes.NewReader(seed)
	}
	sk, _, err := crypto.GenerateEd25519Key(seedReader)
	if err != nil {
		log.Fatalf("unable to generate key: %v", err)
	}
	return libp2p.Identity(sk)
}

func addrInfos(addrs []string) []peer.AddrInfo {
	var list []peer.AddrInfo
	for _, addr := range addrs {
		pi, err := peer.AddrInfoFromString(addr)
		if err != nil {
			log.Error(err)
			continue
		}
		list = append(list, *pi)
	}
	return list
}
