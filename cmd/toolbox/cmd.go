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
	"github.com/spf13/cobra"

	suite "github.com/chronicleprotocol/oracle-suite"
	"github.com/chronicleprotocol/oracle-suite/cmd"
)

type options struct {
	cmd.LoggerFlags
	cmd.FilesFlags
	Config Config
}

func NewRootCommand() *cobra.Command {
	var opts options

	rootCmd := &cobra.Command{
		Use:           "toolbox",
		Version:       suite.Version,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(cmd.NewFilesFlagSet(&opts.FilesFlags))

	rootCmd.AddCommand(
		NewMedianCmd(&opts),
		NewPriceCmd(&opts),
		NewSignerCmd(&opts),
	)

	return rootCmd
}
