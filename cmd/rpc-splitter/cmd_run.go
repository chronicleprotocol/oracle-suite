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
	"syscall"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/internal/httpserver"
	"github.com/chronicleprotocol/oracle-suite/internal/httpserver/middleware"
	"github.com/chronicleprotocol/oracle-suite/internal/rpcsplitter"
)

func NewRunCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"agent"},
		Short:   "Start server",
		Long:    `Start server`,
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, ctxCancel := context.WithCancel(context.Background())
			defer ctxCancel()

			log := opts.Logger()

			handler, err := rpcsplitter.NewHandler(opts.EthRPCURLs, log)
			if err != nil {
				return err
			}

			srv := httpserver.New(ctx, &http.Server{
				Addr:    opts.Listen,
				Handler: handler,
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

			err = srv.Start()
			if err != nil {
				return fmt.Errorf("unable to start the HTTP server: %w", err)
			}

			defer func() {
				err := <-srv.Wait()
				if err != nil {
					log.WithError(err).Error("Error while closing HTTP server")
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}
