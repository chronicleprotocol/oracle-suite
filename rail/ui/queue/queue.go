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

package queue

import (
	"github.com/charmbracelet/bubbletea"
)

// Queue is a slice of tea.Cmd functions with some additional methods
type Queue []tea.Cmd

// Cmd returns a Queue with the given commands
func (q Queue) Cmd(cc ...tea.Cmd) Queue {
	for _, c := range cc {
		if c != nil {
			q = append(q, c)
		}
	}
	return q
}

// Msg returns a Queue with the given messages wrapped in tea.Cmd functions
func (q Queue) Msg(mm ...tea.Msg) Queue {
	for _, m := range mm {
		if m != nil {
			q = q.Cmd(func() tea.Msg { return m })
		}
	}
	return q
}

// Rev returns a Queue with the commands in reverse order
func (q Queue) Rev() Queue {
	var qq Queue
	for i := len(q) - 1; i >= 0; i-- {
		qq = qq.Cmd(q[i])
	}
	return qq
}

// Seq returns a tea.Cmd that runs the commands in sequence
func (q Queue) Seq() tea.Cmd {
	if len(q) == 1 {
		return q[0]
	}
	return tea.Sequence(q...)
}

// Bat returns a tea.Cmd that runs the commands concurrently with no ordering guarantees
func (q Queue) Bat() tea.Cmd {
	if len(q) == 1 {
		return q[0]
	}
	return tea.Batch(q...)
}

// Cmd returns a Queue with the given commands
// Think of it like "queue some commands for me"
func Cmd(v ...tea.Cmd) Queue {
	return Queue{}.Cmd(v...)
}

// Msg returns a Queue with the given messages wrapped in a tea.Cmd
// Think of it like "queue some messages for me"
func Msg(v ...tea.Msg) Queue {
	return Queue{}.Msg(v...)
}
