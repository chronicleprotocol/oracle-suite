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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	getEnv = func(v string) (string, bool) {
		if v == "nil" {
			return "", false
		}
		if v == "num" {
			return "1", true
		}
		return "env:" + v, true
	}
	defer func() { getEnv = os.LookupEnv }()
	tests := []struct {
		config  string
		out     interface{}
		want    interface{}
		wantErr bool
	}{
		{
			config: `{"foo": "bar"}`,
			out:    &struct{ Foo string }{},
			want:   &struct{ Foo string }{Foo: "bar"},
		},
		{
			config: `{"foo": "bar_${env}"}`,
			out:    &struct{ Foo string }{},
			want:   &struct{ Foo string }{Foo: "bar_env:env"},
		},
		{
			config: `{"foo_${env}": "bar"}`,
			out:    map[string]string{},
			want:   map[string]string{"foo_env:env": "bar"},
		},
		{
			config: `{"foo": {"bar": "baz_${env}"}}`,
			out:    &struct{ Foo map[string]string }{},
			want:   &struct{ Foo map[string]string }{Foo: map[string]string{"bar": "baz_env:env"}},
		},
		{
			config: `{"foo": ["bar_${env}"]}`,
			out:    &struct{ Foo []string }{},
			want:   &struct{ Foo []string }{Foo: []string{"bar_env:env"}},
		},
		{
			config: `{"foo": "${num}"}`,
			out:    &struct{ Foo int }{},
			want:   &struct{ Foo int }{Foo: 1},
		},
		{
			config: "foo:\n  - bar_${env}\n  - baz_${env}\n",
			out:    &struct{ Foo []string }{},
			want:   &struct{ Foo []string }{Foo: []string{"bar_env:env", "baz_env:env"}},
		},
		{
			config: `{"foo": "bar_${nil-baz}"}`,
			out:    &struct{ Foo string }{},
			want:   &struct{ Foo string }{Foo: "bar_baz"},
		},
		{
			config:  `{"foo": ["bar_${nil}"]}`,
			out:     &struct{ Foo []string }{},
			wantErr: true,
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			if tt.wantErr {
				assert.Error(t, Parse(tt.out, []byte(tt.config)))
				return
			}
			assert.NoError(t, Parse(tt.out, []byte(tt.config)))
			assert.Equal(t, tt.want, tt.out)
		})
	}
}
