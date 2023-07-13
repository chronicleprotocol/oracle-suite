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

package main

import (
	"fmt"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/cmd"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/gofer"
	gofer2 "github.com/chronicleprotocol/oracle-suite/pkg/config/gofernext"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal"
)

// These are the command options that can be set by CLI flags.
type options struct {
	cmd.LoggerFlags
	cmd.FilesFlags
	Format  formatTypeValue
	Config  gofer.Config
	NoRPC   bool
	Format2 formatTypeValue2
	Config2 gofer2.Config
}

var formatMap = map[marshal.FormatType]string{
	marshal.Plain:  "plain",
	marshal.Trace:  "trace",
	marshal.JSON:   "json",
	marshal.NDJSON: "ndjson",
}

// formatTypeValue is a wrapper for the FormatType to allow implement
// the flag.Value and spf13.pflag.Value interfaces.
type formatTypeValue struct {
	format marshal.FormatType
}

// Will return the default value if none is set
// and will fail if the `format` is set to an unsupported value for some reason.
func (v *formatTypeValue) String() string {
	if v != nil {
		return formatMap[v.format]
	}
	return formatMap[marshal.Plain]
}

func (v *formatTypeValue) Set(s string) error {
	s = strings.ToLower(s)

	for ct, st := range formatMap {
		if s == st {
			v.format = ct
			return nil
		}
	}

	return fmt.Errorf("unsupported format: %s", s)
}

func (v *formatTypeValue) Type() string {
	return "plain|trace|json|ndjson"
}

const (
	formatPlain = "plain"
	formatTrace = "trace"
	formatJSON  = "json"
)

type formatTypeValue2 struct {
	format string
}

func (v *formatTypeValue2) String() string {
	if v.format == "" {
		return formatPlain
	}
	return v.format
}

func (v *formatTypeValue2) Set(s string) error {
	switch strings.ToLower(s) {
	case formatPlain:
		v.format = formatPlain
	case formatTrace:
		v.format = formatTrace
	case formatJSON:
		v.format = formatJSON
	default:
		return fmt.Errorf("unsupported format: %s", s)
	}
	return nil
}

func (v *formatTypeValue2) Type() string {
	return "plain|trace|json"
}
