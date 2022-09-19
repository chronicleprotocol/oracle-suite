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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{
		{
			str:  "",
			want: "",
		},
		{
			str:  "foo",
			want: "foo",
		},
		{
			str:  "${bar}",
			want: "[bar]",
		},
		{
			str:  "${bar:=baz}",
			want: "[baz]",
		},
		{
			str:  "${bar\\:=baz}",
			want: "[bar:=baz]",
		},
		{
			str:  "foo_${bar}_baz",
			want: "foo_[bar]_baz",
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
			str:  "$$",
			want: "$",
		},
		{
			str:  "\\",
			want: "\\",
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
			str:  "${\\",
			want: "${\\",
		},
		{
			str:  "${foo",
			want: "${foo",
		},
		{
			str:  "${foo$$bar${baz\\}",
			want: "${foo$bar${baz}",
		},
		{
			str:  "$0${$",
			want: "$0${$",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			assert.Equal(t, tt.want, Parse(tt.str).Interpolate(func(variable Variable) string {
				if variable.Default != "" {
					return "[" + variable.Default + "]"
				}
				return "[" + variable.Name + "]"
			}))
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		// Literals:
		"",
		"foo",
		string([]byte{0}),
		// Sample inputs:
		"${foo}",
		"${foo:=bar}",
		"$${foo}",
		"\\${foo}",
		// Tokens:
		"$$",
		"\\",
		"${",
		"}",
		":=",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		Parse(s).Interpolate(func(variable Variable) string { return "[" + variable.Name + "]" })
	})
}

func BenchmarkParse(b *testing.B) {
	testString := "before_${foo}_${bar}_after"

	b.Run("parser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Parse(testString).Interpolate(func(variable Variable) string { return "[" + variable.Name + "]" })
		}
	})
	b.Run("preparsed", func(b *testing.B) {
		s := Parse(testString)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Interpolate(func(variable Variable) string { return "[" + variable.Name + "]" })
		}
	})
	b.Run("regexp", func(b *testing.B) {
		rx := regexp.MustCompile(`\$\{[^}]+}`)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rx.ReplaceAllStringFunc(testString, func(s string) string { return "[" + s[2:len(s)-1] + "]" })
		}
	})
}
