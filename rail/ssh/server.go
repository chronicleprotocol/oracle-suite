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

package ssh

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"

	"github.com/chronicleprotocol/oracle-suite/rail/ui"
)

const (
	host = "0.0.0.0"
	port = 23234
)

func NewServer(eventChan <-chan any) *Server {
	return &Server{
		events: eventChan,
	}
}

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	events <-chan any

	model  tea.Model
	server *ssh.Server
}

func (s *Server) Start(ctx context.Context) error {
	if s.ctx != nil {
		return fmt.Errorf("already started %T", s)
	}
	if ctx == nil {
		return fmt.Errorf("nil context for %T", s)
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	var err error
	s.server, err = wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(ui.ProgramHandler(s.events), termenv.ANSI256),
			logging.Middleware(),
		),
	)
	if err != nil {
		return fmt.Errorf("could not init server: %w", err)
	}

	s.wg.Add(1)
	go s.listenAndServe()

	s.wg.Add(1)
	go s.shutdown()

	return nil
}

func (s *Server) Wait() {
	<-s.ctx.Done()
	log.Debugf("waited %T", s.ctx)

	// Cancel the context so that all goroutines exit
	s.cancel()
	log.Debugf("canceled %T", s.cancel)

	// Wait for all goroutines to exit
	s.wg.Wait()
	log.Debugf("waited %T", s.wg)
}

func (s *Server) listenAndServe() {
	defer s.wg.Done()
	defer s.cancel()

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not start server: ", err)
	}
}

func (s *Server) shutdown() {
	defer s.wg.Done()

	<-s.ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server: ", err)
	}
}
