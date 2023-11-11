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
	"github.com/libp2p/go-libp2p/core/event"
	cp "github.com/libp2p/go-libp2p/core/peer"

	"github.com/chronicleprotocol/oracle-suite/rail/node"
	"github.com/chronicleprotocol/oracle-suite/rail/stats"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/model"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

type peers struct {
	list     map[cp.ID]stats.Peer
	ordering []cp.ID
}

func (p *peers) add(np stats.PeerEvent) {
	if p.list == nil {
		p.list = make(map[cp.ID]stats.Peer)
	}
	switch t := np.(type) {
	case event.EvtPeerIdentificationCompleted:
		pid := t.Peer
		if old, ok := p.list[pid]; ok {
			old.Identification = "completed"

			p.list[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.list[pid] = stats.Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Identification: "completed",
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, pid)
	case node.PeerConnectedness:
		pid := t.Peer
		if old, ok := p.list[pid]; ok {
			old.Connectedness = t.Connectedness.String()
			old.UserAgent = t.UserAgent

			p.list[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.list[pid] = stats.Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Connectedness: t.Connectedness.String(),
			UserAgent:     t.UserAgent,
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, pid)
	case node.LocalReachability:
		pid := t.Peer
		if old, ok := p.list[pid]; ok {
			old.Connectedness = t.Reachability.String()

			p.list[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.list[pid] = stats.Peer{
			ID:      pid,
			Addr:    a,
			Network: "local",
			Name:    node.BootName(pid),

			Connectedness: t.Reachability.String(),
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, pid)
	case node.PingResult:
		rtt := t.RTT.String()
		if t.Error != nil {
			rtt = t.Error.Error()
		}
		pid := t.Peer
		if old, ok := p.list[pid]; ok {
			old.Ping = rtt
			p.list[pid] = old
			return
		}
		a := ethkey.PeerIDToAddress(pid)
		p.list[pid] = stats.Peer{
			ID:      pid,
			Addr:    a,
			Network: node.Net(a),
			Name:    node.BootName(pid),

			Ping: rtt,
		}
		p.ordering = sliceutil.AppendUnique(p.ordering, pid)
	case node.PeerMessage:
		func(pid cp.ID) {
			if old, ok := p.list[pid]; ok {
				old.Counter++
				old.Topic = *t.Msg.Topic

				p.list[pid] = old
				return
			}
			a := ethkey.PeerIDToAddress(pid)
			p.list[pid] = stats.Peer{
				ID:      pid,
				Addr:    a,
				Network: node.Net(a),
				Name:    node.BootName(pid),

				Topic: *t.Msg.Topic,
			}
			p.ordering = sliceutil.AppendUnique(p.ordering, pid)
		}(t.Msg.GetFrom())

		func(pid cp.ID) {
			if old, ok := p.list[pid]; ok {
				old.Counter++

				p.list[pid] = old
				return
			}
			a := ethkey.PeerIDToAddress(pid)
			p.list[pid] = stats.Peer{
				ID:      pid,
				Addr:    a,
				Network: node.Net(a),
				Name:    node.BootName(pid),
			}
			p.ordering = sliceutil.AppendUnique(p.ordering, pid)
		}(t.Msg.ReceivedFrom)
	}
}

func mapPeers(peers peers) rowpick.Data {
	cols := []table.Column{
		{Title: "#"},
		{Title: "Net"},
		{Title: "Name"},
		{Title: "UserAgent"},
		{Title: "Status"},
		// {Title: "Identification"},
		// {Title: "Ping"},
		{Title: "Counter"},
		{Title: "Topic"},
	}
	for i := range cols {
		cols[i].Width = len(cols[i].Title)
	}

	rows := make([]table.Row, 0, len(peers.ordering))
	for i, p := range peers.ordering {
		id := peers.list[p].ID.String()
		if peers.list[p].Network != "" && peers.list[p].Network != "local" {
			id = peers.list[p].Addr.String()
		}
		if peers.list[p].Name != "" {
			id = peers.list[p].Name
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			peers.list[p].Network,
			id,
			peers.list[p].UserAgent,
			peers.list[p].Connectedness,
			// peers.list[p].Identification,
			// peers.list[p].Ping,
			fmt.Sprintf("%d", peers.list[p].Counter),
			peers.list[p].Topic,
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
