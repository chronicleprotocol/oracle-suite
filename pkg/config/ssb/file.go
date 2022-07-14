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

package ssb

import (
	"encoding/json"
	"log"
	"net"
	"strconv"

	"go.cryptoscope.co/ssb/invite"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
)

type Caps struct {
	Shs    string `json:"shs"`
	Sign   string `json:"sign,omitempty"`
	Invite string `json:"invite,omitempty"`
}

func LoadCapsFromConfigFile(fileName string) (Caps, error) {
	b, err := config.LoadFile(fileName)
	if err != nil {
		return Caps{}, err
	}
	var c struct {
		Caps Caps `json:"caps"`
	}
	return c.Caps, json.Unmarshal(b, &c)
}

func LoadCapsFile(fileName string) (Caps, error) {
	b, err := config.LoadFile(fileName)
	if err != nil {
		return Caps{}, err
	}
	var c Caps
	return c, json.Unmarshal(b, &c)
}

type connections map[string][]struct {
	Port      int    `json:"port"`
	Transform string `json:"transform,omitempty"`
	Scope     string `json:"scope,omitempty"`
	Host      string `json:"host,omitempty"`
}

func (c connections) hostPort() string {
	for _, v := range c["net"] {
		if v.Scope == "local" {
			if v.Host == "" {
				return "localhost:" + strconv.Itoa(v.Port)
			}
			return v.Host + ":" + strconv.Itoa(v.Port)
		}
	}
	return ""
}

type Config struct {
	Connections struct {
		Incoming connections `json:"incoming"`
		Outgoing connections `json:"outgoing"`
	} `json:"connections"`
	Caps    Caps `json:"caps"`
	Logging struct {
		Level string `json:"level"`
	} `json:"logging"`
	Master []string `json:"master"`
}

func (c Config) Address() (net.Addr, error) {
	log.Println(c.Connections.Incoming.hostPort())
	inv, err := invite.ParseLegacyToken(c.Connections.Incoming.hostPort())
	if err != nil {
		return nil, err
	}
	return inv.Address, nil
}
