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
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func NewRunCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"agent"},
		Short:   "Start the agent",
		Long:    `Start the agent`,
		RunE: func(_ *cobra.Command, _ []string) error {
			srv, err := PrepareService(context.Background(), opts)
			if err != nil {
				return err
			}
			if err = srv.Start(); err != nil {
				return err
			}
			defer srv.CancelAndWait()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}
