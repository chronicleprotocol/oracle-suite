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
	"os"

	"github.com/spf13/cobra"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

const LoggerTag = "MONITOR"

func main() {
	opts := options{Version: suite.Version}
	rootCmd := NewRootCommand(&opts)
	rootCmd.AddCommand(
		NewMedianCmd(&opts),
		NewBalanceCmd(&opts),
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type options struct {
	flag.LoggerFlag
	ConfigFilePath string
	Config         Config
	Version        string
}

type Config struct {
	Ethereum  ethereum.Ethereum `json:"ethereum"`
	Logger    logger.Logger     `json:"logger"`
	Contracts []struct {
		Address string `json:"address"`
		Symbol  string `json:"symbol"`
		Wat     string `json:"wat"`
	} `json:"contracts"`
	Transactors []string `json:"transactors"`
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "monitor",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: opts.Version,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./monitor.json",
		"monitor config file",
	)

	return rootCmd
}
