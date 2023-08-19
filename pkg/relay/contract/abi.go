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

package contract

import (
	"math/big"
	"time"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

var (
	abi         = goethABI.Default
	abiMedian   = make(map[string]*goethABI.Method)
	abiScribe   = make(map[string]*goethABI.Method)
	abiOpScribe = make(map[string]*goethABI.Method)
)

func init() {
	// Types for Scribe and Optimistic Scribe
	abi.Types["PokeData"], _ = abi.ParseType("(uint128 val, uint32 age)")
	abi.Types["SchnorrData"], _ = abi.ParseType("(bytes32 signature, address commitment, bytes signersBlob)")
	abi.Types["ECDSAData"], _ = abi.ParseType("(uint8 v, bytes32 r, bytes32 s)")

	// Median
	abiMedian["poke"], _ = abi.ParseMethod(
		`function poke(
			uint256[] calldata val_, 
			uint256[] calldata age_, 
			uint8[] calldata v, 
			bytes32[] calldata r, 
			bytes32[] calldata s
		)`,
	)
	abiMedian["age"], _ = abi.ParseMethod(`age()`)
	abiMedian["wat"], _ = abi.ParseMethod(`wat()`)
	abiMedian["bar"], _ = abi.ParseMethod(`bar()`)

	// Scribe
	abiScribe["wat"], _ = abi.ParseMethod(`wat()(bytes32)`)
	abiScribe["bar"], _ = abi.ParseMethod(`bar()(uint8)`)
	abiScribe["feeds"], _ = abi.ParseMethod(`feeds()(address[] feeds, uint[] feedIndexes)`)
	abiScribe["poke"], _ = abi.ParseMethod(
		`function poke(
			PokeData calldata pokeData, 
			SchnorrData calldata schnorrData
		)`,
	)

	// Optimistic Scribe
	abiOpScribe["wat"], _ = abi.ParseMethod(`wat()(bytes32)`)
	abiOpScribe["feeds"], _ = abi.ParseMethod(`feeds()(address[] feeds, uint[] feedIndexes)`)
	abiOpScribe["opPoke"], _ = abi.ParseMethod(
		`opPoke(
			PokeData calldata pokeData, 
			SchnorrData calldata schnorrData, 
			ECDSAData calldata ecdsaData
		)`,
	)
}

type PokeData struct {
	Val *bn.DecFixedPointNumber
	Age time.Time
}

type SchnorrData struct {
	Signature   *big.Int
	Commitment  types.Address
	SignersBlob []byte
}

func toPokeDataStruct(p PokeData) PokeDataStruct {
	return PokeDataStruct{
		Val: p.Val.RawBigInt(),
		Age: uint32(p.Age.Unix()),
	}
}

func toSchnorrDataStruct(s SchnorrData) SchnorrDataStruct {
	return SchnorrDataStruct(s)
}

func toECDSADataStruct(s types.Signature) ECDSADataStruct {
	return ECDSADataStruct{
		V: uint8(s.V.Uint64()),
		R: s.R,
		S: s.S,
	}
}

// PokeDataStruct represents the PokeData struct in the IScribe interface.
type PokeDataStruct struct {
	Val *big.Int `abi:"val"`
	Age uint32   `abi:"age"`
}

// SchnorrDataStruct represents the SchnorrData struct in the IScribe interface.
type SchnorrDataStruct struct {
	Signature   *big.Int      `abi:"signature"`
	Commitment  types.Address `abi:"commitment"`
	SignersBlob []byte        `abi:"signersBlob"`
}

// ECDSADataStruct represents the ECDSAData struct in the IScribe interface.
type ECDSADataStruct struct {
	V uint8    `abi:"v"`
	R *big.Int `abi:"r"`
	S *big.Int `abi:"s"`
}
