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

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chronicleprotocol/oracle-suite/rail/stats"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/model"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/queue"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

const (
	state0     model.State = ""
	stateLog   model.State = "log"
	statePeers model.State = "peers"
	stateQuit  model.State = "quit"
)

type app struct {
	model.Logs
	model.State

	table rowpick.Model

	peers peers
}

func (a app) Init() tea.Cmd {
	return nil
}

func (a app) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case model.Transition:
		a.LogF("%#v", msg)
		if !msg.WasUsed() {
			a.State = msg.State()
			return a, msg.Used
		}

	case tea.KeyMsg:
		a.LogF("keypress: %s", msg)
		switch msg.String() {
		case "esc", "enter":
			if a.State == stateLog {
				return a, a.Next(statePeers)
			}
			return a, a.Next(stateLog)
		case "q", "ctrl+c":
			return a, a.Next(stateQuit)
		}

	case tea.WindowSizeMsg: // whoever is interested in window size should get this message in Update(...)
		var c tea.Cmd
		a.table, c = a.table.Update(msg)
		return a, c

	case stats.PeerEvent:
		a.LogF("PeerEvent: %#v", msg)
		a.peers.add(msg)
		message = mapPeers(a.peers)

	default:
		a.LogF("unknown message: %#v", msg)
	}

	switch a.State {
	case state0:
		return a, a.Next(statePeers)
	case statePeers:
		return a.doPeers(message)
	case stateQuit:
		return a, queue.Cmd(tea.ExitAltScreen, tea.Quit).Seq()
	}
	return a, nil
}

func (a app) View() string {
	switch a.State {
	case state0, stateLog, stateQuit:
		return a.String()
	case statePeers:
		return a.table.View()
	}
	return fmt.Sprintf("error: unknown state: %s", a.State)
}
