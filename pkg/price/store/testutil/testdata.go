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

package testutil

import (
	"math/big"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var (
	Address1     = ethereum.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	Address2     = ethereum.HexToAddress("0x8eb3daaf5cb4138f5f96711c09c0cfd0288a36e9")
	PriceAAABBB1 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "AAABBB",
			Val:     big.NewInt(10),
			Age:     time.Unix(100, 0),
			V:       1,
			R:       [32]byte{1},
			S:       [32]byte{2},
			StarkR:  []byte{3},
			StarkS:  []byte{4},
			StarkPK: []byte{5},
		},
		Trace: nil,
	}
	PriceAAABBB2 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "AAABBB",
			Val:     big.NewInt(20),
			Age:     time.Unix(200, 0),
			V:       2,
			R:       [32]byte{3},
			S:       [32]byte{4},
			StarkR:  []byte{5},
			StarkS:  []byte{6},
			StarkPK: []byte{7},
		},
		Trace: nil,
	}
	PriceAAABBB3 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "AAABBB",
			Val:     big.NewInt(30),
			Age:     time.Unix(300, 0),
			V:       3,
			R:       [32]byte{4},
			S:       [32]byte{5},
			StarkR:  []byte{6},
			StarkS:  []byte{7},
			StarkPK: []byte{8},
		},
		Trace: nil,
	}
	PriceAAABBB4 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "AAABBB",
			Val:     big.NewInt(30),
			Age:     time.Unix(400, 0),
			V:       4,
			R:       [32]byte{5},
			S:       [32]byte{6},
			StarkR:  []byte{7},
			StarkS:  []byte{8},
			StarkPK: []byte{9},
		},
		Trace: nil,
	}
	PriceXXXYYY1 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "XXXYYY",
			Val:     big.NewInt(10),
			Age:     time.Unix(100, 0),
			V:       5,
			R:       [32]byte{6},
			S:       [32]byte{7},
			StarkR:  []byte{8},
			StarkS:  []byte{9},
			StarkPK: []byte{10},
		},
		Trace: nil,
	}
	PriceXXXYYY2 = &messages.Price{
		Price: &oracle.Price{
			Wat:     "XXXYYY",
			Val:     big.NewInt(20),
			Age:     time.Unix(200, 0),
			V:       6,
			R:       [32]byte{7},
			S:       [32]byte{8},
			StarkR:  []byte{9},
			StarkS:  []byte{10},
			StarkPK: []byte{11},
		},
		Trace: nil,
	}
)
