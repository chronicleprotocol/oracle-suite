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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/interpolate"
)

var getEnv = os.LookupEnv

func LoadFile(fileName string) (b []byte, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer func() {
		if errClose := f.Close(); err == nil && errClose != nil {
			err = errClose
		}
	}()
	b, err = ioutil.ReadAll(f)
	return b, err
}

func ParseFile(out interface{}, path string) error {
	p, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to load JSON config file: %w", err)
	}
	return Parse(out, b)
}

func Parse(out interface{}, config []byte) error {
	if err := json.Unmarshal(config, out); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}
	return replaceEnvVars(out)
}

func replaceEnvVars(v interface{}) error {
	var err error
	recur(reflect.ValueOf(v), func(s string) string {
		return interpolate.Parse(s).Interpolate(func(key string) string {
			if err != nil {
				return ""
			}
			if !strings.HasPrefix(key, "ENV:") {
				err = fmt.Errorf("environment variable %s does not start with ENV", key)
				return ""
			}
			env, ok := getEnv(key[4:])
			if !ok {
				err = fmt.Errorf("environment variable %s not found", key[4:])
				return ""
			}
			return env
		})
	})
	return err
}

func recur(rv reflect.Value, fn func(rv string) string) {
	switch rv.Kind() {
	case reflect.Struct:
		for n := 0; n < rv.NumField(); n++ {
			recur(rv.Field(n), fn)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			recur(rv.Index(i), fn)
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			if rv.MapIndex(k).Kind() == reflect.String {
				rv.SetMapIndex(k, reflect.ValueOf(fn(rv.MapIndex(k).String())))
				continue
			}
			recur(rv.MapIndex(k), fn)
		}
	case reflect.Ptr, reflect.Interface:
		recur(rv.Elem(), fn)
	case reflect.String:
		if rv.CanAddr() {
			rv.SetString(fn(rv.String()))
		}
	}
}
