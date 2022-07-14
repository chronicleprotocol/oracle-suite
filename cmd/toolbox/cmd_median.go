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
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	oracleGeth "github.com/chronicleprotocol/oracle-suite/pkg/price/oracle/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewMedianCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "median",
		Args:  cobra.ExactArgs(1),
		Short: "commands related to the Medianizer contract",
		Long:  ``,
	}

	cmd.AddCommand(
		NewMedianAgeCmd(opts),
		NewMedianBarCmd(opts),
		NewMedianWatCmd(opts),
		NewMedianValCmd(opts),
		NewMedianFeedsCmd(opts),
		NewMedianPokeCmd(opts),
		NewMedianLiftCmd(opts),
		NewMedianDropCmd(opts),
		NewMedianSetBarCmd(opts),
	)

	return cmd
}

func NewMedianAgeCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "age median_address",
		Args:  cobra.ExactArgs(1),
		Short: "returns the age value (last update time)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			age, err := median.Age(context.Background())
			if err != nil {
				return err
			}

			// Print age:
			fmt.Println(age.String())

			return nil
		},
	}
}

func NewMedianBarCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "bar median_address",
		Args:  cobra.ExactArgs(1),
		Short: "returns the bar value (required quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			bar, err := median.Bar(context.Background())
			if err != nil {
				return err
			}

			fmt.Println(bar)

			return nil
		},
	}
}

func NewMedianWatCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "wat median_address",
		Args:  cobra.ExactArgs(1),
		Short: "returns the wat value (asset name)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			wat, err := median.Wat(context.Background())
			if err != nil {
				return err
			}

			fmt.Println(wat)

			return nil
		},
	}
}

func NewMedianValCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "val median_address",
		Args:  cobra.ExactArgs(1),
		Short: "returns the val value (asset price)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			price, err := median.Val(context.Background())
			if err != nil {
				return err
			}

			fmt.Println(price.String())

			return nil
		},
	}
}

func NewMedianFeedsCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "feeds median_address",
		Args:  cobra.ExactArgs(1),
		Short: "returns list of feeds which are allowed to send prices",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			feeds, err := median.Feeds(context.Background())
			if err != nil {
				return err
			}

			for _, f := range feeds {
				fmt.Println(f.String())
			}

			return nil
		},
	}
}

func NewMedianPokeCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "poke median_address [json_messages_list]",
		Args:  cobra.ExactArgs(1),
		Short: "directly invokes poke method",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[1]))

			// Read JSON and parse it:
			in, err := readInput(args, 1)
			if err != nil {
				return err
			}

			msgs := &[]messages.Price{}
			err = json.Unmarshal(in, msgs)
			if err != nil {
				return err
			}

			var prices []*oracle.Price
			for _, m := range *msgs {
				prices = append(prices, m.Price)
			}

			tx, err := median.Poke(context.Background(), prices, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianLiftCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "lift median_address [addresses...]",
		Args:  cobra.MinimumNArgs(2),
		Short: "adds given addresses to the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			var addresses []ethereum.Address
			for _, a := range args[1:] {
				addresses = append(addresses, ethereum.HexToAddress(a))
			}

			tx, err := median.Lift(context.Background(), addresses, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianDropCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "drop median_address [addresses...]",
		Args:  cobra.MinimumNArgs(2),
		Short: "removes given addresses from the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			geth, _, err := opts.Config.Configure()
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(geth, ethereum.HexToAddress(args[1]))

			var addresses []ethereum.Address
			for _, a := range args[1:] {
				addresses = append(addresses, ethereum.HexToAddress(a))
			}

			tx, err := median.Drop(context.Background(), addresses, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianSetBarCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "set-bar median_address bar",
		Args:  cobra.ExactArgs(2),
		Short: "sets bar variable (quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}
			median := oracleGeth.NewMedian(srv.Client, ethereum.HexToAddress(args[0]))

			bar, ok := (&big.Int{}).SetString(args[1], 10)
			if !ok {
				return errors.New("given value is not an valid number")
			}

			tx, err := median.SetBar(context.Background(), bar, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}
