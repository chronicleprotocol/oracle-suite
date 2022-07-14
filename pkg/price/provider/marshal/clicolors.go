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

package marshal

import "strings"

// colorCode represents ANSII escape code for color formatting.
type colorCode string

const (
	reset colorCode = "\033[0m"
	red   colorCode = "\033[31m"
	green colorCode = "\033[32m"
)

var colorsEnabled = true

// disableColors disabled colors rendering in color function.
func disableColors() {
	colorsEnabled = false
}

// color adds given ANSII escape code at beginning of every line.
func color(str string, color colorCode) string {
	if !colorsEnabled {
		return str
	}

	return string(color) + strings.ReplaceAll(str, "\n", "\n"+string(reset+color)) + string(reset)
}
