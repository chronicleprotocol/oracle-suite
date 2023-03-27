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

	"github.com/defiweb/go-eth/types"
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	medianGeth "github.com/chronicleprotocol/oracle-suite/pkg/price/median/geth"

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
		Use:   "age client median_address",
		Args:  cobra.ExactArgs(2),
		Short: "returns the age value (last update time)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			age, err := med.Age(context.Background())
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
		Use:   "bar client median_address",
		Args:  cobra.ExactArgs(2),
		Short: "returns the bar value (required quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			bar, err := med.Bar(context.Background())
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
		Use:   "wat client median_address",
		Args:  cobra.ExactArgs(2),
		Short: "returns the wat value (asset name)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			wat, err := med.Wat(context.Background())
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
		Use:   "val client median_address",
		Args:  cobra.ExactArgs(2),
		Short: "returns the val value (asset price)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			price, err := med.Val(context.Background())
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
		Use:   "feeds client median_address",
		Args:  cobra.ExactArgs(2),
		Short: "returns list of feeds which are allowed to send prices",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			feeds, err := med.Feeds(context.Background())
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
		Use:   "poke client median_address [json_messages_list]",
		Args:  cobra.MinimumNArgs(2),
		Short: "directly invokes poke method",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			// Read JSON and parse it:
			in, err := readInput(args, 2)
			if err != nil {
				return err
			}

			msgs := &[]messages.Price{}
			err = json.Unmarshal(in, msgs)
			if err != nil {
				return err
			}

			var prices []*median.Price
			for _, m := range *msgs {
				prices = append(prices, m.Price)
			}

			tx, err := med.Poke(context.Background(), prices, true)
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
		Use:   "lift client median_address [addresses...]",
		Args:  cobra.MinimumNArgs(2),
		Short: "adds given addresses to the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			var addresses []types.Address
			for _, a := range args[2:] {
				addresses = append(addresses, types.MustAddressFromHex(a))
			}

			tx, err := med.Lift(context.Background(), addresses, true)
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
		Use:   "drop client median_address [addresses...]",
		Args:  cobra.MinimumNArgs(3),
		Short: "removes given addresses from the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			var addresses []types.Address
			for _, a := range args[2:] {
				addresses = append(addresses, types.MustAddressFromHex(a))
			}

			tx, err := med.Drop(context.Background(), addresses, true)
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
		Use:   "set-bar client median_address bar",
		Args:  cobra.ExactArgs(3),
		Short: "sets bar variable (quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Client:
			cli, ok := srv.Clients[args[0]]
			if !ok {
				return fmt.Errorf("unable to find client %s", args[0])
			}

			//nolint:staticcheck // ethereum.Client is deprecated
			med := medianGeth.NewMedian(geth.NewClient(cli), types.MustAddressFromHex(args[1]))

			bar, ok := (&big.Int{}).SetString(args[2], 10)
			if !ok {
				return errors.New("given value is not an valid number")
			}

			tx, err := med.SetBar(context.Background(), bar, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}
