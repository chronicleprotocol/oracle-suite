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

package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

func Command(name string, config supervisor.Config) *cobra.Command {
	var ConfigFiles FilesFlags
	var LoggerFlags LoggerFlags
	cmd := NewRootCommand(
		name,
		Version,
		NewFilesFlagSet(&ConfigFiles),
		NewLoggerFlagSet(&LoggerFlags),
	)
	cmd.AddCommand(
		NewRunCmd(
			config,
			&ConfigFiles,
			&LoggerFlags,
		),
	)
	return cmd
}

func NewRootCommand(name, version string, sets ...*pflag.FlagSet) *cobra.Command {
	c := &cobra.Command{
		Use:          name,
		Version:      version,
		SilenceUsage: true,
	}
	flags := c.PersistentFlags()
	for _, set := range sets {
		flags.AddFlagSet(set)
	}
	return c
}

func NewRunCmd(c supervisor.Config, f *FilesFlags, l *LoggerFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.NoArgs,
		Aliases: []string{"agent", "server"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := f.Load(c); err != nil {
				return err
			}
			services, err := c.Services(l.Logger())
			if err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			if err = services.Start(ctx); err != nil {
				return err
			}
			return <-services.Wait()
		},
	}
}
