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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	gofer "github.com/chronicleprotocol/oracle-suite/pkg/config/gofernext"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/supervisor"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/treerender"
)

func NewModelsCmd(cfg supervisor.Config, cf *cmd.ConfigFlags, lf *cmd.LoggerFlags) *cobra.Command {
	var format formatTypeValue
	cc := &cobra.Command{
		Use:     "models [MODEL...]",
		Aliases: []string{"model"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "List all supported models",
		RunE: func(cc *cobra.Command, args []string) (err error) {
			if err := cf.Load(cfg); err != nil {
				return err
			}
			services, err := cfg.Services(lf.Logger(), cc.Root().Use, cc.Root().Version)
			if err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer ctxCancel()
			if err = services.Start(ctx); err != nil {
				return err
			}
			s, ok := services.(*gofer.Services)
			if !ok {
				return fmt.Errorf("services are not gofer.Services")
			}
			models, err := s.DataProvider.Models(ctx, getModelsNames(ctx, s.DataProvider, args)...)
			if err != nil {
				return err
			}
			marshaled, err := marshalModels(models, format.String())
			if err != nil {
				return err
			}
			fmt.Println(string(marshaled))
			return nil
		},
	}
	cc.Flags().VarP(
		&format,
		"format",
		"o",
		"output format",
	)
	cc.Flags().BoolVar(
		&treerender.NoColors,
		"no-color",
		false,
		"disable output coloring",
	)
	return cc
}

func marshalModels(models map[string]datapoint.Model, format string) ([]byte, error) {
	switch format {
	case formatPlain:
		return marshalModelsPlain(models)
	case formatTrace:
		return marshalModelsTrace(models)
	case formatJSON:
		return marshalModelsJSON(models)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func marshalModelsPlain(models map[string]datapoint.Model) ([]byte, error) {
	var buf bytes.Buffer
	for i, name := range maputil.SortKeys(models, sort.Strings) {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(name)
	}
	return buf.Bytes(), nil
}

func marshalModelsTrace(models map[string]datapoint.Model) ([]byte, error) {
	var buf bytes.Buffer
	for _, name := range maputil.SortKeys(models, sort.Strings) {
		bts, err := models[name].MarshalTrace()
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("Model for %s:\n", name))
		buf.Write(bts)
	}
	return buf.Bytes(), nil
}

func marshalModelsJSON(models map[string]datapoint.Model) ([]byte, error) {
	return json.Marshal(models)
}
