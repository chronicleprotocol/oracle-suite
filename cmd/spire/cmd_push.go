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
	"io"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewPushCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
	}

	cmd.AddCommand(NewPushPriceCmd(opts))

	return cmd
}

func NewPushPriceCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.MaximumNArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) (err error) {
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			sup, cli, err := PrepareClientServices(ctx, opts)
			if err != nil {
				return err
			}
			if err = sup.Start(); err != nil {
				return err
			}
			defer func() {
				ctxCancel()
				if sErr := <-sup.Wait(); err == nil { // Ignore sErr if another error has already occurred.
					err = sErr
				}
			}()
			in := os.Stdin
			if len(args) == 1 {
				in, err = os.Open(args[0])
				if err != nil {
					return err
				}
			}
			// Read JSON and parse it:
			input, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			msg := &messages.Price{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}
			// Send price message to RPC client:
			err = cli.PublishPrice(msg)
			if err != nil {
				return err
			}
			return
		},
	}
}
