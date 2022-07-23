//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gofer",
		Version: opts.Version,
		Short:   "Tool for providing reliable data in the blockchain ecosystem",
		Long: `
Gofer is a tool that provides reliable asset prices taken from various sources.

It is a tool that allows for easy data retrieval from various sources
with aggregates that increase reliability in the DeFi environment.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./config.json",
		"config file",
	)
	rootCmd.PersistentFlags().VarP(
		&opts.Format,
		"format",
		"f",
		"output format",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.NoRPC,
		"norpc",
		false,
		"disable the use of RPC agent",
	)

	return rootCmd
}
