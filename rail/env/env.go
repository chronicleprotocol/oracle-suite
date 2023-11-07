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
	"encoding/hex"
	"os"
	"strings"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("rail/env")

// String returns a string from the environment variable with the given key.
// If the variable is not set, the default value is returned.
// Empty values are allowed and valid.
func String(key, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Debugw("env not set", "var", key)
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
		log.Debugw("env not set", "var", key)
		return def
	}
	if v == "" {
		return []string{}
	}
	v = strings.Trim(v, separator)
	return strings.Split(v, separator)
}

func HexBytes(key string, def []byte) []byte {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Debugw("env not set", "var", key)
		return def
	}
	b, err := hex.DecodeString(v)
	if err != nil {
		log.Warnw("unable to decode hex", "var", key, "err", err)
		return def
	}
	return b
}

func HexBytesSize(key string, l int, def []byte) []byte {
	b := HexBytes(key, def)
	if len(b) != l {
		log.Warnf("invalid bytes length - want: %d, got: %d", l, len(b))
	}
	return b
}
