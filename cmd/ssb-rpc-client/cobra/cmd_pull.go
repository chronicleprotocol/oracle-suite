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

package cobra

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func Pull(opts *Options) *cobra.Command {
	var id, contentType string
	var limit int64
	cmd := &cobra.Command{
		Use:          "pull",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := opts.SSBConfig()
			if err != nil {
				return err
			}
			c, err := conf.Client(cmd.Context())
			if err != nil {
				return err
			}
			if id == "" {
				b, err := c.WhoAmI()
				if err != nil {
					return err
				}
				var w struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(b, &w); err != nil {
					return err
				}
				id = w.ID
				log.Println("defaulting to id: ", id)
			}
			last, err := c.ReceiveLast(id, contentType, limit)
			if err != nil {
				return err
			}
			if len(last) > 0 {
				fmt.Println(string(last))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(
		&id,
		"id",
		"",
		"feed id",
	)
	cmd.Flags().StringVar(
		&contentType,
		"type",
		"",
		"content type",
	)
	cmd.Flags().Int64Var(
		&limit,
		"limit",
		-1,
		"max message count",
	)
	return cmd
}
