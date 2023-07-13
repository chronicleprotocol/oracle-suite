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

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/sysmon"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/libp2p"
)

type BootstrapConfig struct {
	Transport transport.Config `hcl:"transport,block"`
	Logger    *logger.Config   `hcl:"logger,block,optional"`

	Remain hcl.Body `hcl:",remain"` // To ignore unknown blocks.
}

func NewBootstrapCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "bootstrap",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"boot"},
		Short:   "Starts bootstrap node",
		Long:    ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := opts.LoadConfigFiles(&opts.BootstrapConfig); err != nil {
				return fmt.Errorf(`config error: %w`, err)
			}
			l, err := opts.BootstrapConfig.Logger.Logger(logger.Dependencies{
				BaseLogger: opts.Logger(),
			})
			if err != nil {
				return fmt.Errorf(`ethereum config error: %w`, err)
			}
			t, err := opts.BootstrapConfig.Transport.LibP2PBootstrap(transport.BootstrapDependencies{
				Logger: l,
			})
			if err != nil {
				return fmt.Errorf(`transport config error: %w`, err)
			}
			if _, ok := t.(*libp2p.P2P); !ok {
				return errors.New("spire-bootstrap works only with the libp2p transport")
			}

			s := supervisor.New(l)
			s.Watch(t, sysmon.New(time.Minute, l))
			if l, ok := l.(supervisor.Service); ok {
				s.Watch(l)
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			if err := s.Start(ctx); err != nil {
				return err
			}
			return <-s.Wait()
		},
	}
}
