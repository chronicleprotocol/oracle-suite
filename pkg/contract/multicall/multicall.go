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

package multicall

import (
	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
)

var (
	multicallAddress = types.MustAddressFromHex("0xcA11bde05977b3631167028862bE2a173976CA11")
	multicallAbi     = abi.MustParseSignatures(
		`struct Call3{address target; bool allowFailure; bytes callData;}`,
		`struct Result{bool success; bytes returnData;}`,
		`function aggregate3(Call3[] calldata calls) public payable returns (Result[] memory returnData)`,
	)
)

type Call struct {
	Target    types.Address `abi:"target"`
	CallData  []byte        `abi:"callData"`
	AllowFail bool          `abi:"allowFailure"`
}

type Result struct {
	Success bool   `abi:"success"`
	Data    []byte `abi:"returnData"`
}

type Multicall struct {
	client rpc.RPC
}

func NewMulticall(client rpc.RPC) *Multicall {
	return &Multicall{client: client}
}

func (m *Multicall) Aggregate3(calls []Call) contract.TypedSelfTransactableCaller[[]Result] {
	method := multicallAbi.Methods["aggregate3"]
	return contract.NewTypedTransactableCall[[]Result](
		contract.CallOpts{
			Client:  m.client,
			Address: multicallAddress,
			Encoder: contract.NewCallEncoder(method, calls),
			Decoder: contract.NewCallDecoder(method),
		},
	)
}
