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

package model

import (
	"fmt"
	"strings"
)

type Logs []string

// LogF appends log message with formatting
func (l *Logs) LogF(format string, a ...any) {
	*l = append(*l, fmt.Sprintf(format, a...))
}

// Log appends log message
func (l *Logs) Log(s string) {
	*l = append(*l, s)
}

func (l Logs) String() string {
	if len(l) == 0 {
		return ""
	}
	return " ❇ " + strings.Join(l, "\n ❇ ") + "\n"
}
