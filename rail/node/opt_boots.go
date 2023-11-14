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

package node

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var defaultBoots = []string{
	"/dns/spire-bootstrap1.chroniclelabs.io/tcp/8000/p2p/12D3KooWFYkJ1SghY4KfAkZY9Exemqwnh4e4cmJPurrQ8iqy2wJG",
	"/dns/spire-bootstrap2.chroniclelabs.io/tcp/8000/p2p/12D3KooWD7eojGbXT1LuqUZLoewRuhNzCE2xQVPHXNhAEJpiThYj",
	"/dns/spire-bootstrap1.staging.chroniclelabs.io/tcp/8000/p2p/12D3KooWHoSyTgntm77sXShoeX9uNkqKNMhHxKtskaHqnA54SrSG",
	"/ip4/178.128.141.30/tcp/8000/p2p/12D3KooWLaMPReGaxFc6Z7BKWTxZRbxt3ievW8Np7fpA6y774W9T",
	"/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi",
	"/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR",
}

func BootName(pid peer.ID) string {
	for _, addr := range defaultBoots {
		a, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Error(err)
			continue
		}
		ma, p := peer.SplitAddr(a)
		if p == pid {
			return ma.String()
		}
	}
	return ""
}
