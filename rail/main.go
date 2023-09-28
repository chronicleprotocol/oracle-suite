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

package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"os/signal"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/chronicleprotocol/oracle-suite/rail/service"
)

var log = logging.Logger("rail")

func env(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return v
}

func main() {
	separator := env("CFG_ITEM_SEPARATOR", "\n")

	{
		// logging.SetAllLoggers(logging.LevelInfo)
		logLevel, ok := os.LookupEnv("CFG_LOG_LEVEL")
		if ok {
			if err := logging.SetLogLevel("rail", logLevel); err != nil {
				log.Fatal(err)
			}
			if err := logging.SetLogLevel("rail/service", logLevel); err != nil {
				log.Fatal(err)
			}
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	// Options to configure libp2p node with
	options := []libp2p.Option{
		libp2p.ListenAddrStrings([]string{
			"/ip4/0.0.0.0/tcp/8000",
			"/ip4/0.0.0.0/udp/8000/quic-v1",
			"/ip4/0.0.0.0/udp/8000/quic-v1/webtransport",
			"/ip6/::/tcp/8000",
			"/ip6/::/udp/8000/quic-v1",
			"/ip6/::/udp/8000/quic-v1/webtransport",
		}...),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.Ping(false),
	}

	// Things to do when libp2p node is ready
	actions := []service.Action{
		service.LogListeningAddresses,
		service.LogEvents,
	}

	{
		addrs := os.Args[1:]
		if len(addrs) == 0 {
			addrs = defaultBoots
		}
		boots := addrInfos(addrs)
		options = append(options, service.Bootstrap(ctx, boots...))
	}

	{
		var gSubOpts []pubsub.Option
		directPeersVal, ok := os.LookupEnv("CFG_LIBP2P_DIRECT_PEERS_ADDRS")
		if ok && len(directPeersVal) > 0 {
			gSubOpts = append(gSubOpts, pubsub.WithDirectPeers(addrInfos(strings.Split(directPeersVal, separator))))
		}
		actions = append(actions, service.GossipSub(ctx, gSubOpts...))
	}

	{
		idChan := make(chan peer.ID)
		actions = append(actions,
			service.Pinger(ctx, idChan),
			service.ExtractIDs(idChan),
		)
	}

	{
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			if err := http.ListenAndServe(":8080", nil); err != nil {
				log.Error(err)
			}
		}()
	}

	{
		seedReader := rand.Reader
		if seed := os.Getenv("CFG_LIBP2P_PK_SEED"); seed != "" {
			seed, err := hex.DecodeString(seed)
			if err != nil {
				log.Fatal(err)
			}
			if len(seed) != ed25519.SeedSize {
				log.Fatal("invalid seed length - needs to be 32 bytes")
			}
			seedReader = bytes.NewReader(seed)
		}
		sk, _, err := crypto.GenerateEd25519Key(seedReader)
		if err != nil {
			log.Fatal(err)
		}
		options = append(options, libp2p.Identity(sk))
	}

	if err := service.Railing(options...)(actions...)()(ctx); err != nil {
		log.Error(err)
	}
}

var defaultBoots = []string{
	"/dns/spire-bootstrap1.chroniclelabs.io/tcp/8000/p2p/12D3KooWFYkJ1SghY4KfAkZY9Exemqwnh4e4cmJPurrQ8iqy2wJG",
	"/dns/spire-bootstrap2.chroniclelabs.io/tcp/8000/p2p/12D3KooWD7eojGbXT1LuqUZLoewRuhNzCE2xQVPHXNhAEJpiThYj",
	"/dns/spire-bootstrap1.staging.chroniclelabs.io/tcp/8000/p2p/12D3KooWHoSyTgntm77sXShoeX9uNkqKNMhHxKtskaHqnA54SrSG",
	"/ip4/178.128.141.30/tcp/8000/p2p/12D3KooWLaMPReGaxFc6Z7BKWTxZRbxt3ievW8Np7fpA6y774W9T",
	"/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi",
	"/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR",
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
