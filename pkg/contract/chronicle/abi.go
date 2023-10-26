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

package chronicle

import (
	"math/big"
	"time"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const (
	pokeStorageSlot   = 4
	opPokeStorageSlot = 8
)

var (
	abi = goethABI.NewABI()

	abiMedian      *goethABI.Contract
	abiScribe      *goethABI.Contract
	abiOpScribe    *goethABI.Contract
	abiWatRegistry *goethABI.Contract
	abiChainlog    *goethABI.Contract
)

func init() {
	// Types for Scribe and Optimistic Scribe
	abi.Types["PokeData"], _ = abi.ParseType("(uint128 val, uint32 age)")
	abi.Types["SchnorrData"], _ = abi.ParseType("(bytes32 signature, address commitment, bytes signersBlob)")
	abi.Types["ECDSAData"], _ = abi.ParseType("(uint8 v, bytes32 r, bytes32 s)")

	abiMedian, _ = abi.ParseSignatures(
		`age()(uint256 age)`,
		`wat()(bytes32 wat)`,
		`bar()(uint8 bar)`,
		`poke(
			uint256[] calldata val_, 
			uint256[] calldata age_, 
			uint8[] calldata v, 
			bytes32[] calldata r, 
			bytes32[] calldata s
		)`,
	)

	abiScribe, _ = abi.ParseSignatures(
		`error StaleMessage(uint32 givenAge, uint32 currentAge)`,
		`error FutureMessage(uint32 givenAge, uint32 currentTimestamp)`,
		`error BarNotReached(uint8 numberSigners, uint8 bar)`,
		`error SignerNotFeed(address signer)`,
		`error SignersNotOrdered()`,
		`error SchnorrSignatureInvalid()`,

		`wat()(bytes32 wat)`,
		`bar()(uint8 bar)`,
		`feeds()(address[] feeds, uint[] feedIndexes)`,
		`poke(PokeData pokeData, SchnorrData schnorrData)`,
		`poke_optimized_7136211(PokeData pokeData, SchnorrData schnorrData)`,
	)

	abiOpScribe, _ = abi.ParseSignatures(
		`error StaleMessage(uint32 givenAge, uint32 currentAge)`,
		`error FutureMessage(uint32 givenAge, uint32 currentTimestamp)`,
		`error BarNotReached(uint8 numberSigners, uint8 bar)`,
		`error SignerNotFeed(address signer)`,
		`error SignersNotOrdered()`,
		`error SchnorrSignatureInvalid()`,
		`error InChallengePeriod()`,
		`error NoOpPokeToChallenge()`,
		`error SchnorrDataMismatch(uint160 gotHash, uint160 wantHash)`,

		`wat()(bytes32 wat)`,
		`bar()(uint8 bar)`,
		`opChallengePeriod()(uint16 opChallengePeriod)`,
		`feeds()(address[] feeds, uint[] feedIndexes)`,
		`opPoke(PokeData pokeData, SchnorrData schnorrData, ECDSAData ecdsaData)`,
		`opPoke_optimized_397084999(PokeData pokeData, SchnorrData schnorrData, ECDSAData ecdsaData)`,
	)

	abiWatRegistry, _ = abi.ParseSignatures(
		`bar(bytes32 wat)(uint8 bar)`,
		`feeds(bytes32 wat)(address[] feeds)`,
	)

	abiChainlog, _ = abi.ParseSignatures(
		`tryGet(bytes32 key)(bool ok, address address)`,
	)

	abiScribe.Methods["poke"] = abiScribe.Methods["poke_optimized_7136211"]
	abiOpScribe.Methods["opPoke"] = abiOpScribe.Methods["opPoke_optimized_397084999"]
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

func toPokeDataStruct(p PokeData) PokeDataStruct {
	return PokeDataStruct{
		Val: p.Val.SetPrec(ScribePricePrecision).RawBigInt(),
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
