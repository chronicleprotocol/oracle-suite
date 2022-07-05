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

type BalancerV2Suite struct {
	suite.Suite
	addresses ContractAddresses
	client    *ethereumMocks.Client
	origin    *BaseExchangeHandler
}

func (suite *BalancerV2Suite) SetupSuite() {
	suite.addresses = ContractAddresses{
		"Ref:RETH/WETH": "0xae78736Cd615f374D3085123A210448E74Fc6393",
		"RETH/WETH":     "0x1E19CF2D73a72Ef1332C882F20534B6519Be0276",
		"STETH/WETH":    "0x32296969ef14eb0c6d29669c550d4a0449130230",
		"WETH/YFI":      "0x186084ff790c65088ba694df11758fae4943ee9e",
	}
}
func (suite *BalancerV2Suite) TearDownSuite() {
	suite.addresses = nil
}

func (suite *BalancerV2Suite) SetupTest() {
	suite.client = &ethereumMocks.Client{}
	o, err := NewBalancerV2(suite.client, suite.addresses, []int64{0, 10, 20})
	suite.NoError(err)
	suite.origin = NewBaseExchangeHandler(o, nil)
}

func (suite *BalancerV2Suite) TearDownTest() {
	suite.origin = nil
	suite.client = nil
}

func (suite *BalancerV2Suite) Origin() Handler {
	return suite.origin
}

func TestBalancerV2Suite(t *testing.T) {
	suite.Run(t, new(BalancerV2Suite))
}

func (suite *BalancerV2Suite) TestSuccessResponse() {
	resp := [][]byte{
		common.BigToHash(big.NewInt(0.94 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.98 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.99 * 1e18)).Bytes(),
	}
	suite.client.On(
		"CallBlocks",
		mock.Anything,
		ethereum.Call{
			Address: ethereum.HexToAddress("0x32296969Ef14EB0c6d29669C550D4a0449130230"),
			Data:    ethereum.HexToBytes("0xb10be7390000000000000000000000000000000000000000000000000000000000000000"),
		},
		[]int64{0, 10, 20},
	).Return(resp, nil).Once()

	pair := Pair{Base: "STETH", Quote: "WETH"}

	results1 := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.97, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	suite.client.AssertNumberOfCalls(suite.T(), "CallBlocks", 1)

	results2 := suite.origin.Fetch([]Pair{pair.Inverse()})
	suite.Require().Error(results2[0].Error)
}

func (suite *BalancerV2Suite) TestSuccessResponseWithRef() {
	resp := [][]byte{
		common.BigToHash(big.NewInt(0.94 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.98 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.99 * 1e18)).Bytes(),
	}
	suite.client.On(
		"CallBlocks",
		mock.Anything,
		ethereum.Call{
			Address: ethereum.HexToAddress("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"),
			Data:    ethereum.HexToBytes("0xb10be7390000000000000000000000000000000000000000000000000000000000000000"),
		},
		[]int64{0, 10, 20},
	).Return(resp, nil).Once()

	resp1 := [][]byte{
		common.BigToHash(big.NewInt(0.2 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.6 * 1e18)).Bytes(),
		common.BigToHash(big.NewInt(0.7 * 1e18)).Bytes(),
	}
	suite.client.On(
		"CallBlocks",
		mock.Anything,
		ethereum.Call{
			Address: ethereum.HexToAddress("0x1E19CF2D73a72Ef1332C882F20534B6519Be0276"),
			Data:    ethereum.HexToBytes("0xb867ee5a000000000000000000000000ae78736cd615f374d3085123a210448e74fc6393"),
		},
		[]int64{0, 10, 20},
	).Return(resp1, nil).Once()

	pair := Pair{Base: "RETH", Quote: "WETH"}

	results1 := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.485, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	suite.client.AssertNumberOfCalls(suite.T(), "CallBlocks", 2)

	results2 := suite.origin.Fetch([]Pair{pair.Inverse()})
	suite.Require().Error(results2[0].Error)
}

func (suite *BalancerV2Suite) TestFailOnWrongPair() {
	pair := Pair{Base: "x", Quote: "y"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Require().EqualError(cr[0].Error, "failed to get contract address for pair: x/y")
}
