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

type FilesFlags struct {
	//TODO: think of ways to make it a pflag.Value interface
	paths []string
}

func NewFilesFlagSet(cfp *FilesFlags) *pflag.FlagSet {
	fs := pflag.NewFlagSet("config", pflag.PanicOnError)
	fs.StringSliceVarP(
		&cfp.paths,
		"config", "c",
		[]string{"./config.hcl"},
		"config file",
	)
	return fs
}

func (cf FilesFlags) LoadConfigFiles(c any) error {
	return config.LoadFiles(c, cf.paths)
}
