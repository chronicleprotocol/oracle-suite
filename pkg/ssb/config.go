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
	"context"
	"errors"
	"net"
	"os"

	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/client"
	"go.mindeco.de/log"
	"go.mindeco.de/log/level"
	"go.mindeco.de/log/term"
)

type Config struct {
	Keys ssb.KeyPair
	Shs  string
	Addr net.Addr
}

func (cfg *Config) Client(ctx context.Context) (*Client, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}
	rpc, err := client.NewTCP(
		cfg.Keys,
		cfg.Addr,
		client.WithSHSAppKey(cfg.Shs),
		client.WithContext(ctx),
		client.WithLogger(logger()),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		ctx:    ctx,
		doneCh: make(chan struct{}),
		rpc:    rpc,
	}, nil
}

func logger() log.Logger {
	colorFn := func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] != "level" {
				continue
			}
			switch keyvals[i+1].(level.Value).String() {
			case "debug":
				return term.FgBgColor{Fg: term.DarkGray}
			case "info":
				return term.FgBgColor{Fg: term.Gray}
			case "warn":
				return term.FgBgColor{Fg: term.Yellow}
			case "error":
				return term.FgBgColor{Fg: term.Red}
			case "crit":
				return term.FgBgColor{Fg: term.Gray, Bg: term.DarkRed}
			default:
				return term.FgBgColor{}
			}
		}
		return term.FgBgColor{}
	}
	l := term.NewColorLogger(os.Stderr, log.NewLogfmtLogger, colorFn)
	return level.NewFilter(l, level.AllowAll())
}
