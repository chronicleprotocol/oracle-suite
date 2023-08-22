//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package teleportevm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/defiweb/go-eth/types"
)

func Test_packTeleportGUID(t *testing.T) {
	g, err := unpackTeleportGUID(teleportTestGUID)
	require.NoError(t, err)

	b, err := packTeleportGUID(g)
	require.NoError(t, err)
	assert.Equal(t, teleportTestGUID.Bytes(), b)
}

func Test_unpackTeleportGUID(t *testing.T) {
	g, err := unpackTeleportGUID(teleportTestGUID)

	require.NoError(t, err)
	assert.Equal(t, types.MustHashFromHex("0x1111111111111111111111111111111111111111111111111111111111111111", types.PadNone), g.SourceDomain)
	assert.Equal(t, types.MustHashFromHex("0x2222222222222222222222222222222222222222222222222222222222222222", types.PadNone), g.TargetDomain)
	assert.Equal(t, types.MustHashFromHex("0x0000000000000000000000003333333333333333333333333333333333333333", types.PadNone), g.Receiver)
	assert.Equal(t, types.MustHashFromHex("0x0000000000000000000000004444444444444444444444444444444444444444", types.PadNone), g.Operator)
	assert.Equal(t, big.NewInt(55), g.Amount)
	assert.Equal(t, big.NewInt(66), g.Nonce)
	assert.Equal(t, int64(77), g.Timestamp)
}
