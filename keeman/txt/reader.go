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
	"bufio"
	"io"
	"strings"
	"unicode"
)

func ReadNonEmptyLines(r io.Reader, limit int, withComments bool) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		if !withComments {
			text = stripFromFirstChar(text, "#;")
		}

		text = strings.Trim(text, "\t \n")
		if len(text) > 0 {
			lines = append(lines, text)
		}

		if 0 < limit && limit <= len(lines) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func stripFromFirstChar(s, chars string) string {
	if cut := strings.IndexAny(s, chars); cut >= 0 {
		return strings.TrimRightFunc(s[:cut], unicode.IsSpace)
	}
	return s
}
