package main

import (
	"fmt"
	"strings"

	gofer "github.com/chronicleprotocol/oracle-suite/pkg/config/gofernext"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

const (
	formatPlain = "plain"
	formatTrace = "trace"
	formatJSON  = "json"
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
	case formatPlain:
		v.format = formatPlain
	case formatTrace:
		v.format = formatTrace
	case formatJSON, "":
		v.format = formatJSON
	default:
		return fmt.Errorf("unsupported format")
	}
	return nil
}

func (v *formatTypeValue) Type() string {
	return "plain|trace|json"
}
