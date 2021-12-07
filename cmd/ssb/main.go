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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/client"
	"go.cryptoscope.co/ssb/invite"
	"go.cryptoscope.co/ssb/message"
	"go.mindeco.de/log"
	"go.mindeco.de/log/level"
	"go.mindeco.de/log/term"
)

const contentJSON = `{
		"type": "YFIUSD",
		"version": "1.4.2",
		"price": 29554.098632102,
		"priceHex": "0000000000000000000000000000000000000000000006422182799241b0fc00",
		"time": 1607032876,
		"timeHex": "000000000000000000000000000000000000000000000000000000005fc9602c",
		"hash": "9082f4d92f41e539615d293c48e29bf4a9c6d45de289b53b8033928b4ce3a453",
		"signature": "652516621550c5396068c55cd1f4f15d0a2a290dca5a5e54dea8f6bdf3b731f9304f02deb989bcde82ae77832fd83a718323ec602d3a26e34a5160a3740e276e1b",
		"sources": {"binance": "29545.9351062372","coinbase": "29531.4600000000","ftx": "29560.0000000000","gemini": "29615.0500000000","huobi": "29548.1972642045","uniswap": "29640.2199634951"}
	}`

var logger log.Logger

func init() {
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

	logger = term.NewColorLogger(os.Stderr, log.NewLogfmtLogger, colorFn)
	logger = level.NewFilter(logger, level.AllowAll())
}

func main() {
	capsFile, err := loadCapsFile("./local.caps.json")
	if err != nil {
		handle(err)
	}
	keyPair, err := ssb.LoadKeyPair("./local.ssb.json")
	if err != nil {
		handle(err)
	}

	invRelay, err := invite.ParseLegacyToken("localhost:8009:@go5eQFIWwQng+dN911MDRDHgmN7gZ741wE01e52iTeU=.ed25519~dQ64FuGjIvAIOu798nD/HY/R1t/tivQD9QR++VVBnpM=")
	if err != nil {
		handle(err)
	}
	println(invRelay.String())

	ctx := context.Background()
	c, err := client.NewTCP(
		keyPair,
		invRelay.Address,
		client.WithSHSAppKey(capsFile.Shs),
		client.WithContext(ctx),
		client.WithLogger(logger),
	)
	if err != nil {
		handle(err)
	}
	defer closeOrPanic(c)

	whoami, err := c.Whoami()
	if err != nil {
		handle(err)
	}
	println(whoami.Ref(), whoami.ShortRef())

	src, err := c.CreateLogStream(message.CreateLogArgs{
		CommonArgs: message.CommonArgs{
			Live: true,
		},
		StreamArgs: message.StreamArgs{
			Limit:   -1,
			Reverse: true,
		},
	})
	if err != nil {
		handle(err)
	}

	for nxt := src.Next(ctx); nxt; nxt = src.Next(ctx) {
		b, err := src.Bytes()
		if err != nil {
			handle(err)
		}
		println(string(b))
	}

	// invFeed, err := invite.ParseLegacyToken("localhost:8008:@KM4CFRlL8LcUMc0OgNNZov7Nz9oKac3HuRY4IeMIkCc=.ed25519~dQ64FuGjIvAIOu798nD/HY/R1t/tivQD9QR++VVBnpM=")
	// if err != nil {
	// 	handle(err)
	// }
	// println(invFeed.String())
}

type FeedAssetPrice struct {
	Type      string          `json:"type"`
	Version   string          `json:"version"`
	Price     float64         `json:"price"`
	PriceHex  string          `json:"priceHex"`
	Time      int             `json:"time"`
	TimeHex   string          `json:"timeHex"`
	Hash      string          `json:"hash"`
	Signature string          `json:"signature"`
	Sources   json.RawMessage `json:"sources"`
}

func closeOrPanic(c io.Closer) {
	err := c.Close()
	if err != nil {
		handle(err)
	}
}
func handle(err error) {
	if err := level.Error(logger).Log("msg", err); err != nil {
		panic(err)
	}
	os.Exit(1)
}

type caps struct {
	Shs    string `json:"shs"`
	Sign   string `json:"sign"`
	Invite string `json:"invite"`
}

func loadCapsFile(fileName string) (caps, error) {
	b, err := loadFile(fileName)
	if err != nil {
		return caps{}, err
	}
	var c caps
	return c, json.Unmarshal(b, &c)
}

func loadFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}

		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer closeOrPanic(f)

	return ioutil.ReadAll(f)
}
