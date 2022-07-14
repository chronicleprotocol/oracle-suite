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

package marshal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/marshal/testutil"
)

func TestTrace_Graph(t *testing.T) {
	disableColors()

	var err error
	b := &bytes.Buffer{}
	m := newTrace()

	ab := provider.Pair{Base: "A", Quote: "B"}
	ns := testutil.Models(ab)

	err = m.Write(b, ns[ab])
	assert.NoError(t, err)

	err = m.Flush()
	assert.NoError(t, err)

	expected := `
Graph for A/B:
───median(pair:A/B)
   ├──origin(origin:a, pair:A/B)
   ├──indirect(pair:A/B)
   │  └──origin(origin:a, pair:A/B)
   └──median(pair:A/B)
      ├──origin(origin:a, pair:A/B)
      └──origin(origin:b, pair:A/B)
`[1:]

	assert.Equal(t, expected, b.String())
}

func TestTrace_Prices(t *testing.T) {
	disableColors()

	var err error
	b := &bytes.Buffer{}
	m := newTrace()

	ab := provider.Pair{Base: "A", Quote: "B"}
	ts := testutil.Prices(ab)

	err = m.Write(b, ts[ab])
	assert.NoError(t, err)

	err = m.Flush()
	assert.NoError(t, err)

	expected := `
Price for A/B:
───aggregator(method:median, minimumSuccessfulSources:1, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
   ├──origin(origin:a, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
   ├──aggregator(method:indirect, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
   │  └──origin(origin:a, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
   └──aggregator(method:median, minimumSuccessfulSources:1, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
      ├──origin(origin:a, pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z)
      └──origin(origin:b, pair:A/B, price:20, timestamp:1970-01-01T00:00:20Z)
            Error: something
`[1:]

	assert.Equal(t, expected, b.String())
}
