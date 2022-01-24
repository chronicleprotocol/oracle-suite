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

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ssb"
)

func Push(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:     "push",
		Aliases: []string{"publish"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := opts.SSBConfig()
			if err != nil {
				return err
			}
			c, err := conf.Client(cmd.Context())
			if err != nil {
				return err
			}
			var fap ssb.FeedAssetPrice
			for _, a := range args {
				err = json.Unmarshal([]byte(a), &fap)
				if err != nil {
					return err
				}
				resp, err := c.Transmit(fap)
				if err != nil {
					return err
				}
				if len(resp) > 0 {
					fmt.Println(string(resp))
				}
			}
			return nil
		},
	}
}
