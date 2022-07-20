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

package teleportevm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_packTeleportGUID(t *testing.T) {
	g, err := unpackTeleportGUID(teleportTestGUID)
	require.NoError(t, err)

	b, err := packTeleportGUID(g)
	require.NoError(t, err)
	assert.Equal(t, teleportTestGUID, b)
}

func Test_unpackTeleportGUID(t *testing.T) {
	g, err := unpackTeleportGUID(teleportTestGUID)

	require.NoError(t, err)
	assert.Equal(t, common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"), g.sourceDomain)
	assert.Equal(t, common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"), g.targetDomain)
	assert.Equal(t, common.HexToHash("0x0000000000000000000000003333333333333333333333333333333333333333"), g.receiver)
	assert.Equal(t, common.HexToHash("0x0000000000000000000000004444444444444444444444444444444444444444"), g.operator)
	assert.Equal(t, big.NewInt(55), g.amount)
	assert.Equal(t, big.NewInt(66), g.nonce)
	assert.Equal(t, int64(77), g.timestamp)
}
