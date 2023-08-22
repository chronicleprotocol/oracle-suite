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

package formatter

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// XFilterFormatter removes all fields with the "x-" prefix. This will allow
// adding more data fields to logs without making the CLI output to messy.
type XFilterFormatter struct {
	Formatter logrus.Formatter
}

func (f *XFilterFormatter) Format(e *logrus.Entry) ([]byte, error) {
	data := logrus.Fields{}
	for k, v := range e.Data {
		if !strings.HasPrefix(k, "x-") {
			data[k] = v
		}
	}
	e.Data = data
	return f.Formatter.Format(e)
}
