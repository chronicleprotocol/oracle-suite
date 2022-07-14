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
	"net"

	"github.com/spf13/cobra"
	"go.cryptoscope.co/netwrap"
	"go.cryptoscope.co/secretstream"
	ssbServer "go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/invite"

	suite "github.com/chronicleprotocol/oracle-suite"
	ssbConf "github.com/chronicleprotocol/oracle-suite/pkg/config/ssb"
	"github.com/chronicleprotocol/oracle-suite/pkg/ssb"
)

type Options struct {
	CapsPath string
	KeysPath string
	SsbHost  string
	SsbPort  int
	Verbose  bool
}

func (opts *Options) SSBConfig() (*ssb.Config, error) {
	keys, err := ssbServer.LoadKeyPair(opts.KeysPath)
	if err != nil {
		return nil, err
	}
	caps, err := ssbConf.LoadCapsFile(opts.CapsPath)
	if err != nil {
		return nil, err
	}
	if caps.Shs == "" || caps.Sign == "" {
		caps, err = ssbConf.LoadCapsFromConfigFile(opts.CapsPath)
		if err != nil {
			return nil, err
		}
	}
	if caps.Invite != "" {
		inv, err := invite.ParseLegacyToken(caps.Invite)
		if err != nil {
			return nil, err
		}
		return &ssb.Config{
			Keys: keys,
			Shs:  caps.Shs,
			Addr: inv.Address,
		}, nil
	}
	ip := net.ParseIP(opts.SsbHost)
	if ip == nil {
		resolvedAddr, err := net.ResolveIPAddr("ip", opts.SsbHost)
		if err != nil {
			return nil, err
		}
		ip = resolvedAddr.IP
	}
	return &ssb.Config{
		Keys: keys,
		Shs:  caps.Shs,
		Addr: netwrap.WrapAddr(
			&net.TCPAddr{
				IP:   ip,
				Port: opts.SsbPort,
			},
			secretstream.Addr{PubKey: keys.ID().PubKey()},
		),
	}, nil
}

func Root() (*Options, *cobra.Command) {
	return &Options{}, &cobra.Command{
		Version:            suite.Version,
		Use:                "ssb-rpc-client",
		DisableAutoGenTag:  true,
		DisableSuggestions: true,
		SilenceUsage:       true,
	}
}
