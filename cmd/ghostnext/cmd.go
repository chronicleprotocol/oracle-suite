package main

import (
	"github.com/spf13/cobra"

	suite "github.com/chronicleprotocol/oracle-suite"
	ghost "github.com/chronicleprotocol/oracle-suite/pkg/config/ghostnext"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

type options struct {
	flag.LoggerFlag
	ConfigFilePath []string
	Config         ghost.Config
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "ghost",
		Version:       suite.Version,
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().AddFlagSet(flag.NewLoggerFlagSet(&opts.LoggerFlag))
	rootCmd.PersistentFlags().StringSliceVarP(
		&opts.ConfigFilePath,
		"config", "c",
		[]string{"./config.hcl"},
		"ghost config file",
	)

	return rootCmd
}
