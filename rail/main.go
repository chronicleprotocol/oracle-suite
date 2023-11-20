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

	logging "github.com/ipfs/go-log/v2"

	"github.com/chronicleprotocol/oracle-suite/rail/com"
	"github.com/chronicleprotocol/oracle-suite/rail/metrics"
	"github.com/chronicleprotocol/oracle-suite/rail/node"
	"github.com/chronicleprotocol/oracle-suite/rail/ssh"
)

var log = logging.Logger("rail")

func main() {
	logging.SetLogLevel("rail", "DEBUG")
	// logging.SetLogLevel("rail/metrics", "DEBUG")
	// logging.SetLogLevel("rail/service", "DEBUG")
	// logging.SetLogLevel("rail/node", "DEBUG")
	// logging.SetLogLevel("rail/node.Pinger", "DEBUG")
	// logging.SetLogLevel("rail/ui", "DEBUG")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	eventChan := make(chan any, 100)
	defer close(eventChan)

	var actions []node.Action
	{
		idExtractor, idChan := node.IDExtractor()
		actions = []node.Action{
			idExtractor,
			node.PingIDsIntoChan(idChan, eventChan),
			node.MessagesIntoChan(eventChan),
			node.EventsIntoChan(eventChan),
			// node.LogEvents,
			// node.LogListeningAddresses,
		}
	}

	com.RunServicesAndWait(
		ctx,
		&metrics.Prometheus{},
		node.NewNode(ctx, os.Args[1:], actions),
		// ui.NewProgram(eventChan),
		ssh.NewServer(eventChan),
	)
}
