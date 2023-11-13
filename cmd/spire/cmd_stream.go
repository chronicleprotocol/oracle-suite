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

func NewStreamCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:   "stream [TOPIC...]",
		Args:  cobra.MinimumNArgs(0),
		Short: "Streams data from the network",
		RunE: func(cmd *cobra.Command, topics []string) (err error) {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			logger := lf.Logger()
			if len(topics) == 0 {
				topics = messages.AllMessagesMap.Keys()
			}
			services, err := cfg.StreamServices(logger, cmd.Root().Use, cmd.Root().Version, topics...)
			if err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				ctxCancel()
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
			sinkCh := sink.Chan()
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg, ok := <-sinkCh:
					if !ok {
						return nil
					}
					if raw {
						m := mm{
							Meta: msg.Meta,
							Data: msg.Message,
						}
						jsonMsg, err := json.Marshal(m)
						if err != nil {
							lf.Logger().WithError(err).Error("Failed to marshal message")
							continue
						}
						fmt.Println(string(jsonMsg))
						continue
					}
					jsonMsg, err := json.Marshal(handleMessage(msg))
					if err != nil {
						lf.Logger().WithError(err).Error("Failed to marshal message")
						continue
					}
					fmt.Println(string(jsonMsg))
				}
			}
		},
	}
	cmd.AddCommand(
		NewStreamPricesCmd(cfg, cf, lf),
		NewTopicsCmd(),
	)
	cmd.Flags().BoolVar(
		&raw,
		"raw",
		false,
		"show raw messages",
	)
	var format string
	cmd.Flags().StringVarP(&format, "output", "o", "", "(here for backward compatibility)")
	return cmd
}

func NewTopicsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "topics",
		Args:  cobra.ExactArgs(0),
		Short: "List all available topics",
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, topic := range messages.AllMessagesMap.Keys() {
				fmt.Println(topic)
			}
			return nil
		},
	}
}

func NewStreamPricesCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	var legacy bool
	cmd := &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "Prints price messages as they are received",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			topic := messages.DataPointV1MessageName
			if legacy {
				topic = messages.PriceV1MessageName //nolint:staticcheck
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := cfg.StreamServices(lf.Logger(), cmd.Root().Use, cmd.Root().Version, topic)
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				ctxCancel()
				if sErr := <-services.Wait(); err == nil {
					err = sErr
				}
			}()
			msgCh := services.Transport.Messages(topic)
			for {
				select {
				case <-ctx.Done():
					return err
				case msg, ok := <-msgCh:
					if !ok {
						return err
					}
					jsonMsg, err := json.Marshal(msg.Message)
					if err != nil {
						lf.Logger().WithError(err).Error("Failed to marshal message")
						continue
					}
					fmt.Println(string(jsonMsg))
				}
			}
		},
	}
	cmd.Flags().BoolVar(
		&legacy,
		"legacy",
		false,
		"legacy mode",
	)
	return cmd
}
