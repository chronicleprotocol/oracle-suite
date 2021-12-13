//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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

package pflag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/makerdao/oracle-suite/pkg/log"
	logLogrus "github.com/makerdao/oracle-suite/pkg/log/logrus"
	logrusFormatter "github.com/makerdao/oracle-suite/pkg/log/logrus/formatter"
)

type LoggerFlag struct {
	verbosity
	formatter
}

func NewLoggerFlagSet(logger *LoggerFlag) *pflag.FlagSet {
	fs := pflag.NewFlagSet("log", pflag.PanicOnError)
	fs.VarP(
		&logger.verbosity,
		"log.verbosity",
		"v",
		"verbosity level",
	)
	fs.Var(
		&logger.formatter,
		"log.format",
		"log format",
	)
	return fs
}

func (logger *LoggerFlag) Logger() log.Logger {
	l := logrus.New()
	l.SetLevel(logger.Verbosity())
	l.SetFormatter(logger.Formatter())
	return logLogrus.New(l)
}

const defaultVerbosity = logrus.WarnLevel

type verbosity struct {
	wasSet    bool
	verbosity logrus.Level
}

// String implements the pflag.Value interface.
func (f *verbosity) String() string {
	if !f.wasSet {
		return defaultVerbosity.String()
	}
	return f.verbosity.String()
}

// Set implements the pflag.Value interface.
func (f *verbosity) Set(v string) (err error) {
	f.verbosity, err = logrus.ParseLevel(v)
	if err != nil {
		return err
	}
	f.wasSet = true
	return err
}

// Type implements the pflag.Value interface.
func (f *verbosity) Type() string {
	var s string
	for _, l := range logrus.AllLevels {
		if len(s) > 0 {
			s += "|"
		}
		s += l.String()
	}
	return s
}

func (f *verbosity) Verbosity() logrus.Level {
	if f.verbosity == 0 {
		return defaultVerbosity
	}
	return f.verbosity
}

// formattersMap is a map of supported logrus formatters. It is safe to add
// custom formatters to this map.
var formattersMap = map[string]func() logrus.Formatter{
	"text": func() logrus.Formatter {
		return &logrusFormatter.XFilterFormatter{Formatter: &logrus.TextFormatter{}}
	},
	"json": func() logrus.Formatter {
		return &logrusFormatter.JSONFormatter{}
	},
}

const defaultFormatter = "text"

// formatter implements pflag.Value. It represents a flag that allow
// to choose a different logrus formatter.
type formatter struct {
	format string
}

// String implements the pflag.Value interface.
func (f *formatter) String() string {
	if f.format == "" {
		return defaultFormatter
	}
	return f.format
}

// Set implements the pflag.Value interface.
func (f *formatter) Set(v string) error {
	v = strings.ToLower(v)
	if _, ok := formattersMap[v]; !ok {
		return fmt.Errorf("unsupported format")
	}
	f.format = v
	return nil
}

// Type implements the pflag.Value interface.
func (f *formatter) Type() string {
	return "text|json"
}

// Formatter returns the logrus.Formatter for selected type.
func (f *formatter) Formatter() logrus.Formatter {
	if f.format == "" {
		return formattersMap[defaultFormatter]()
	}
	return formattersMap[f.format]()
}