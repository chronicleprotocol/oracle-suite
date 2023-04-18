package main

import (
	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gofernext",
		Version: opts.Version,
		Short:   "Tool for providing reliable data in the blockchain ecosystem",
		Long: `
Gofer is a tool that provides reliable asset prices taken from various sources.

It is a tool that allows for easy data retrieval from various sources
with aggregates that increase reliability in the DeFi environment.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringSliceVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		[]string{"./config.hcl"},
		"config file",
	)

	return rootCmd
}
