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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

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

// ParseFile parses the given YAML config file from the byte slice and assigns
// decoded values into the out value.
func ParseFile(out interface{}, path string) error {
	p, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer func() {
		if errClose := f.Close(); err == nil && errClose != nil {
			err = errClose
		}
	}()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to load JSON config file: %w", err)
	}
	return Parse(out, b)
}

// Parse parses the given YAML config from the byte slice and assigns decoded
// values into the out value.
func Parse(out interface{}, config []byte) error {
	n := yaml.Node{}
	if err := yaml.Unmarshal(config, &n); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	if err := yamlReplaceEnvVars(&n); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	if err := n.Decode(out); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	return nil
}

// yamlReplaceEnvVars replaces recursively all environment variables in the
// given YAML node.
func yamlReplaceEnvVars(n *yaml.Node) error {
	return yamlVisitScalarNodes(n, func(n *yaml.Node) error {
		var err error
		parsed := interpolate.Parse(n.Value)
		if parsed.HasVars() {
			n.Value = parsed.Interpolate(func(v interpolate.Variable) string {
				env, ok := getEnv(v.Name)
				if !ok {
					if v.HasDefault {
						return v.Default
					}
					err = fmt.Errorf("environment variable %s not set", v.Name)
					return ""
				}
				return env
			})
			// Removing the style and tag will make the YAML decoder more
			// forgiving. Otherwise, it will complain about type mismatches.
			n.Style = 0
			n.Tag = ""
		}
		return err
	})
}

func yamlVisitScalarNodes(n *yaml.Node, fn func(n *yaml.Node) error) error {
	switch n.Kind {
	default:
		for _, c := range n.Content {
			if err := yamlVisitScalarNodes(c, fn); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		return fn(n)
	}
	return nil
}
