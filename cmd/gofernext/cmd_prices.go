package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
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
			for _, tick := range ticks {
				if err := tick.Validate(); err != nil {
					fmt.Printf("%s: %v\n", tick.Pair, err)
				} else {
					fmt.Printf("%s: %s\n", tick.Pair, tick.Price.String())
				}
			}
			return nil
		},
	}
}
