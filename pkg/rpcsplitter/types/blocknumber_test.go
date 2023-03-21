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

package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BlockNumberType_Unmarshal(t *testing.T) {
	tests := []struct {
		arg        string
		want       BlockNumber
		wantErr    bool
		isTag      bool
		isEarliest bool
		isLatest   bool
		isPending  bool
	}{
		{arg: `"0x0"`, want: Uint64ToBlockNumber(0)},
		{arg: `"0xF"`, want: Uint64ToBlockNumber(15)},
		{arg: `"0"`, want: Uint64ToBlockNumber(0)},
		{arg: `"F"`, want: Uint64ToBlockNumber(15)},
		{arg: `"earliest"`, want: EarliestBlockNumber, isTag: true, isEarliest: true},
		{arg: `"latest"`, want: LatestBlockNumber, isTag: true, isLatest: true},
		{arg: `"pending"`, want: PendingBlockNumber, isTag: true, isPending: true},
		{arg: `"foo"`, wantErr: true},
		{arg: `"0xZ"`, wantErr: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			v := &BlockNumber{}
			err := v.UnmarshalJSON([]byte(tt.arg))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, *v)
				assert.Equal(t, tt.isTag, v.IsTag())
				assert.Equal(t, tt.isEarliest, v.IsEarliest())
				assert.Equal(t, tt.isLatest, v.IsLatest())
				assert.Equal(t, tt.isPending, v.IsPending())
			}
		})
	}
}

func Test_BlockNumberType_Marshal(t *testing.T) {
	tests := []struct {
		arg  BlockNumber
		want string
	}{
		{arg: Uint64ToBlockNumber(0), want: `"0x0"`},
		{arg: Uint64ToBlockNumber(15), want: `"0xf"`},
		{arg: EarliestBlockNumber, want: `"earliest"`},
		{arg: LatestBlockNumber, want: `"latest"`},
		{arg: PendingBlockNumber, want: `"pending"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			j, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(j))
		})
	}
}
