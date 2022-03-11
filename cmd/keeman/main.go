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

	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/cobra"
)

func main() {
	opts, cmd := cobra.Command()
	cmd.PersistentFlags().StringVarP(
		&opts.InputFile,
		"input",
		"i",
		"",
		"input file path",
	)
	cmd.PersistentFlags().StringVarP(
		&opts.OutputFile,
		"output",
		"o",
		"",
		"output file path",
	)
	cmd.PersistentFlags().BoolVarP(
		&opts.Verbose,
		"verbose",
		"v",
		false,
		"verbose logging",
	)
	cmd.AddCommand(
		cobra.NewDerive(opts),
		cobra.NewDeriveTf(),
		cobra.GenerateSeed(opts),
		cobra.NewList(opts),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
