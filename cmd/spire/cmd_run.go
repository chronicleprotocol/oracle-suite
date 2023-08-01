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
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func NewAgentCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Aliases: []string{"agent", "server"},
		Args:    cobra.ExactArgs(0),
		Short:   "Starts the Spire agent",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := opts.Load(&opts.Config); err != nil {
				return err
			}
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
			s, err := opts.Config.AgentServices(opts.Logger())
			if err != nil {
				return err
			}
			if err := s.Start(ctx); err != nil {
				return err
			}
			return <-s.Wait()
		},
	}
}
