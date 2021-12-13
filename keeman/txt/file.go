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

package txt

import (
	"fmt"
	"log"
	"os"
)

func ReadLines(filename string, limit int, withComment bool) ([]string, error) {
	file, fileClose, err := FileOpen(filename)
	if err != nil {
		return nil, err
	}
	defer fileClose()
	return FileReadLines(file, limit, withComment)
}

func FileReadLines(file *os.File, limit int, withComment bool) ([]string, error) {
	if fileIsEmpty(file) {
		return nil, fmt.Errorf("file is empty: %s", file.Name())
	}
	return ReadNonEmptyLines(file, limit, withComment)
}

func FileOpen(filename string) (*os.File, func(), error) {
	file, err := os.Open(filename)
	return file, func() {
		if err := file.Close(); err != nil {
			log.Fatal(file.Name(), err)
		}
	}, err
}

func fileIsEmpty(file *os.File) bool {
	info, err := file.Stat()
	return err != nil || info.Size() == 0 && info.Mode()&os.ModeNamedPipe == 0
}
