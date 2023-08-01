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

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"
)

// exitCode to be returned by the application.
var exitCode = 0

func main() {
	opts := options{
		Format: formatTypeValue{format: marshal.NDJSON},
	}
	rootCmd := cmd.NewRootCommand(
		"gofer",
		cmd.Version,
		cmd.NewLoggerFlagSet(&opts.LoggerFlags),
		cmd.NewFilesFlagSet(&opts.FilesFlags),
	)
	rootCmd.PersistentFlags().VarP(
		&opts.Format,
		"format",
		"o",
		"output format",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.NoRPC,
		"norpc",
		false,
		"disable the use of RPC agent",
	)
	rootCmd.AddCommand(
		cmd.NewRunCmd(
			&opts.Config,
			&opts.FilesFlags,
			&opts.LoggerFlags,
		),
		NewPairsCmd(&opts),
		NewPricesCmd(&opts),
		NewModelsCmd(&opts),
		NewDataCmd(&opts),
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n", err)
		if exitCode == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exitCode)
}

func getModelsNames(ctx context.Context, provider datapoint.Provider, args []string) []string {
	if len(args) == 0 {
		return provider.ModelNames(ctx)
	}
	return args
}
