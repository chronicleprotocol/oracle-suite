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

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

func NewPricesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "prices [PAIR...]",
		Aliases: []string{"price"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "Return prices for given PAIRs",
		Long:    `Return prices for given PAIRs.`,
		RunE: func(c *cobra.Command, args []string) (err error) {
			if err := config.LoadFiles(&opts.Config, opts.ConfigFilePath); err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer ctxCancel()
			services, err := opts.Config.Services(opts.Logger())
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			ticks, err := services.PriceProvider.Ticks(ctx, services.PriceProvider.ModelNames(ctx)...)
			if err != nil {
				return err
			}
			marshaled, err := marshalTicks(ticks, opts.Format.format)
			if err != nil {
				return err
			}
			fmt.Println(string(marshaled))
			return nil
		},
	}
}

func marshalTicks(ticks map[string]provider.Tick, format string) ([]byte, error) {
	switch format {
	case formatPlain:
		return marshalTicksPlain(ticks)
	case formatTrace:
		return marshalTicksTrace(ticks)
	case formatJSON:
		return marshalTicksJSON(ticks)
	default:
		return nil, fmt.Errorf("unsupported format")
	}
}

func marshalTicksPlain(ticks map[string]provider.Tick) ([]byte, error) {
	var buf bytes.Buffer
	for _, name := range maputil.SortKeys(ticks, sort.Strings) {
		bts, err := ticks[name].MarshalText()
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("Price for %s:\n", name))
		buf.Write(bts)
	}
	return buf.Bytes(), nil
}

func marshalTicksTrace(ticks map[string]provider.Tick) ([]byte, error) {
	var buf bytes.Buffer
	for _, name := range maputil.SortKeys(ticks, sort.Strings) {
		bts, err := ticks[name].MarshalTrace()
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("Price for %s:\n", name))
		buf.Write(bts)
	}
	return buf.Bytes(), nil
}

func marshalTicksJSON(ticks map[string]provider.Tick) ([]byte, error) {
	return json.Marshal(ticks)
}
