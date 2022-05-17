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

package rpcsplitter

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_addError(t *testing.T) {
	tests := []struct {
		err  error
		errs []error
		want string
	}{
		{
			err:  nil,
			errs: []error{errors.New("a")},
			want: "a",
		},
		{
			err:  errors.New("a"),
			errs: []error{errors.New("a")},
			want: "a",
		},
		{
			err:  errors.New("a"),
			errs: []error{errors.New("b")},
			want: "the following errors occurred: [a, b]",
		},
		{
			err:  errors.New("a"),
			errs: []error{errors.New("b"), errors.New("c")},
			want: "the following errors occurred: [a, b, c]",
		},
		{
			err:  errors.New("a"),
			errs: []error{nil, errors.New("b")},
			want: "the following errors occurred: [a, b]",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n), func(t *testing.T) {
			assert.Equal(t, tt.want, addError(tt.err, tt.errs...).Error())
		})
	}
}
