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
)

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
	err := json.Unmarshal(config, out)
	if err != nil {
		return err
	}
	return nil
}

func LoadFile(fileName string) (b []byte, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer func() {
		err = f.Close()
	}()
	b, err = ioutil.ReadAll(f)
	return b, err
}
