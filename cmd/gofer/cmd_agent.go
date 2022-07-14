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

	"github.com/spf13/cobra"
)

func NewAgentCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Args:  cobra.NoArgs,
		Short: "Start an RPC server",
		Long:  `Start an RPC server.`,
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			sup, err := PrepareAgentServices(ctx, opts)
			if err != nil {
				return err
			}
			if err = sup.Start(); err != nil {
				return err
			}
			return <-sup.Wait()
		},
	}
}
