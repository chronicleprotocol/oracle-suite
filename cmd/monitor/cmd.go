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
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	loggerConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/logger"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
)

type options struct {
	flag.LoggerFlag
	ConfigFilePath string
	Config         Config
	Version        string
}

type Config struct {
	Ethereum  ethereumConfig.Ethereum `json:"ethereum"`
	Logger    loggerConfig.Logger     `json:"logger"`
	Contracts []struct {
		Address string `json:"address"`
		Symbol  string `json:"symbol"`
		Wat     string `json:"wat"`
	} `json:"contracts"`
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "monitor",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: opts.Version,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./monitor.json",
		"monitor config file",
	)

	return rootCmd
}

const LoggerTag = "MONITOR"

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
					log.Errorf(contract.Wat, contract.Symbol)
					continue
				}

				log.
					WithField("val", val.String()).
					WithField("wat", contract.Wat).
					WithField("symbol", contract.Symbol).
					WithField("addr", addr.Hex()).
					Info("current median price")
			}

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
