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
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver"
	"github.com/chronicleprotocol/oracle-suite/pkg/httpserver/middleware"
	"github.com/chronicleprotocol/oracle-suite/pkg/rpcsplitter"
)

func NewRunCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"agent"},
		Short:   "Start server",
		Long:    `Start server`,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			log := opts.Logger()
			var server, err = rpcsplitter.NewServer(
				rpcsplitter.WithEndpoints(opts.EthRPCURLs),
				rpcsplitter.WithTotalTimeout(time.Duration(opts.TotalTimeoutSec)*time.Second),
				rpcsplitter.WithGracefulTimeout(time.Duration(opts.GracefulTimeoutSec)*time.Second),
				rpcsplitter.WithRequirements(minimumRequiredResponses(len(opts.EthRPCURLs)), opts.MaxBlocksBehind),
				rpcsplitter.WithLogger(opts.Logger()),
			)
			if err != nil {
				return err
			}

			srv := httpserver.New(&http.Server{
				Addr:    opts.Listen,
				Handler: server,
			})

			srv.Use(&middleware.Recover{
				Recover: func(err interface{}) {
					log.WithField("panic", fmt.Sprintf("%s", err)).Error("Server handler crashed")
				},
			})

			srv.Use(&middleware.Logger{Log: log})

			if opts.EnableCORS {
				srv.Use(&middleware.CORS{
					Origin:  func(r *http.Request) string { return r.Header.Get("Origin") },
					Headers: func(*http.Request) string { return "Content-Type" },
					Methods: func(*http.Request) string { return "POST" },
				})
			}

			err = srv.Start(ctx)
			if err != nil {
				return fmt.Errorf("unable to start the HTTP server: %w", err)
			}

			defer func() {
				err := <-srv.Wait()
				if err != nil {
					log.WithError(err).Error("Error while closing HTTP server")
				}
			}()

			return <-srv.Wait()
		},
	}
}

func minimumRequiredResponses(endpoints int) int {
	if endpoints < 2 {
		return endpoints
	}
	return endpoints - 1
}
