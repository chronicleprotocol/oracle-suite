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

package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
)

func NewNode(ctx context.Context, boots []string, acts []Action) *Node {
	return &Node{
		opts: initialOptions(ctx, boots),
		acts: acts,
	}
}

type Node struct {
	ctx context.Context
	wg  sync.WaitGroup

	opts []libp2p.Option
	acts []Action

	host host.Host
}

func (s *Node) Start(ctx context.Context) error {
	if s.ctx != nil {
		return fmt.Errorf("already started %T", s)
	}
	if ctx == nil {
		return fmt.Errorf("nil context for %T", s)
	}
	s.ctx = ctx
	var err error
	s.host, err = libp2p.New(s.opts...)
	if err != nil {
		return err
	}
	for _, act := range s.acts {
		if err := act(s); err != nil {
			return err
		}
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		<-s.ctx.Done()
		log.Debugf("closing %T", s.host)
		if err := s.host.Close(); err != nil {
			log.Error(err)
		}
		log.Debugf("closed %T", s.host)
	}()
	return nil
}

func (s *Node) Wait() {
	s.wg.Wait()
}
