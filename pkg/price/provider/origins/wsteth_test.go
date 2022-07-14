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

package origins

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"

	"github.com/stretchr/testify/suite"
)

type WrappedStakedETHSuite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.Client
	origin    *BaseExchangeHandler
}

func (suite *WrappedStakedETHSuite) SetupSuite() {
	suite.addresses = ContractAddresses{
		"WSTETH/STETH": "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0",
	}
}
func (suite *WrappedStakedETHSuite) TearDownSuite() {
	suite.addresses = nil
}

func (suite *WrappedStakedETHSuite) SetupTest() {
	suite.client = &ethereumMocks.Client{}
	o, err := NewWrappedStakedETH(suite.client, suite.addresses, []int64{0, 10, 20})
	suite.NoError(err)
	suite.origin = NewBaseExchangeHandler(o, nil)
}

func (suite *WrappedStakedETHSuite) TearDownTest() {
	suite.client = nil
	suite.origin = nil
}

func (suite *WrappedStakedETHSuite) Origin() Handler {
	return suite.origin
}

func TestWrappedStakedETHSuite(t *testing.T) {
	suite.Run(t, new(WrappedStakedETHSuite))
}

func (suite *WrappedStakedETHSuite) TestSuccessResponse() {
	resp := [][]byte{
		common.BigToHash(big.NewInt(0.94 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.98 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.99 * 1e18)).Bytes(),
	}
	suite.client.On(
		"CallBlocks",
		mock.Anything,
		ethereum.Call{
			Address: ethereum.HexToAddress("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
			Data:    ethereum.HexToBytes("0x035faf82"),
		},
		[]int64{0, 10, 20},
	).Return(resp, nil).Once()

	pair := Pair{Base: "WSTETH", Quote: "STETH"}

	results1 := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.97, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	suite.client.AssertNumberOfCalls(suite.T(), "CallBlocks", 1)
}

func (suite *WrappedStakedETHSuite) TestSuccessResponse_Inverted() {
	resp := [][]byte{
		common.BigToHash(big.NewInt(0.94 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.98 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.99 * 1e18)).Bytes(),
	}
	suite.client.On(
		"CallBlocks",
		mock.Anything,
		ethereum.Call{
			Address: ethereum.HexToAddress("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"),
			Data:    ethereum.HexToBytes("0x9576a0c8"),
		},
		[]int64{0, 10, 20},
	).Return(resp, nil).Once()

	pair := Pair{Base: "STETH", Quote: "WSTETH"}

	results1 := suite.origin.Fetch([]Pair{pair})

	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.97, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	suite.client.AssertNumberOfCalls(suite.T(), "CallBlocks", 1)
}

func (suite *WrappedStakedETHSuite) TestFailOnWrongPair() {
	pair := Pair{Base: "x", Quote: "y"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Require().EqualError(cr[0].Error, "failed to get contract address for pair: x/y")
}
