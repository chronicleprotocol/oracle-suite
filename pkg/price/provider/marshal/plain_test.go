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

func TestPlain_Nodes(t *testing.T) {
	var err error
	b := &bytes.Buffer{}
	m := newPlain()

	ab := provider.Pair{Base: "A", Quote: "B"}
	cd := provider.Pair{Base: "C", Quote: "D"}
	ns := testutil.Models(ab, cd)

	err = m.Write(b, ns[ab])
	assert.NoError(t, err)

	err = m.Write(b, ns[cd])
	assert.NoError(t, err)

	err = m.Flush()
	assert.NoError(t, err)

	expected := `
A/B
C/D
`[1:]

	assert.Equal(t, expected, b.String())
}

func TestPlain_Prices(t *testing.T) {
	var err error
	b := &bytes.Buffer{}
	m := newPlain()

	ab := provider.Pair{Base: "A", Quote: "B"}
	cd := provider.Pair{Base: "C", Quote: "D"}
	ns := testutil.Prices(ab, cd)

	err = m.Write(b, ns[ab])
	assert.NoError(t, err)

	cdt := ns[cd]
	cdt.Error = "something"
	err = m.Write(b, ns[cd])
	assert.NoError(t, err)

	err = m.Flush()
	assert.NoError(t, err)

	expected := `
A/B 10.000000
C/D - something
`[1:]

	assert.Equal(t, expected, b.String())
}
