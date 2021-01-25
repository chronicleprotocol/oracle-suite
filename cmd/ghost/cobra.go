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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	ghostConfig "github.com/makerdao/gofer/pkg/ghost/config"
	ghostJSON "github.com/makerdao/gofer/pkg/ghost/config/json"
	"github.com/makerdao/gofer/pkg/gofer"
	goferJSON "github.com/makerdao/gofer/pkg/gofer/config/json"
	"github.com/makerdao/gofer/pkg/gofer/feeder"
	"github.com/makerdao/gofer/pkg/gofer/origins"
	"github.com/makerdao/gofer/pkg/log"
	logLogrus "github.com/makerdao/gofer/pkg/log/logrus"
)

func newLogger(level string) (log.Logger, error) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	lr := logrus.New()
	lr.SetLevel(ll)

	return logLogrus.New(lr), nil
}

func newGofer(opts *options, path string, log log.Logger) (*gofer.Gofer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = goferJSON.ParseJSONFile(&opts.GoferConfig, absPath)
	if err != nil {
		return nil, err
	}

	g, err := opts.GoferConfig.BuildGraphs()
	if err != nil {
		return nil, err
	}

	return gofer.NewGofer(g, feeder.NewFeeder(origins.DefaultSet(), log)), nil
}

func newGhost(opts *options, path string, gof *gofer.Gofer, log log.Logger) (*ghostConfig.Instances, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = ghostJSON.ParseJSONFile(&opts.GhostConfig, absPath)
	if err != nil {
		return nil, err
	}

	i, err := opts.GhostConfig.Configure(ghostConfig.Dependencies{
		Context: context.Background(),
		Gofer:   gof,
		Logger:  log,
	})
	if err != nil {
		return nil, err
	}

	return i, nil
}

func NewRunCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			ghostAbsPath, err := filepath.Abs(o.GhostConfigFilePath)
			if err != nil {
				return err
			}

			goferAbsPath, err := filepath.Abs(o.GoferConfigFilePath)
			if err != nil {
				return err
			}

			l, err := newLogger(o.LogVerbosity)
			if err != nil {
				return err
			}

			gof, err := newGofer(o, goferAbsPath, l)
			if err != nil {
				return err
			}

			ins, err := newGhost(o, ghostAbsPath, gof, l)
			if err != nil {
				return err
			}

			err = ins.Ghost.Start()
			if err != nil {
				return err
			}
			defer func() {
				err := ins.Ghost.Stop()
				if err != nil {
					l.Errorf("GHOST", "Unable to stop Ghost: %s", err)
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "ghost",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.LogVerbosity,
		"log.verbosity", "v",
		"info",
		"verbosity level",
	)
	rootCmd.PersistentFlags().StringVarP(
		&opts.GhostConfigFilePath,
		"config", "c",
		"./ghost.json",
		"ghost config file",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.GoferConfigFilePath,
		"config.gofer",
		"./gofer.json",
		"gofer config file",
	)

	return rootCmd
}