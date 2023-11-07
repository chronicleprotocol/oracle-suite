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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/globals"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/hcl"
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
func (ff *ConfigFlags) Load(c any) (bool, error) {
	if len(ff.paths) == 0 {
		if err := config.LoadEmbeds(c, ff.embeds); err != nil {
			return false, err
		}
	} else {
		if err := config.LoadFiles(c, ff.paths); err != nil {
			return false, err
		}
	}
	switch {
	case globals.ShowEnvVarsUsedInConfig:
		for _, v := range globals.EnvVars {
			fmt.Println(v)
		}
		return true, nil
	case globals.RenderConfigJSON:
		marshaled, err := json.Marshal(c)
		if err != nil {
			return true, err
		}
		fmt.Println(string(marshaled))
		return true, nil
	case globals.RenderConfigHCL:
		f := hclwrite.NewFile()
		body := f.Body()
		err := hcl.Encode(c, body)
		if err != nil {
			return true, err
		}
		fmt.Println(string(f.Bytes()))
		return true, nil
	}
	return false, nil
}

// FlagSet binds CLI args [--config or -c] for config files as a pflag.FlagSet.
func (ff *ConfigFlags) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("config", pflag.PanicOnError)
	fs.StringSliceVarP(
		&ff.paths,
		"config",
		"c",
		[]string{},
		"config file",
	)
	fs.BoolVar(
		&globals.ShowEnvVarsUsedInConfig,
		"config.env",
		false,
		"show environment variables used in config files and exit",
	)
	fs.BoolVar(
		&globals.RenderConfigJSON,
		"config.json",
		false,
		"render config as JSON and exit",
	)
	fs.BoolVar(
		&globals.RenderConfigHCL,
		"config.hcl",
		false,
		"render config as HCL and exit",
	)
	return fs
}
