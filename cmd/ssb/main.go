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

	"github.com/chronicleprotocol/oracle-suite/cmd/ssb/cobra"
)

func main() {
	opts, cmd := cobra.Root()
	cmd.PersistentFlags().BoolVarP(
		&opts.Verbose,
		"verbose",
		"v",
		false,
		"verbose logging",
	)
	cmd.PersistentFlags().StringVarP(
		&opts.CapsPath,
		"caps",
		"c",
		"./local.caps.json",
		"caps file path",
	)
	cmd.PersistentFlags().StringVarP(
		&opts.KeysPath,
		"keys",
		"k",
		"./local.ssb.json",
		"caps file path",
	)
	cmd.PersistentFlags().StringVarP(
		&opts.SsbHost,
		"host",
		"h",
		"127.0.0.1",
		"ssb server host",
	)
	cmd.PersistentFlags().IntVarP(
		&opts.SsbPort,
		"port",
		"p",
		8008,
		"ssb server port",
	)
	cmd.AddCommand(
		cobra.Publish(opts),
		cobra.Pull(opts),
		cobra.Log(opts),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
