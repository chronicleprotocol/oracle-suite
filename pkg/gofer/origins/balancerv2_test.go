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
	}
	suite.client = &ethereumMocks.Client{}
}
func (suite *BalancerV2Suite) TearDownSuite() {
	suite.addresses = nil
	suite.client = nil
}

func (suite *BalancerV2Suite) SetupTest() {
	balancerV2Finance, err := NewBalancerV2(suite.client, suite.addresses, []int64{0, 10, 20})
	suite.NoError(err)
	suite.origin = NewBaseExchangeHandler(balancerV2Finance, nil)
}

func (suite *BalancerV2Suite) TearDownTest() {
	suite.origin = nil
}

func (suite *BalancerV2Suite) Origin() Handler {
	return suite.origin
}

func TestBalancerV2Suite(t *testing.T) {
	suite.Run(t, new(BalancerV2Suite))
}

func (suite *BalancerV2Suite) TestSuccessResponse() {
	suite.client.On("BlockNumber", mock.Anything).Return(big.NewInt(100), nil)

	// resp1 := common.BigToHash(big.NewInt(0.97 * 1e18))
	// suite.client.On("Call", mock.Anything, ethereum.Call{
	// 	Address: ethereum.HexToAddress("0x32296969Ef14EB0c6d29669C550D4a0449130230"),
	// 	Data:    ethereum.HexToBytes("0xb10be7390000000000000000000000000000000000000000000000000000000000000000"),
	// }).Return(resp1.Bytes(), nil).Once()

	resp2 := common.BigToHash(big.NewInt(0.98 * 1e18))
	suite.client.On("Call", mock.Anything, ethereum.Call{
		Address: ethereum.HexToAddress("0x32296969Ef14EB0c6d29669C550D4a0449130230"),
		Data:    ethereum.HexToBytes("0xb10be7390000000000000000000000000000000000000000000000000000000000000000"),
	}).Return(resp2.Bytes(), nil).Once()

	// resp3 := common.BigToHash(big.NewInt(0.99 * 1e18))
	// suite.client.On("Call", mock.Anything, ethereum.Call{
	// 	Address: ethereum.HexToAddress("0x32296969Ef14EB0c6d29669C550D4a0449130230"),
	// 	Data:    ethereum.HexToBytes("0xb10be7390000000000000000000000000000000000000000000000000000000000000000"),
	// }).Return(resp3.Bytes(), nil).Once()

	pair := Pair{Base: "STETH", Quote: "WETH"}

	results1 := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.98, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	results2 := suite.origin.Fetch([]Pair{pair.Inverse()})
	suite.Require().Error(results2[0].Error)
}

func (suite *BalancerV2Suite) TestFailOnWrongPair() {
	pair := Pair{Base: "x", Quote: "y"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Require().EqualError(cr[0].Error, "failed to get contract address for pair: x/y")
}
