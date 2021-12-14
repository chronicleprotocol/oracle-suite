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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewPriceCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price",
		Args:  cobra.ExactArgs(0),
		Short: "commands related to the price messages",
		Long:  ``,
	}

	cmd.AddCommand(
		NewPriceSignCmd(opts),
		NewPriceVerifyCmd(),
	)

	return cmd
}

func NewPriceSignCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "sign [json_message]",
		Args:  cobra.MaximumNArgs(1),
		Short: "signs given JSON price message and returns JSON with VRS fields",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			srv, err := PrepareServices(opts)
			if err != nil {
				return err
			}

			// Read JSON and parse it:
			input, err := readInput(args, 0)
			if err != nil {
				return err
			}
			msg := &messages.Price{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}

			// Sign price:
			err = msg.Price.Sign(srv.Signer)
			if err != nil {
				return err
			}

			// Marshall to JSON:
			signedMsg, err := msg.Marshall()
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", string(signedMsg))

			return nil
		},
	}
}

func NewPriceVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [json_message]",
		Args:  cobra.MaximumNArgs(1),
		Short: "verifies given JSON price message",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			signer := geth.NewSigner(nil)

			// Read JSON and parse it:
			input, err := readInput(args, 0)
			if err != nil {
				return err
			}
			msg := &messages.Price{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}

			// Print message parameters:
			fields := msg.Price.Fields(signer)
			for _, k := range sortFields(fields) {
				fmt.Printf("%-4s %s\n", k, fields[k])
			}

			return nil
		},
	}
}
