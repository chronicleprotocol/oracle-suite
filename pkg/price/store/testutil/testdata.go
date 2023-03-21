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

	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/median"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var (
	Address1     = types.MustAddressFromHex("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	Address2     = types.MustAddressFromHex("0x8eb3daaf5cb4138f5f96711c09c0cfd0288a36e9")
	PriceAAABBB1 = &messages.Price{
		Price: &median.Price{
			Wat: "AAABBB",
			Val: big.NewInt(10),
			Age: time.Unix(100, 0),
			Sig: types.Signature{
				V: big.NewInt(1),
				R: big.NewInt(1),
				S: big.NewInt(2),
			},
		},
		Trace: nil,
	}
	PriceAAABBB2 = &messages.Price{
		Price: &median.Price{
			Wat: "AAABBB",
			Val: big.NewInt(20),
			Age: time.Unix(200, 0),
			Sig: types.Signature{
				V: big.NewInt(2),
				R: big.NewInt(3),
				S: big.NewInt(4),
			},
		},
		Trace: nil,
	}
	PriceAAABBB3 = &messages.Price{
		Price: &median.Price{
			Wat: "AAABBB",
			Val: big.NewInt(30),
			Age: time.Unix(300, 0),
			Sig: types.Signature{
				V: big.NewInt(3),
				R: big.NewInt(4),
				S: big.NewInt(5),
			},
		},
		Trace: nil,
	}
	PriceAAABBB4 = &messages.Price{
		Price: &median.Price{
			Wat: "AAABBB",
			Val: big.NewInt(30),
			Age: time.Unix(400, 0),
			Sig: types.Signature{
				V: big.NewInt(4),
				R: big.NewInt(5),
				S: big.NewInt(6),
			},
		},
		Trace: nil,
	}
	PriceXXXYYY1 = &messages.Price{
		Price: &median.Price{
			Wat: "XXXYYY",
			Val: big.NewInt(10),
			Age: time.Unix(100, 0),
			Sig: types.Signature{
				V: big.NewInt(5),
				R: big.NewInt(6),
				S: big.NewInt(7),
			},
		},
		Trace: nil,
	}
	PriceXXXYYY2 = &messages.Price{
		Price: &median.Price{
			Wat: "XXXYYY",
			Val: big.NewInt(20),
			Age: time.Unix(200, 0),
			Sig: types.Signature{
				V: big.NewInt(6),
				R: big.NewInt(7),
				S: big.NewInt(8),
			},
		},
		Trace: nil,
	}
)
