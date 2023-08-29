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
	"github.com/chronicleprotocol/oracle-suite/pkg/util/chanutil"
)

func NewStreamCmd(c *spire.Config, f *cmd.FilesFlags, l *cmd.LoggerFlags) *cobra.Command {
	var raw bool
	cc := &cobra.Command{
		Use:   "stream [TOPIC...]",
		Args:  cobra.MinimumNArgs(0),
		Short: "Streams data from the network",
		RunE: func(cc *cobra.Command, topics []string) (err error) {
			if err := f.Load(c); err != nil {
				return err
			}
			logger := l.Logger()
			if len(topics) == 0 {
				topics = transport.AllMessagesMap.Keys()
			}
			services, err := c.StreamServices(logger, cc.Root().Use, cc.Root().Version, topics...)
			if err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				if sErr := <-services.Wait(); err == nil {
					err = sErr
				}
			}()
			sink := chanutil.NewFanIn[transport.ReceivedMessage]()
			for _, s := range topics {
				ch := services.Transport.Messages(s)
				if ch == nil {
					return fmt.Errorf("unconfigured topic: %s", s)
				}
				if err := sink.Add(ch); err != nil {
					return err
				}
				logger.
					WithField("name", s).
					Info("Subscribed to topic")
			}
			type mm struct {
				Data any            `json:"data"`
				Meta transport.Meta `json:"meta"`
			}
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-sink.Chan():
					if raw {
						jsonMsg, err := json.Marshal(msg.Message)
						if err != nil {
							return err
						}
						fmt.Println(string(jsonMsg))
						continue
					}
					m := mm{
						Meta: msg.Meta,
						Data: msg.Message,
					}
					jsonMsg, err := json.Marshal(m)
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
	cc.Flags().BoolVar(
		&raw,
		"raw",
		false,
		"show only data without meta",
	)
	return cc
}

func NewTopicsCmd() *cobra.Command {
	cc := &cobra.Command{
		Use:   "topics",
		Args:  cobra.ExactArgs(0),
		Short: "List all available topics",
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, topic := range transport.AllMessagesMap.Keys() {
				fmt.Println(topic)
			}
			return nil
		},
	}
	return cc
}

func NewStreamPricesCmd(c *spire.Config, f *cmd.FilesFlags, l *cmd.LoggerFlags) *cobra.Command {
	var legacy bool
	cc := &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "Prints price messages as they are received",
		RunE: func(cc *cobra.Command, _ []string) (err error) {
			if err := f.Load(c); err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := c.StreamServices(l.Logger(), cc.Root().Use, cc.Root().Version)
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
