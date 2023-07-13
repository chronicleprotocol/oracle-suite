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

package cmd

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	logrus2 "github.com/chronicleprotocol/oracle-suite/pkg/log/logrus"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/logrus/formatter"
)

type LoggerFlags struct {
	verbosityFlag
	formatterFlag
}

func NewLoggerFlagSet(logger *LoggerFlags) *pflag.FlagSet {
	fs := pflag.NewFlagSet("log", pflag.PanicOnError)
	fs.VarP(
		&logger.verbosityFlag,
		"log.verbosity",
		"v",
		"verbosity level",
	)
	fs.VarP(
		&logger.formatterFlag,
		"log.format",
		"f",
		"log format",
	)
	return fs
}

func (logger *LoggerFlags) Logger() log.Logger {
	l := logrus.New()
	l.SetLevel(logger.Verbosity())
	l.SetFormatter(logger.Formatter())
	return logrus2.New(l)
}

const defaultVerbosity = logrus.InfoLevel

type verbosityFlag struct {
	wasSet    bool
	verbosity logrus.Level
}

// String implements the pflag.Value interface.
func (f *verbosityFlag) String() string {
	if !f.wasSet {
		return defaultVerbosity.String()
	}
	return f.verbosity.String()
}

// Set implements the pflag.Value interface.
func (f *verbosityFlag) Set(v string) (err error) {
	f.verbosity, err = logrus.ParseLevel(v)
	if err != nil {
		return err
	}
	f.wasSet = true
	return err
}

// Type implements the pflag.Value interface.
func (f *verbosityFlag) Type() string {
	var s string
	for _, l := range logrus.AllLevels {
		if l == logrus.TraceLevel || l == logrus.FatalLevel { // Don't display unused log levels
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += l.String()
	}
	return s
}

func (f *verbosityFlag) Verbosity() logrus.Level {
	if !f.wasSet {
		return defaultVerbosity
	}
	return f.verbosity
}

// formattersMap is a map of supported logrus formatters. It is safe to add
// custom formatters to this map.
var formattersMap = map[string]func() logrus.Formatter{
	"text": func() logrus.Formatter {
		return &formatter.FieldSerializerFormatter{
			UseJSONRawMessage: false,
			Formatter: &formatter.XFilterFormatter{
				Formatter: &logrus.TextFormatter{},
			},
		}
	},
	"json": func() logrus.Formatter {
		return &formatter.FieldSerializerFormatter{
			UseJSONRawMessage: true,
			Formatter:         &formatter.JSONFormatter{},
		}
	},
}

const defaultFormatter = "text"

// formatter implements pflag.Value. It represents a flag that allow
// to choose a different logrus formatterFlag.
type formatterFlag struct {
	format string
}

// String implements the pflag.Value interface.
func (f *formatterFlag) String() string {
	if f.format == "" {
		return defaultFormatter
	}
	return f.format
}

// Set implements the pflag.Value interface.
func (f *formatterFlag) Set(v string) error {
	v = strings.ToLower(v)
	if _, ok := formattersMap[v]; !ok {
		return fmt.Errorf("unsupported format: %s", v)
	}
	f.format = v
	return nil
}

// Type implements the pflag.Value interface.
func (f *formatterFlag) Type() string {
	return "text|json"
}

// Formatter returns the logrus.Formatter for selected type.
func (f *formatterFlag) Formatter() logrus.Formatter {
	if f.format == "" {
		return formattersMap[defaultFormatter]()
	}
	return formattersMap[f.format]()
}
