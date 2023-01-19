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
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/median/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

func NewMedianCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "median",
		Version: opts.Version,
		Args:    cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			err := config.ParseFile(&opts.Config, opts.ConfigFilePath)
			if err != nil {
				return fmt.Errorf(`config error: %w`, err)
			}

			log, err := opts.Config.Logger.Configure(loggerConfig.Dependencies{
				AppName:    "monitor",
				BaseLogger: opts.Logger(),
			})
			log = log.WithField("tag", LoggerTag)

			if err != nil {
				return fmt.Errorf(`logger config error: %w`, err)
			}

			client, err := opts.Config.Ethereum.ConfigureEthereumClient(nil, log)
			if err != nil {
				return fmt.Errorf(`ethereum config error: %w`, err)
			}

			for _, contract := range opts.Config.Contracts {
				if contract.Address == "" {
					continue
				}
				addr := common.HexToAddress(contract.Address)
				val, err := geth.NewMedian(client, addr).Val(ctx)
				if err != nil {
					log.Error("median", contract.Wat, contract.Symbol)
					continue
				}

				log.
					WithField("val", val.String()).
					WithField("wat", contract.Wat).
					WithField("symbol", contract.Symbol).
					WithField("addr", addr.Hex()).
					Info("current median price")
			}

			// for _, aa := range opts.Config.Transactors {
			// 	if aa == "" {
			// 		continue
			// 	}
			// 	addr := common.HexToAddress(aa)
			// 	val, err := client.Balance(ctx, addr)
			// 	if err != nil {
			// 		log.Errorf("balance", addr)
			// 		continue
			// 	}
			//
			// 	log.
			// 		WithField("balance", val.String()).
			// 		WithField("account", addr.Hex()).
			// 		Info("current account balance")
			// }

			if l, ok := log.(supervisor.Service); ok {
				ctx, cancelFn := context.WithCancel(ctx)
				cancelFn()
				if err := l.Start(ctx); err != nil {
					return err
				}
				return <-l.Wait()
			}

			return nil
		},
	}
}
