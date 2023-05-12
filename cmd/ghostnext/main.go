package main

import (
	"os"
)

func main() {
	var opts options
	rootCmd := NewRootCommand(&opts)

	rootCmd.AddCommand(
		NewRunCmd(&opts),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
