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
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	ghost "github.com/chronicleprotocol/oracle-suite/pkg/config/ghostnext"
	morphConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/morph"
)

func NewRunCmd(c *ghost.Config, f *cmd.ConfigFlags, l *cmd.LoggerFlags) *cobra.Command {
	cc := &cobra.Command{
		Use:     "run",
		Args:    cobra.NoArgs,
		Short:   "Run the main service",
		Aliases: []string{"agent", "server"},
		RunE: func(cc *cobra.Command, _ []string) error {
			if err := f.Load(c); err != nil {
				return err
			}
			services, err := c.Services(l.Logger(), cc.Root().Use, cc.Root().Version)
			if err != nil {
				return err
			}

			var morph morphConfig.Config
			cf := cmd.ConfigFlagsForConfig(morph)
			if err := cf.Load(&morph); err != nil {
				return err
			}
			morphService, err := morph.Configure(morphConfig.Dependencies{
				Clients: services.Clients,
				Logger:  services.Logger,
				Base:    c,
			})
			if err != nil {
				return err
			}

			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			if err = services.Start(ctx); err != nil {
				return err
			}
			if err = morphService.Start(ctx); err != nil {
				return err
			}

			defer func() {
				if sErr := <-services.Wait(); err == nil {
					err = sErr
				}
			}()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-morphService.Wait():
					ctxCancel()
					return nil
				}
			}
		},
	}
	flags := cc.Flags()
	flags.AddFlagSet(f.FlagSet())
	flags.AddFlagSet(l.FlagSet())
	return cc
}
