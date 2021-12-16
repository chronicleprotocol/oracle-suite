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

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/ssb"
)

func Publish(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use: "publish",
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
				err = c.Publish(fap)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

const ContentJSON = `{"type":"YFIUSD","version":"1.4.2","price":29554.098632102,"priceHex":"0000000000000000000000000000000000000000000006422182799241b0fc00","time":1607032876,"timeHex":"000000000000000000000000000000000000000000000000000000005fc9602c","hash":"9082f4d92f41e539615d293c48e29bf4a9c6d45de289b53b8033928b4ce3a453","signature":"652516621550c5396068c55cd1f4f15d0a2a290dca5a5e54dea8f6bdf3b731f9304f02deb989bcde82ae77832fd83a718323ec602d3a26e34a5160a3740e276e1b","sources":{"binance":"29545.9351062372","coinbase":"29531.4600000000","ftx":"29560.0000000000","gemini":"29615.0500000000","huobi":"29548.1972642045","uniswap":"29640.2199634951"}}`
