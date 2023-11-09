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

// Package model is for implementing a simple state machine.
// In a [tea.Model.Update] method one can use a `case` for a [Transition] message to change the state.
//
//	case state.Transition:
//	if !msg.Used() {
//		m.State = msg.State()
//		return m, msg.AsUsed()
//	}
package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

// State machine
type State string

// Next creates a command that returns a Transition message with previous and next states
func (s State) Next(ns State) tea.Cmd {
	return func() tea.Msg { return Transition{prev: s, state: ns} }
}

// Transition is a message used to indicate a Transition between states
// The used field is to indicate that the message has been seen and processed
type Transition struct {
	prev  State
	state State
	used  bool
}

// WasUsed returns true if the Transition has been used
func (t Transition) WasUsed() bool {
	return t.used
}

// State returns the desired state after the Transition
func (t Transition) State() State {
	return t.state
}

// Used returns a tea.Msg with a Transition that is marked as used
func (t Transition) Used() tea.Msg {
	t.used = true
	return t
}
