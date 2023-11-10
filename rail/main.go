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
	"context"
	"os"
	"os/signal"
	"sync"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/chronicleprotocol/oracle-suite/rail/metrics"
	"github.com/chronicleprotocol/oracle-suite/rail/node"
	"github.com/chronicleprotocol/oracle-suite/rail/ui"
)

var log = logging.Logger("rail")

func main() {
	logging.SetLogLevel("rail", "DEBUG")
	// logging.SetLogLevel("rail/metrics", "DEBUG")
	// logging.SetLogLevel("rail/service", "DEBUG")
	// logging.SetLogLevel("rail/node", "DEBUG")
	// logging.SetLogLevel("rail/ui", "DEBUG")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	options := []libp2p.Option{
		libp2p.Ping(false),
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
		libp2p.EnableRelay(),
		libp2p.EnableRelayService(),
		// libp2p.EnableAutoRelayWithStaticRelays(),
		// libp2p.EnableAutoRelayWithPeerSource(),
		node.Bootstraps(ctx, os.Args[1:]),
		node.Seed(),
	}

	eventChan := make(chan any, 100)
	actions := []node.Action{
		// node.LogListeningAddresses,
		// node.LogEvents,
		node.Gossip(ctx),
		node.Events(eventChan),
	}

	if false {
		idChan := make(chan peer.ID)
		defer close(idChan)
		actions = append(actions,
			node.Pinger(ctx, idChan, eventChan),
			node.ExtractIDs(idChan),
		)
	}

	runServicesAndWait(
		ctx,
		&metrics.Prometheus{},
		node.NewNode(options...)(actions...),
		ui.NewProgram(eventChan),
	)
}

func runServicesAndWait(ctx context.Context, services ...service) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, s := range services {
		log.Debugf("start %T", s)
		if err := s.Start(ctx); err != nil {
			log.Fatal(err)
		}
		log.Debugf("started %T", s)
	}

	var wg sync.WaitGroup
	wg.Add(len(services))
	for _, s := range services {
		go func(s service) {
			defer wg.Done()
			log.Debugf("wait %T", s)
			s.Wait()
			cancel()
			log.Debugf("done %T", s)
		}(s)
	}

	wg.Wait()
	log.Debug("all services finished")
}

type service interface {
	Start(ctx context.Context) error
	Wait()
}
