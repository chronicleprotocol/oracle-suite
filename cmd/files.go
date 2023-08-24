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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/globals"
)

// FilesFlags is used to load multiple config files.
type FilesFlags struct {
	paths []string
}

// Load loads the config files into the given config struct.
func (ff *FilesFlags) Load(c any) error {
	if err := config.LoadFiles(c, ff.paths); err != nil {
		return err
	}
	if globals.ShowEnvVarsUsedInConfig {
		for _, v := range globals.EnvVars {
			fmt.Println(v)
		}
		os.Exit(0)
	}
	return nil
}

// FlagSet binds CLI args [--config or -c] for config files as a pflag.FlagSet.
func (ff *FilesFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("config", pflag.PanicOnError)
	fs.StringSliceVarP(
		&ff.paths,
		"config",
		"c",
		[]string{"./config.hcl"},
		"config file",
	)
	fs.BoolVar(
		&globals.ShowEnvVarsUsedInConfig,
		"config.env",
		false,
		"show environment variables used in config files",
	)
	return fs
}
