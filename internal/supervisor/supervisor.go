//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package supervisor

import (
	"context"
	"reflect"
)

// Service that could be managed by Supervisor.
type Service interface {
	// Start starts the service.
	Start(ctx context.Context) error
	// Wait returns a channel that is blocked while service is running.
	// When the service is stopped, the channel will be closed. If an error
	// occurs, an error will be sent to the channel before closing it.
	Wait() chan error
}

// Supervisor manages long-running services that implement the Service
// interface. If any of the managed services fail, all other services are
// stopped. This ensures that all services are running or none.
type Supervisor struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	started   bool
	waitCh    chan error
	services  []Service
}

// New returns a new instance of *Supervisor.
func New(ctx context.Context) *Supervisor {
	ctx, ctxCancel := context.WithCancel(ctx)
	return &Supervisor{ctx: ctx, ctxCancel: ctxCancel, waitCh: make(chan error)}
}

// Watch add one or more services to a supervisor. Services must be added
// before invoking the Start method, otherwise it panics.
func (s *Supervisor) Watch(services ...Service) {
	if s.started {
		panic("supervisor was already started")
	}
	s.services = append(s.services, services...)
}

// Start starts all watched services. It can be invoked only once, otherwise
// it panics.
func (s *Supervisor) Start() error {
	if s.started {
		panic("supervisor was already started")
	}
	s.started = true
	for _, srv := range s.services {
		if err := srv.Start(s.ctx); err != nil {
			s.ctxCancel()
			close(s.waitCh)
			return err
		}
	}
	go s.serviceWatcher()
	return nil
}

// Wait returns a channel that is blocked until at least one service is
// running. When all services are stopped, the channel will be closed.
// If an error occurs in any of the services, it will be sent to the
// channel before closing it.
func (s *Supervisor) Wait() chan error {
	return s.waitCh
}

func (s *Supervisor) serviceWatcher() {
	var err error
	for len(s.services) > 0 {
		cases := make([]reflect.SelectCase, len(s.services))
		for i, srv := range s.services {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(srv.Wait())}
		}
		n, v, _ := reflect.Select(cases)
		if !v.IsNil() {
			if err == nil {
				err = v.Interface().(error)
			}
			s.ctxCancel()
		}
		s.services = append(s.services[:n], s.services[n+1:]...)
	}
	if err != nil {
		s.waitCh <- err
	}
	close(s.waitCh)
}
