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

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

func NewDataCmd(opts *options) *cobra.Command {
	c := &cobra.Command{
		Use:     "data [models...]",
		Aliases: []string{"data"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "Return data points for given models.",
		Long:    `Return data points for given models.`,
		RunE: func(c *cobra.Command, args []string) (err error) {
			if err := opts.LoadConfigFiles(&opts.Config2); err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer ctxCancel()
			services, err := opts.Config2.Services(opts.Logger())
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			ticks, err := services.DataProvider.DataPoints(ctx, getModelsNames(ctx, services.DataProvider, args)...)
			if err != nil {
				return err
			}
			marshaled, err := marshalDataPoints(ticks, opts.Format2.String())
			if err != nil {
				return err
			}
			fmt.Println(string(marshaled))
			return nil
		},
	}
	c.Flags().VarP(
		&opts.Format2,
		"format",
		"o",
		"output format",
	)
	return c
}

func marshalDataPoints(points map[string]datapoint.Point, format string) ([]byte, error) {
	switch format {
	case formatPlain:
		return marshalDataPointsPlain(points)
	case formatTrace:
		return marshalDataPointsTrace(points)
	case formatJSON:
		return marshalDataPointsJSON(points)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func marshalDataPointsPlain(points map[string]datapoint.Point) ([]byte, error) {
	var buf bytes.Buffer
	for i, name := range maputil.SortKeys(points, sort.Strings) {
		if i > 0 {
			buf.WriteString("\n")
		}
		if err := points[name].Validate(); err != nil {
			buf.WriteString(fmt.Sprintf("%s: %s", name, err))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s", name, points[name].Value.Print()))
		}
	}
	return buf.Bytes(), nil
}

func marshalDataPointsTrace(points map[string]datapoint.Point) ([]byte, error) {
	var buf bytes.Buffer
	for _, name := range maputil.SortKeys(points, sort.Strings) {
		bts, err := points[name].MarshalTrace()
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("Data point for %s:\n", name))
		buf.Write(bts)
	}
	return buf.Bytes(), nil
}

func marshalDataPointsJSON(points map[string]datapoint.Point) ([]byte, error) {
	return json.Marshal(points)
}
