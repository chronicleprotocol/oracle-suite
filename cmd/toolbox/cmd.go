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
	cobraPkg "github.com/chronicleprotocol/oracle-suite/pkg/cobra"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

type options struct {
	flag.LoggerFlag
	ConfigFilePath string
	Config         Config
}

func NewRootCommand() *cobra.Command {
	var opts options

	rootCmd := &cobra.Command{
		Use:           "toolbox",
		Version:       suite.Version,
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./config.hcl",
		"spire config file",
	)

	rootCmd.AddCommand(
		NewMedianCmd(&opts),
		NewPriceCmd(&opts),
		NewSignerCmd(&opts),
		cobraPkg.NewConfigCmd[Config](opts.ConfigFilePath),
	)

	return rootCmd
}
