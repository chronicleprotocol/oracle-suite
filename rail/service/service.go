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

package service

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
)

var log = logging.Logger("rail/service")

type StartFn func(context.Context) error
type PostNodeFn func(...Action) PostMeshFn
type PostMeshFn func(...Mesh) StartFn

func Railing(opts ...libp2p.Option) PostNodeFn {
	return func(acts ...Action) PostMeshFn {
		return func(mesh ...Mesh) StartFn {
			return func(ctx context.Context) error {
				s := &Rail{
					opts:     opts,
					acts:     acts,
					postMesh: mesh,
					errCh:    make(chan error),
				}
				if err := s.Start(ctx); err != nil {
					return err
				}
				return <-s.Wait()
			}
		}
	}
}

type Rail struct {
	host  host.Host
	opts  []libp2p.Option
	ctx   context.Context
	errCh chan error

	acts     []Action
	postMesh []Mesh
}

func (s *Rail) Start(ctx context.Context) (err error) {
	log.Info("starting P2P")
	s.ctx = ctx
	s.host, err = libp2p.New(s.opts...)
	if err != nil {
		return err
	}
	go func() {
		<-s.ctx.Done()
		log.Info("stopping P2P")
		s.errCh <- s.host.Close()
		close(s.errCh)
	}()
	for _, act := range s.acts {
		if err := act(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *Rail) Wait() <-chan error {
	return s.errCh
}
