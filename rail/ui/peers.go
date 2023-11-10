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

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
	"github.com/defiweb/go-eth/types"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	cp "github.com/libp2p/go-libp2p/core/peer"

	"github.com/chronicleprotocol/oracle-suite/rail/node"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/model"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

type Peer struct {
	ID             cp.ID
	Addr           types.Address
	Connectedness  network.Connectedness
	Reachability   network.Reachability
	Network        string
	Identification string
	Ping           string
}
type PeerEvent any

type peers struct {
	list     map[cp.ID]Peer
	ordering []cp.ID
}

func (p *peers) Add(np PeerEvent) {
	if p.list == nil {
		p.list = make(map[cp.ID]Peer)
	}
	switch t := np.(type) {
	case event.EvtPeerIdentificationCompleted:
		if old, ok := p.list[t.Peer]; ok {
			old.Identification = "completed"
			p.list[t.Peer] = old
			return
		}
		a := ethkey.PeerIDToAddress(t.Peer)
		p.list[t.Peer] = Peer{
			ID:             t.Peer,
			Addr:           a,
			Network:        node.Feeds.Net(a),
			Identification: "completed",
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, t.Peer)
	case event.EvtPeerConnectednessChanged:
		if old, ok := p.list[t.Peer]; ok {
			old.Connectedness = t.Connectedness
			p.list[t.Peer] = old
			return
		}
		a := ethkey.PeerIDToAddress(t.Peer)
		p.list[t.Peer] = Peer{
			ID:            t.Peer,
			Addr:          a,
			Network:       node.Feeds.Net(a),
			Connectedness: t.Connectedness,
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, t.Peer)
	case node.PingResult:
		rtt := t.RTT.String()
		if t.Error != nil {
			rtt = t.Error.Error()
		}
		if old, ok := p.list[t.Peer]; ok {
			old.Ping = rtt
			p.list[t.Peer] = old
			return
		}
		a := ethkey.PeerIDToAddress(t.Peer)
		p.list[t.Peer] = Peer{
			ID:      t.Peer,
			Addr:    a,
			Network: node.Feeds.Net(a),
			Ping:    rtt,
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, t.Peer)
	}
}

func mapPeers(peers peers) rowpick.Data {
	cols := []table.Column{
		{Title: "#"},
		{Title: "ID"},
		{Title: "Addr"},
		{Title: "Connectedness"},
		{Title: "Identification"},
		{Title: "Ping"},
		{Title: "Network"},
	}
	for i := range cols {
		cols[i].Width = len(cols[i].Title) + 2
	}

	rows := make([]table.Row, 0, len(peers.ordering))
	for i, p := range peers.ordering {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			peers.list[p].ID.String(),
			peers.list[p].Addr.String(),
			peers.list[p].Connectedness.String(),
			peers.list[p].Identification,
			peers.list[p].Ping,
			peers.list[p].Network,
		})
	}

	return rowpick.Data{
		Cols: cols,
		Rows: rows,
		Mapper: func(m table.Model) tea.Cmd {
			return func() tea.Msg {
				return rowpick.Done{
					Idx: m.Cursor(),
					Row: m.SelectedRow(),
				}
			}
		},
	}
}

func (a app) doPeers(message tea.Msg) (tea.Model, tea.Cmd) {
	var c tea.Cmd

	switch msg := message.(type) {
	case model.Transition:
		a.LogF("entered values")
		return a, queue.Cmd(tea.EnterAltScreen).Seq()

	case rowpick.Done:
		a.LogF("#%d %s", msg.Idx, strings.Join(msg.Row, " "))
		return a, a.Next(stateQuit)
	}

	a.table, c = a.table.Update(message)
	return a, c
}
