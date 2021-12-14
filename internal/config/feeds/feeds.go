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

package feeds

import (
	"errors"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

type Feeds []string

var ErrInvalidEthereumAddress = errors.New("invalid ethereum address")

func (f *Feeds) Addresses() ([]ethereum.Address, error) {
	var addrs []ethereum.Address
	for _, addr := range *f {
		if !ethereum.IsHexAddress(addr) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEthereumAddress, addr)
		}
		addrs = append(addrs, ethereum.HexToAddress(addr))
	}
	return addrs, nil
}
