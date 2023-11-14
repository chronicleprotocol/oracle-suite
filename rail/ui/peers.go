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
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p/crypto/ethkey"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/chronicleprotocol/oracle-suite/rail/stats"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/model"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

func mapModels(models stats.Stats, net string) rowpick.Data {
	cols := []table.Column{
		{Title: "#"},
		{Title: "Model"},
	}

	var peers []peer.ID
	for _, p := range models.PeerOrdering {
		if (net == "" && models.PeerModelCount[p] == 0) ||
			(net != "" && (models.Nets[net][p] == 0 || models.Peers[p].Network != net)) {
			continue
		}
		peers = append(peers, p)
		cols = append(cols, table.Column{Title: ethkey.PeerIDToAddress(p).String()[:4]})
	}
	for i := range cols {
		cols[i].Width = len(cols[i].Title)
	}

	modelsOrdered := maputil.SortKeys(models.Models, sort.Strings)
	rows := make([]table.Row, 0, len(modelsOrdered))
	for i, m := range modelsOrdered {
		row := table.Row{fmt.Sprintf("%d", i+1), m}
		for _, p := range peers {
			a := fmt.Sprintf("%d", models.Models[m][p])
			if a == "0" {
				a = ""
			}
			row = append(row, a)
		}
		rows = append(rows, row)
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

func mapPeers(peers stats.Stats, net string) rowpick.Data {
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

	rows := make([]table.Row, 0, len(peers.PeerOrdering))
	for i, p := range peers.PeerOrdering {
		if net != "" && peers.Peers[p].Network != net {
			continue
		}
		id := peers.Peers[p].ID.String()
		if peers.Peers[p].Network != "" && peers.Peers[p].Network != "local" {
			id = peers.Peers[p].Addr.String()
		}
		if peers.Peers[p].Name != "" {
			id = peers.Peers[p].Name
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			peers.Peers[p].Network,
			id,
			peers.Peers[p].UserAgent,
			peers.Peers[p].Connectedness,
			// peers.list[p].Identification,
			// peers.list[p].Ping,
			fmt.Sprintf("%d", peers.Peers[p].Counter),
			peers.Peers[p].Topic,
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
