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
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	logging "github.com/ipfs/go-log/v2"

	"github.com/chronicleprotocol/oracle-suite/rail/stats"
	"github.com/chronicleprotocol/oracle-suite/rail/ui/rowpick"
)

var log = logging.Logger("rail/ui")

func NewProgram(eventChan <-chan any) *Program {
	return &Program{
		events: eventChan,
	}
}

type Program struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	events <-chan any

	model   tea.Model
	program *tea.Program
}

func (s *Program) Start(ctx context.Context) error {
	if s.ctx != nil {
		return fmt.Errorf("already started %T", s)
	}
	if ctx == nil {
		return fmt.Errorf("nil context for %T", s)
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.program = tea.NewProgram(app{
		table: rowpick.NewModel(),
	}, tea.WithContext(s.ctx))

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer log.Debugf("gone %T", s.program)
		defer s.cancel()

		var err error
		s.model, err = s.program.Run()
		if err != nil {
			log.Error(err)
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer log.Debugf("gone %T", s.program)
		defer s.cancel()

		for {
			select {
			case <-s.ctx.Done():
				log.Debugf("kill %T", s.program)
				s.program.Quit()
				log.Debugf("killed %T", s.program)
				return
			case e := <-s.events:
				log.Debugf("recv %T", e)
				s.program.Send(stats.PeerEvent(e))
			}
		}
	}()
	return nil
}

func (s *Program) Wait() {
	// Wait until the program exits
	s.program.Wait()
	log.Debugf("waited %T", s.program)
	// Cancel the context so that all goroutines exit
	s.cancel()
	log.Debugf("canceled %T", s.cancel)
	// Wait for all goroutines to exit
	s.wg.Wait()
	log.Debugf("waited %T", s.wg)
}
