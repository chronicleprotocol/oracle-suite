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

package rand

import (
	"bytes"
	"encoding/binary"
	"math/rand"
)

func SeededRandBytes(seedBytes []byte, len int) ([]byte, error) {
	var seed int64
	buf := bytes.NewBuffer(seedBytes)
	err := binary.Read(buf, binary.BigEndian, &seed)
	if err != nil {
		return nil, err
	}
	rand.Seed(seed)
	rb := make([]byte, len)
	rand.Read(rb) //nolint:gosec
	return rb, nil
}

func SeededRandBytesGen(seedBytes []byte, len int) (func() []byte, error) {
	var seed int64
	buf := bytes.NewBuffer(seedBytes)
	err := binary.Read(buf, binary.BigEndian, &seed)
	if err != nil {
		return nil, err
	}
	rand.Seed(seed)
	return func() []byte {
		rb := make([]byte, len)
		rand.Read(rb) //nolint:gosec
		return rb
	}, nil
}
