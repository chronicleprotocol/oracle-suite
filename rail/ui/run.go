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
	"context"

	tea "github.com/charmbracelet/bubbletea"
	logging "github.com/ipfs/go-log/v2"

	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

var log = logging.Logger("rail/ui")

func Appling(eventChan <-chan any) *App {
	return &App{
		events: eventChan,
		program: tea.NewProgram(app{
			table: rowpick.NewModel(),
		}),
	}
}

type App struct {
	ctx context.Context
	err error

	events <-chan any

	model   tea.Model
	program *tea.Program
}

func (s *App) Start(ctx context.Context) {
	log.Debugf("starting %#v", s)
	s.ctx = ctx
	go s.eventLoop()
	go func() { s.model, s.err = s.program.Run() }()
}

func (s *App) Wait() error {
	s.program.Wait()
	log.Debugf("stopping %#v", s)
	return s.err
}

func (s *App) eventLoop() {
	for e := range s.events {
		s.program.Send(e)
	}
}
