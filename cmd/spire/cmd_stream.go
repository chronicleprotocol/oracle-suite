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

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func NewStreamCmd(c *spire.Config, f *cmd.FilesFlags, l *cmd.LoggerFlags) *cobra.Command {
	cc := &cobra.Command{
		Use:   "stream TOPIC",
		Args:  cobra.ExactArgs(1),
		Short: "Streams data from the network",
		RunE: func(_ *cobra.Command, arg []string) (err error) {
			if err := f.Load(c); err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := c.StreamServices(l.Logger())
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
			var msgCh <-chan transport.ReceivedMessage
			switch arg[0] {
			case messages.PriceV0MessageName: //nolint:staticcheck
				msgCh = services.Transport.Messages(messages.PriceV0MessageName) //nolint:staticcheck
			case messages.PriceV1MessageName: //nolint:staticcheck
				msgCh = services.Transport.Messages(messages.PriceV1MessageName) //nolint:staticcheck
			case messages.DataPointV1MessageName:
				msgCh = services.Transport.Messages(messages.DataPointV1MessageName)
			case messages.GreetV1MessageName:
				msgCh = services.Transport.Messages(messages.GreetV1MessageName)
			case messages.EventV1MessageName:
				msgCh = services.Transport.Messages(messages.EventV1MessageName)
			default:
				return fmt.Errorf("unknown message type: %s", arg[0])
			}
			for {
				select {
				case <-ctx.Done():
					return nil
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
	cc.AddCommand(
		NewStreamPricesCmd(c, f, l),
		NewTopicsCmd(),
	)
	return cc
}

func NewTopicsCmd() *cobra.Command {
	var legacy bool
	cc := &cobra.Command{
		Use:   "topics",
		Args:  cobra.ExactArgs(0),
		Short: "List all available topics",
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, topic := range transport.AllTopics {
				fmt.Println(topic)
			}
			return nil
		},
	}
	cc.Flags().BoolVar(
		&legacy,
		"legacy",
		false,
		"legacy mode",
	)
	return cc
}

func NewStreamPricesCmd(c *spire.Config, f *cmd.FilesFlags, l *cmd.LoggerFlags) *cobra.Command {
	var legacy bool
	cc := &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "Prints price messages as they are received",
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			if err := f.Load(c); err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := c.StreamServices(l.Logger())
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
			var msgCh <-chan transport.ReceivedMessage
			if legacy {
				msgCh = services.Transport.Messages(messages.PriceV1MessageName) //nolint:staticcheck
			} else {
				msgCh = services.Transport.Messages(messages.DataPointV1MessageName)
			}
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
	cc.Flags().BoolVar(
		&legacy,
		"legacy",
		false,
		"legacy mode",
	)
	return cc
}
