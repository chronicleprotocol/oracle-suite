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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	oracleGeth "github.com/chronicleprotocol/oracle-suite/pkg/price/oracle/geth"
)

func NewSpectreCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spectre",
		Args:  cobra.ExactArgs(1),
		Short: "commands related to the Spectre app",
		Long:  ``,
	}

	cmd.AddCommand(
		NewSpectreMedianCmd(opts),
	)

	return cmd
}

func NewSpectreMedianCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "median [pairs...]",
		Args:  cobra.MinimumNArgs(0),
		Short: "returns information about medianizers defined in the spectre config",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				for name := range opts.Config.Medianizers() {
					args = append(args, name)
				}
			}
			for _, name := range args {
				config, ok := opts.Config.Medianizers()[name]
				if !ok {
					fmt.Printf("%s is missing in the Spectre config\n\n", name)
					continue
				}

				median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(config.Contract))
				ctx := context.Background()
				wat, err := median.Wat(ctx)
				if err != nil {
					return err
				}
				bar, err := median.Bar(ctx)
				if err != nil {
					return err
				}
				age, err := median.Age(ctx)
				if err != nil {
					return err
				}
				val, err := median.Val(ctx)
				if err != nil {
					return err
				}
				feeds, err := median.Feeds(ctx)
				if err != nil {
					return err
				}

				fmt.Println(name)
				fmt.Printf("Wat: %s\n", wat)
				fmt.Printf("Bar: %d\n", bar)
				fmt.Printf("Age: %s\n", age.String())
				fmt.Printf("Val: %s\n", val.String())
				fmt.Print("Feeds:\n")
				for _, f := range feeds {
					fmt.Println(f.String())
				}
				fmt.Print("\n")
			}

			return nil
		},
	}
}
