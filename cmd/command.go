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
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
)

// NewRootCommand returns a Cobra command with the given name and version.
// It also adds all the provided pflag.FlagSet items to the command's persistent flags.
func NewRootCommand(name, version string, sets ...FlagSetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Version:      version,
		SilenceUsage: true,
	}
	flags := cmd.PersistentFlags()
	for _, set := range sets {
		flags.AddFlagSet(set.FlagSet())
	}
	return cmd
}

func NewRunCmd(cfg supervisor.Config, cf *ConfigFlags, lf *LoggerFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Args:    cobra.NoArgs,
		Short:   "Run the main service",
		Aliases: []string{"agent", "server"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cf.Load(cfg); err != nil {
				return err
			}
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

func NewRenderConfigCmd(cfg supervisor.Config, cf *ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Args:  cobra.NoArgs,
		Short: "Render the config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			body := &hcl.Block{}
			if err := hcl.Encode(cfg, body); err.HasErrors() {
				return err
			}
			content, diags := body.Bytes()
			if diags.HasErrors() {
				return diags
			}
			fmt.Println(string(content))
			return nil
		},
	}
	flags := cmd.Flags()
	flags.AddFlagSet(cf.FlagSet())
	return cmd
}

type FlagSetter interface {
	FlagSet() *pflag.FlagSet
}
