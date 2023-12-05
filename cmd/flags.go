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
	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
)

func ConfigFlagsForConfig(d config.HasDefaults) ConfigFlags {
	return ConfigFlagsWithEmbeds(d.DefaultEmbeds()...)
}

func ConfigFlagsWithEmbeds(embeds ...[]byte) ConfigFlags {
	return ConfigFlags{
		embeds: embeds,
	}
}

// ConfigFlags is used to load multiple config files.
type ConfigFlags struct {
	paths  []string
	embeds [][]byte
}

// Load loads the config files into the given config struct.
func (cf *ConfigFlags) Load(cfg any) error {
	if len(cf.paths) == 0 {
		if err := config.LoadEmbeds(cfg, cf.embeds); err != nil {
			return err
		}
	} else {
		if err := config.LoadFiles(cfg, cf.paths); err != nil {
			return err
		}
	}
	return nil
}

// FlagSet binds CLI args [--config or -c] for config files as a pflag.FlagSet.
func (cf *ConfigFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("config", pflag.PanicOnError)
	fs.StringSliceVarP(
		&cf.paths,
		"config",
		"c",
		[]string{},
		"config file",
	)
	return fs
}
