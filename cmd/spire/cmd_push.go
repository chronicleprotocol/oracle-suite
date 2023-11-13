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

package main

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewPushCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Args:  cobra.ExactArgs(1),
		Short: "Push a message to the network (require agent)",
	}
	cmd.AddCommand(
		NewPushPriceCmd(cfg, cf, lf),
	)
	return cmd
}

func NewPushPriceCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.MaximumNArgs(1),
		Short: "Push a data point message to the network",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := cfg.ClientServices(lf.Logger(), cmd.Root().Use, cmd.Root().Version)
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				ctxCancel()
				if sErr := <-services.Wait(); err == nil { // Ignore sErr if another error has already occurred.
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
			msg := &messages.DataPoint{}
			err = msg.Unmarshall(input)
			if err != nil {
				return err
			}
			// Send price message to RPC client:
			err = services.SpireClient.Publish(msg)
			if err != nil {
				return err
			}
			return
		},
	}
}
