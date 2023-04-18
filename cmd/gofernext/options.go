package main

import (
	"github.com/chronicleprotocol/oracle-suite/pkg/config/gofernext"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/flag"
)

// These are the command options that can be set by CLI flags.
type options struct {
	flag.LoggerFlag
	ConfigFilePath []string
	Config         gofer.Config
	Version        string
}
