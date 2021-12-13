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

	"github.com/makerdao/oracle-suite/internal/config/spire"
	"github.com/makerdao/oracle-suite/pkg/pflag"
)

type options struct {
	pflag.LoggerFlag
	ConfigFilePath    string
	Config            Config
	Version           string
	TransportOverride string
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "spire",
		Version:       opts.Version,
		Short:         "",
		Long:          "",
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(pflag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./config.json",
		"spire config file",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.TransportOverride,
		"transport",
		"",
		"`"+spire.TransportLibP2P+"|"+spire.TransportLibSSB+"` transport spire connects to",
	)

	rootCmd.AddCommand(
		NewAgentCmd(opts),
		NewPullCmd(opts),
		NewPushCmd(opts),
	)

	return rootCmd
}