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

package env

import (
	"os"
	"strings"
)

// String returns a string from the environment variable with the given key.
// If the variable is not set, the default value is returned.
// Empty values are allowed and valid.
func String(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return v
}

// separator is used to split the environment variable values.
// It is taken from CFG_ITEM_SEPARATOR environment variable and defaults to a newline.
var separator = String("CFG_ITEM_SEPARATOR", "\n")

// Strings returns a slice of strings from the environment variable with the
// given key. If the variable is not set, the default value is returned.
// The value is split by the separator defined in the CFG_ITEM_SEPARATOR.
// Values are trimmed of the separator before splitting.
// If the environment variable exists but is empty, an empty slice is returned.
func Strings(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	if v == "" {
		return []string{}
	}
	v = strings.Trim(v, separator)
	return strings.Split(v, separator)
}
