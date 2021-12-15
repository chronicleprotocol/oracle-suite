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
	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/invite"

	ssb2 "github.com/chronicleprotocol/oracle-suite/cmd/keeman/ssb"
	ssb3 "github.com/chronicleprotocol/oracle-suite/pkg/ssb"
)

type Options struct {
	CapsPath string
	KeysPath string
	Verbose  bool
}

func Root() (*Options, *cobra.Command) {
	return &Options{}, &cobra.Command{
		Use: "ssb",
	}
}
func Push(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use: "push",
		RunE: func(cmd *cobra.Command, args []string) error {
			caps, err := ssb2.LoadCapsFile(opts.CapsPath)
			if err != nil {
				return err
			}
			if len(caps.Shs) == 0 {
				caps, err = ssb2.LoadCapsFromConfigFile(opts.CapsPath)
				if err != nil {
					return err
				}
			}
			keys, err := ssb.LoadKeyPair(opts.KeysPath)
			if err != nil {
				return err
			}
			inv, err := invite.ParseLegacyToken(caps.Invite)
			if err != nil {
				return err
			}
			conf := ssb3.ClientConfig{
				Keys:   keys,
				Shs:    caps.Shs,
				Invite: inv,
			}
			c, err := ssb3.NewClient(cmd.Context(), conf)
			if err != nil {
				return err
			}
			var fap ssb2.FeedAssetPrice
			err = json.Unmarshal([]byte(ssb2.ContentJSON), &fap)
			if err != nil {
				return err
			}
			return c.PublishPrice(fap)
		},
	}
}
func Pull(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use: "pull",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
