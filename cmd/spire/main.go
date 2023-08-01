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
	"os"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
)

type options struct {
	cmd.LoggerFlags
	cmd.FilesFlags
	Config            spire.Config
	BootstrapConfig   BootstrapConfig
	TransportOverride string
}

func main() {
	var opts options
	rootCmd := cmd.NewRootCommand(
		"spire",
		cmd.Version,
		cmd.NewLoggerFlagSet(&opts.LoggerFlags),
		cmd.NewFilesFlagSet(&opts.FilesFlags),
	)
	rootCmd.AddCommand(
		cmd.NewRunCmd(
			&opts.Config,
			&opts.FilesFlags,
			&opts.LoggerFlags,
		),
		NewStreamCmd(&opts),
		NewPullCmd(&opts),
		NewPushCmd(&opts),
		NewBootstrapCmd(&opts),
	)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
