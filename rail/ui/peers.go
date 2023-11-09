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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/libp2p/go-libp2p/core/network"
	cp "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/chronicleprotocol/oracle-suite/rail/ui/model"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

type peer struct {
	ID            cp.ID
	Addrs         []ma.Multiaddr
	Connectedness network.Connectedness
	Reachability  network.Reachability
}

type peers map[cp.ID]peer

func (a app) doValues(message tea.Msg) (tea.Model, tea.Cmd) {
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
