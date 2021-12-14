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

package mocks

import (
	"reflect"

	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
)

type Gofer struct {
	mock.Mock
}

func (g *Gofer) Models(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Model, error) {
	args := g.Called(interfaceSlice(pairs)...)
	return args.Get(0).(map[gofer.Pair]*gofer.Model), args.Error(1)
}

func (g *Gofer) Price(pair gofer.Pair) (*gofer.Price, error) {
	args := g.Called(pair)
	return args.Get(0).(*gofer.Price), args.Error(1)
}

func (g *Gofer) Prices(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Price, error) {
	args := g.Called(interfaceSlice(pairs)...)
	return args.Get(0).(map[gofer.Pair]*gofer.Price), args.Error(1)
}

func (g *Gofer) Pairs() ([]gofer.Pair, error) {
	args := g.Called()
	return args.Get(0).([]gofer.Pair), args.Error(1)
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("interfaceSlice() given a non-slice type")
	}
	if s.IsNil() {
		return nil
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}
