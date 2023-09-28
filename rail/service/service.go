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

func Railing(opts ...libp2p.Option) func(...Action) func(context.Context) error {
	return func(acts ...Action) func(context.Context) error {
		return func(ctx context.Context) error {
			s := &Rail{
				libOpts: opts,
				acts:    acts,
				wait:    make(chan error),
			}
			if err := s.Start(ctx); err != nil {
				return err
			}
			return <-s.Wait()
		}
	}
}

type Rail struct {
	libOpts []libp2p.Option

	host host.Host
	wait chan error

	acts []Action
}

func (s *Rail) Start(ctx context.Context) (err error) {
	log.Info("starting P2P")
	s.host, err = libp2p.New(s.libOpts...)
	if err != nil {
		return err
	}
	for _, act := range s.acts {
		if err := act(s); err != nil {
			return err
		}
	}
	go func() {
		<-ctx.Done()
		log.Info("stopping P2P")
		s.wait <- s.host.Close()
		close(s.wait)
	}()
	return nil
}

func (s *Rail) Wait() <-chan error {
	return s.wait
}
