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
	pokeStorageSlot   = 4 // storage slot where the poke data is stored in Scribe and OpScribe contracts
	opPokeStorageSlot = 8 // storage slot where the optimistic poke data is stored in OpScribe contract
)

var (
	abi = goethABI.NewABI()

	abiMedian       *goethABI.Contract
	abiScribe       *goethABI.Contract
	abiOpScribe     *goethABI.Contract
	abiWatRegistry  *goethABI.Contract
	abiFeedRegistry *goethABI.Contract
	abiChainlog     *goethABI.Contract
)

func init() {
	// The bytes32_string type is a null-terminated string represented as a bytes32.
	abi.Types["bytes32_string"] = bytes32StringType{}

	// The uint256_feedBloom type is a 256-bit bitfield, where n-th bit is set if
	// the feed with first byte of its address equal to n is present in the bloom
	// filter.
	abi.Types["uint256_feedBloom"] = uint256FeedBloomType{}

	// Types for Scribe and Optimistic Scribe.
	abi.Types["PokeData"] = abi.MustParseType("(uint128 val, uint32 age)")
	abi.Types["SchnorrData"] = abi.MustParseType("(bytes32 signature, address commitment, bytes feedIDs)")
	abi.Types["ECDSAData"] = abi.MustParseType("(uint8 v, bytes32 r, bytes32 s)")

	abiMedian = abi.MustParseSignatures(
		`age()(uint256 age)`,
		`wat()(bytes32_string wat)`,
		`bar()(uint8 bar)`,
		`poke(
			uint256[] calldata val_, 
			uint256[] calldata age_, 
			uint8[] calldata v, 
			bytes32[] calldata r, 
			bytes32[] calldata s
		)`,
	)

	abiScribe = abi.MustParseSignatures(
		`error StaleMessage(uint32 givenAge, uint32 currentAge)`,
		`error FutureMessage(uint32 givenAge, uint32 currentTimestamp)`,
		`error BarNotReached(uint8 numberSigners, uint8 bar)`,
		`error SignerNotFeed(address signer)`,
		`error SignersNotOrdered()`,
		`error SchnorrSignatureInvalid()`,

		`wat()(bytes32_string wat)`,
		`bar()(uint8 bar)`,
		`feeds()(address[] feeds)`,
		`poke(PokeData pokeData, SchnorrData schnorrData)`,
		`poke_optimized_7136211(PokeData pokeData, SchnorrData schnorrData)`,
	)

	abiOpScribe = abi.MustParseSignatures(
		`error StaleMessage(uint32 givenAge, uint32 currentAge)`,
		`error FutureMessage(uint32 givenAge, uint32 currentTimestamp)`,
		`error BarNotReached(uint8 numberSigners, uint8 bar)`,
		`error SignerNotFeed(address signer)`,
		`error SignersNotOrdered()`,
		`error SchnorrSignatureInvalid()`,
		`error InChallengePeriod()`,
		`error NoOpPokeToChallenge()`,
		`error SchnorrDataMismatch(uint160 gotHash, uint160 wantHash)`,

		`wat()(bytes32_string wat)`,
		`bar()(uint8 bar)`,
		`opChallengePeriod()(uint16 opChallengePeriod)`,
		`feeds()(address[] feeds)`,
		`opPoke(PokeData pokeData, SchnorrData schnorrData, ECDSAData ecdsaData)`,
		`opPoke_optimized_397084999(PokeData pokeData, SchnorrData schnorrData, ECDSAData ecdsaData)`,
	)

	abiFeedRegistry = abi.MustParseSignatures(
		`event FeedLifted(address indexed caller, address indexed feed)`,
		`event FeedDropped(address indexed caller, address indexed feed)`,

		`function feeds() external view returns (address[] memory)`,
		`function feeds(address feed) external view returns (bool)`,
	)

	abiWatRegistry = abi.MustParseSignatures(
		`event Embraced(address indexed caller, bytes32 indexed wat)`,
		`event Abandoned(address indexed caller, bytes32 indexed wat)`,
		`event ConfigUpdated(
			address indexed caller,
			bytes32 indexed wat,
			uint8 oldBar,
			uint8 newBar,
			uint oldFeedsBloom,
			uint newFeedsBloom
    	)`,
		`event DeploymentUpdated(
			address indexed caller,
			bytes32 indexed wat,
			uint chainId,
			address oldDeployment,
			address newDeployment
    	)`,

		`wats() external view returns (bytes32_string[] memory)`,
		`exists(bytes32_string wat) external view returns (bool)`,
		`config(bytes32_string wat) external view returns (uint8 bar, uint256_feedBloom bloom)`,
		`chains(bytes32_string wat) external view returns (uint[] memory)`,
		`deployment(bytes32_string wat, uint chainId) external view returns (address)`,
	)

	abiChainlog = abi.MustParseSignatures(
		`tryGet(bytes32_string key)(bool ok, address address)`,
	)

	abiScribe.Methods["poke"] = abiScribe.Methods["poke_optimized_7136211"]
	abiOpScribe.Methods["opPoke"] = abiOpScribe.Methods["opPoke_optimized_397084999"]
	abiFeedRegistry.Methods["feeds(address)"] = abiFeedRegistry.Methods["feeds2"]
}

type PokeData struct {
	Val *bn.DecFixedPointNumber
	Age time.Time
}

type SchnorrData struct {
	Signature  *big.Int
	Commitment types.Address
	FeedIDs    FeedIDs
}

// PokeDataStruct represents the PokeData struct in the IScribe interface.
type PokeDataStruct struct {
	Val *big.Int `abi:"val"`
	Age uint32   `abi:"age"`
}

// SchnorrDataStruct represents the SchnorrData struct in the IScribe interface.
type SchnorrDataStruct struct {
	Signature  *big.Int      `abi:"signature"`
	Commitment types.Address `abi:"commitment"`
	FeedIDs    []byte        `abi:"feedIDs"`
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
	return SchnorrDataStruct{
		Signature:  s.Signature,
		Commitment: s.Commitment,
		FeedIDs:    s.FeedIDs.FeedIDs(),
	}
}

func toECDSADataStruct(s types.Signature) ECDSADataStruct {
	return ECDSADataStruct{
		V: uint8(s.V.Uint64()),
		R: s.R,
		S: s.S,
	}
}
