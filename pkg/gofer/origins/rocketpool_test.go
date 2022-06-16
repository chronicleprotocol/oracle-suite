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

package origins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseSelector(t *testing.T) {
	tests := []struct {
		selector string
		want     []byte
		wantErr  assert.ErrorAssertionFunc
	}{
		{"ddd(uint256)(bytes32)", []byte(`s`), assert.NoError},
	}
	for k, tt := range tests {
		t.Run(fmt.Sprintf("parseSelector Test%d", k), func(t *testing.T) {
			got, err := parseSelector(tt.selector)
			if !tt.wantErr(t, err, fmt.Sprintf("parseSelector(%v)", tt.selector)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseSelector(%v)", tt.selector)
		})
	}
}
