package main

import (
	"fmt"
	"strings"

	gofer "github.com/chronicleprotocol/oracle-suite/pkg/config/gofernext"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

// These are the command options that can be set by CLI flags.
type options struct {
	flag.LoggerFlag
	ConfigFilePath []string
	Format         formatTypeValue
	Config         gofer.Config
	Version        string
}

type formatTypeValue struct {
	format string
}

func (v *formatTypeValue) String() string {
	return v.format
}

func (v *formatTypeValue) Set(s string) error {
	switch strings.ToLower(s) {
	case "plain":
		v.format = "plain"
	case "trace":
		v.format = "trace"
	case "json":
		v.format = "json"
	default:
		return fmt.Errorf("unsupported format")
	}
	return nil
}

func (v *formatTypeValue) Type() string {
	return "plain|trace|json"
}
