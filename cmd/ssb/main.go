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
	"io"
	log2 "log"
	"os"
	"sync"
	"time"

	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/ssb"
	"go.cryptoscope.co/ssb/client"
	"go.cryptoscope.co/ssb/invite"
	"go.cryptoscope.co/ssb/message"
	"go.mindeco.de/log"
	"go.mindeco.de/log/level"
	"go.mindeco.de/log/term"

	ssb2 "github.com/chronicleprotocol/oracle-suite/cmd/keeman/ssb"
	"github.com/chronicleprotocol/oracle-suite/cmd/ssb/cobra"
)

func main() {
	opts, cmd := cobra.Root()
	cmd.PersistentFlags().BoolVarP(
		&opts.Verbose,
		"verbose",
		"v",
		false,
		"verbose logging",
	)
	cmd.PersistentFlags().StringVar(
		&opts.CapsPath,
		"caps",
		"./local.caps.json",
		"caps file path",
	)
	cmd.PersistentFlags().StringVar(
		&opts.KeysPath,
		"keys",
		"./local.ssb.json",
		"caps file path",
	)
	cmd.AddCommand(
		cobra.Push(opts),
		cobra.Pull(opts),
	)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

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

var capsFilePath = "./local.caps.json"

var keyPairPath = "./local.ssb.json"
var feedInvite = "localhost:8008"

var relayInvite = "localhost:8009"

func main0() {
	logger.Log("caps", capsFilePath, "key", keyPairPath)

	caps, err := ssb2.LoadCapsFile(capsFilePath)
	if err != nil {
		handle(err)
		return
	}
	if len(caps.Shs) == 0 {
		caps, err = ssb2.LoadCapsFromConfigFile(capsFilePath)
		if err != nil {
			handle(err)
			return
		}
	}
	keyPair, err := ssb.LoadKeyPair(keyPairPath)
	if err != nil {
		handle(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		invRelay, err := invite.ParseLegacyToken(feedInvite)
		if err != nil {
			handle(err)
			return
		}
		logger.Log("side", "relay", "inv", invRelay.String())

		ctx := context.Background()
		c, err := client.NewTCP(
			keyPair,
			invRelay.Address,
			client.WithSHSAppKey(caps.Shs),
			client.WithContext(ctx),
			client.WithLogger(logger),
		)
		if err != nil {
			handle(err)
			return
		}
		defer closeOrPanic(c)

		whoami, err := c.Whoami()
		if err != nil {
			handle(err)
			return
		}
		logger.Log("side", "relay", "ref", whoami.Ref(), "short", whoami.ShortRef())

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
			return
		}

		for nxt := src.Next(ctx); nxt; nxt = src.Next(ctx) {
			b, err := src.Bytes()
			if err != nil {
				handle(err)
				return
			}
			println(string(b))
		}
	}()

	go func() {
		defer wg.Done()
		invFeed, err := invite.ParseLegacyToken(feedInvite)
		if err != nil {
			handle(err)
			return
		}
		logger.Log("side", "feed", "inv", invFeed.String())

		ctx := context.Background()
		c, err := client.NewTCP(
			keyPair,
			invFeed.Address,
			client.WithSHSAppKey(caps.Shs),
			client.WithContext(ctx),
			client.WithLogger(logger),
		)
		if err != nil {
			handle(err)
			return
		}
		defer closeOrPanic(c)

		whoami, err := c.Whoami()
		if err != nil {
			handle(err)
			return
		}
		logger.Log("side", "feed", "ref", whoami.Ref(), "short", whoami.ShortRef())

		var fap ssb2.FeedAssetPrice
		if err := json.Unmarshal([]byte(ssb2.ContentJSON), &fap); err != nil {
			handle(err)
			return
		}

		for {
			var resp string
			err := c.Async(ctx, &resp, muxrpc.TypeString, muxrpc.Method{"publish"}, fap)
			if err != nil {
				handle(err)
			}
			logger.Log("func", "publish", "side", "feed", "resp", resp)
			time.Sleep(2 * time.Second)
		}
	}()
	wg.Wait()
}

func closeOrPanic(c io.Closer) {
	err := c.Close()
	if err != nil {
		handle(err)
	}
}
func handle(err error) {
	if err := level.Error(logger).Log("msg", err); err != nil {
		log2.Println(err)
	}
}
