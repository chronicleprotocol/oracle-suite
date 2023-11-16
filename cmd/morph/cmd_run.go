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
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

func NewRunCmd(cfg supervisor.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Args:    cobra.NoArgs,
		Short:   "Run the main service",
		Aliases: []string{"agent", "server"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			// todo, add flag for rpc url, flag for config registry address
			s, err := cfg.Services(lf.Logger(), cmd.Root().Use, cmd.Root().Version)
			if err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer ctxCancel()
			if err = s.Start(ctx); err != nil {
				return err
			}
			return <-s.Wait()
		},
	}
	flags := cmd.Flags()
	flags.AddFlagSet(cf.FlagSet())
	flags.AddFlagSet(lf.FlagSet())
	return cmd
}
