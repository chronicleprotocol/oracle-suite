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
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
)

func NewPullCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Args:  cobra.ExactArgs(1),
		Short: "Pulls data from the Spire datastore (require agent)",
	}
	cmd.AddCommand(
		NewPullPriceCmd(cfg, cf, lf),
		NewPullPricesCmd(cfg, cf, lf),
	)
	return cmd
}

func NewPullPriceCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "price ASSET_PAIR FEED",
		Args:  cobra.ExactArgs(2),
		Short: "Pulls latest price for a given pair and feed",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			p, err := services.SpireClient.PullPrice(args[0], args[1])
			if err != nil {
				return err
			}
			if p == nil {
				return errors.New("there is no price in the datastore for a given feed and asset pair")
			}
			bts, err := json.Marshal(p)
			if err != nil {
				return err
			}
			_, err = fmt.Printf("%s\n", string(bts))
			return err
		},
	}
}

type pullPricesOptions struct {
	FilterPair string
	FilterFrom string
}

func NewPullPricesCmd(cfg *spire.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	var pullPricesOpts pullPricesOptions
	cmd := &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "Pulls all prices",
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
			p, err := services.SpireClient.PullPrices(pullPricesOpts.FilterPair, pullPricesOpts.FilterFrom)
			if err != nil {
				return err
			}
			bts, err := json.Marshal(p)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(bts))
			return
		},
	}
	cmd.PersistentFlags().StringVar(
		&pullPricesOpts.FilterFrom,
		"filter.from",
		"",
		"",
	)
	cmd.PersistentFlags().StringVar(
		&pullPricesOpts.FilterPair,
		"filter.pair",
		"",
		"",
	)
	return cmd
}
