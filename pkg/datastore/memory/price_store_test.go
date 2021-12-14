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

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/datastore/memory/testutil"
)

func TestPriceStore_Add(t *testing.T) {
	ps := NewPriceStore()

	ps.Add(testutil.Address1, testutil.PriceAAABBB1)
	ps.Add(testutil.Address1, testutil.PriceXXXYYY1)
	ps.Add(testutil.Address2, testutil.PriceAAABBB1)
	ps.Add(testutil.Address2, testutil.PriceXXXYYY1)

	aaabbb := ps.AssetPair("AAABBB")
	xxxyyy := ps.AssetPair("XXXYYY")

	assert.Equal(t, 2, len(aaabbb))
	assert.Equal(t, 2, len(xxxyyy))
	assert.Contains(t, aaabbb, testutil.PriceAAABBB1)
	assert.Contains(t, xxxyyy, testutil.PriceXXXYYY1)
}

func TestPriceStore_Add_UseNewerPrice(t *testing.T) {
	ps := NewPriceStore()

	// Second price should replace first one because is younger:
	ps.Add(testutil.Address1, testutil.PriceAAABBB1)
	ps.Add(testutil.Address1, testutil.PriceAAABBB2)

	// Second price should be ignored because is older:
	ps.Add(testutil.Address1, testutil.PriceXXXYYY2)
	ps.Add(testutil.Address1, testutil.PriceXXXYYY1)

	aaabbb := ps.AssetPair("AAABBB")
	xxxyyy := ps.AssetPair("XXXYYY")

	assert.Equal(t, testutil.PriceAAABBB2, aaabbb[0])
	assert.Equal(t, testutil.PriceXXXYYY2, xxxyyy[0])
}

func TestPriceStore_Feeder(t *testing.T) {
	ps := NewPriceStore()

	ps.Add(testutil.Address1, testutil.PriceAAABBB1)
	ps.Add(testutil.Address1, testutil.PriceAAABBB2)
	ps.Add(testutil.Address1, testutil.PriceXXXYYY1)
	ps.Add(testutil.Address1, testutil.PriceXXXYYY2)
	ps.Add(testutil.Address2, testutil.PriceAAABBB1)
	ps.Add(testutil.Address2, testutil.PriceAAABBB2)
	ps.Add(testutil.Address2, testutil.PriceXXXYYY1)
	ps.Add(testutil.Address2, testutil.PriceXXXYYY2)

	assert.Equal(t, testutil.PriceAAABBB2, ps.Feeder("AAABBB", testutil.Address1))
	assert.Equal(t, testutil.PriceXXXYYY2, ps.Feeder("XXXYYY", testutil.Address1))
}
