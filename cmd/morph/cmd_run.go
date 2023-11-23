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
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/morph"
)

type morphOptions struct {
	bin            string
	args           string
	rpcURL         string
	configRegistry string
}

var options *morphOptions

func NewRunCmd(cfg *morph.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	options = &morphOptions{}
	cmd := &cobra.Command{
		Use:     "run",
		Args:    validateArgs,
		Short:   "Run the main service",
		Aliases: []string{"agent", "server"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Prioritize the arguments over embedded config
			if len(options.rpcURL) > 0 {
				os.Setenv("CFG_CHAIN_RPC_URLS", options.rpcURL)
			}
			if len(options.bin) > 0 {
				os.Setenv("CFG_APP_BIN", options.bin)
				os.Setenv("CFG_APP_ARGS", options.args)
			}
			if len(options.configRegistry) > 0 {
				os.Setenv("CFG_CONFIG_REGISTRY", options.configRegistry)
			}

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
	cmd.PersistentFlags().StringVar(
		&options.bin,
		"bin",
		"",
		"Path to executable binary",
	)
	cmd.PersistentFlags().StringVar(
		&options.args,
		"args",
		"",
		"Arguments to binary including command",
	)
	cmd.PersistentFlags().StringVar(
		&options.rpcURL,
		"config-rpc",
		"",
		"RPC URL",
	)
	cmd.PersistentFlags().StringVar(
		&options.configRegistry,
		"config-registry",
		"",
		"Address to ConfigRegistry Smart Contract",
	)
	return cmd
}

func validateArgs(_ *cobra.Command, _ []string) error {
	if options == nil { // never happen
		return fmt.Errorf("invalid function call")
	}
	if len(options.rpcURL) > 0 {
		if len(options.rpcURL) < 7 || (!strings.HasPrefix(options.rpcURL, "https://") && !strings.HasPrefix(options.rpcURL, "http://")) {
			return fmt.Errorf("--config-rpc should start with https://")
		}
	}
	if len(options.configRegistry) > 0 {
		if _, err := types.AddressFromHex(options.configRegistry); err != nil {
			return fmt.Errorf("--config-registry should be valid address")
		}
	}
	if len(options.bin) > 0 && len(options.args) == 0 {
		// --args should include at least command if --bin was given
		// i.e. --bin ./ghost --args run
		// if --args has additional values, should pass inside the quotation mark
		// i.e. --bin ./ghost --args "run --raw"
		return fmt.Errorf("--args should include command at least")
	}
	return nil
}
