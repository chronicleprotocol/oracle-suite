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

package interpolate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{
		{
			str:  "foo",
			want: "foo",
		},
		{
			str:  "${bar}",
			want: "[bar]",
		},
		{
			str:  "${bar\\}foo}",
			want: "[bar}foo]",
		},
		{
			str:  "foo_${bar}",
			want: "foo_[bar]",
		},
		{
			str:  "${foo}_${bar}",
			want: "[foo]_[bar]",
		},
		{
			str:  "$${foo}_$${bar}",
			want: "${foo}_${bar}",
		},
		{
			str:  "\\${foo}_\\${bar}",
			want: "${foo}_${bar}",
		},
		{
			str:  "$$${foo}_$$${bar}",
			want: "$[foo]_$[bar]",
		},
		{
			str:  "\\\\${foo}_\\\\${bar}",
			want: "\\[foo]_\\[bar]",
		},
		{
			str:  "${",
			want: "${",
		},
		{
			str:  "}",
			want: "}",
		},
		{
			str:  "${foo",
			want: "${foo",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, Parse(tt.str).Interpolate(func(name string) string { return "[" + name + "]" }))
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"foo",
		"${bar}",
		"${bar\\}foo}",
		"foo_${bar}",
		"${foo}_${bar}",
		"$${foo}_$${bar}",
		"\\${foo}_\\${bar}",
		"$$${foo}_$$${bar}",
		"\\\\${foo}_\\\\${bar}",
		"${",
		"}",
		"${foo",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		Parse(s).Interpolate(func(name string) string { return "[" + name + "]" })
	})
}
