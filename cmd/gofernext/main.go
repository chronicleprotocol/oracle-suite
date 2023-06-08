package main

import (
	"context"
	"fmt"
	"os"

	suite "github.com/chronicleprotocol/oracle-suite"
)

// exitCode to be returned by the application.
var exitCode = 0

func main() {
	opts := options{
		Version: suite.Version,
	}

	rootCmd := NewRootCommand(&opts)
	rootCmd.AddCommand(
		NewModelsCmd(&opts),
		NewDataCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n", err)
		if exitCode == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exitCode)
}

func getModelsNames(ctx context.Context, provider datapoint.Provider, args []string) []string {
	if len(args) == 0 {
		return provider.ModelNames(ctx)
	}
	return args
}
