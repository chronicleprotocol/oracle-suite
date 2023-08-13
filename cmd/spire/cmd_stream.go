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
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewStreamCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Args:  cobra.ExactArgs(1),
		Short: "Streams data from the network",
	}

	cmd.AddCommand(
		NewStreamPricesCmd(opts),
	)

	return cmd
}

func NewStreamPricesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "Prints price messages as they are received",
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			if err := opts.Load(&opts.Config); err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := opts.Config.StreamServices(opts.Logger())
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				if sErr := <-services.Wait(); err == nil {
					err = sErr
				}
			}()
			msgCh := services.Transport.Messages(messages.DataPointV1MessageName)
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-msgCh:
					jsonMsg, err := json.Marshal(msg.Message)
					if err != nil {
						return err
					}
					fmt.Println(string(jsonMsg))
				}
			}
		},
	}
}
