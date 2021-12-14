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

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

type options struct {
	Listen     string
	EnableCORS bool
	EthRPCURLs []string
	flag.LoggerFlag
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "rpc-splitter",
		Version:       suite.Version,
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringVarP(
		&opts.Listen,
		"listen",
		"l",
		"127.0.0.1:8545",
		"listen address",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&opts.EnableCORS,
		"enable-cors",
		"c",
		false,
		"enables CORS requests for all origins",
	)
	rootCmd.PersistentFlags().StringSliceVar(
		&opts.EthRPCURLs,
		"eth-rpc",
		[]string{},
		"list of ethereum nodes",
	)
	err := rootCmd.MarkPersistentFlagRequired("eth-rpc")
	if err != nil {
		panic(err)
	}
	return rootCmd
}
