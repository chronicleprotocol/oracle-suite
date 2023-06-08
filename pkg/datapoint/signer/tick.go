package signer

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type Tick struct {
	signer    wallet.Key
	recoverer crypto.Recoverer
}

func NewTick(signer wallet.Key, recoverer crypto.Recoverer) *Tick {
	return &Tick{
		signer:    signer,
		recoverer: recoverer,
	}
}

func (t Tick) Supports(_ context.Context, data datapoint.Point) bool {
	_, ok := data.Value.(value.Tick)
	return ok
}

func (t Tick) Sign(_ context.Context, model string, data datapoint.Point) (*types.Signature, error) {
	return t.signer.SignMessage(hashTick(model, data.Value.(value.Tick).Price, data.Time).Bytes())
}

func (t Tick) Recover(_ context.Context, model string, data datapoint.Point, signature types.Signature) (*types.Address, error) {
	return t.recoverer.RecoverMessage(
		hashTick(model, data.Value.(value.Tick).Price, data.Time).Bytes(),
		signature,
	)
}

// hashTick is an equivalent of keccak256(abi.encodePacked(val, age, wat))) in Solidity.
func hashTick(model string, price *bn.FloatNumber, time time.Time) types.Hash {
	// Price (val):
	val := make([]byte, 32)
	price.Mul(value.RealNumberPrecision).BigInt().FillBytes(val)

	// Time (age):
	age := make([]byte, 32)
	binary.BigEndian.PutUint64(age[24:], uint64(time.Unix()))

	// Asset name (wat):
	wat := make([]byte, 32)
	copy(wat, model)

	// Hash:
	hash := make([]byte, 96)
	copy(hash[0:32], val)
	copy(hash[32:64], age)
	copy(hash[64:96], wat)
	return crypto.Keccak256(hash)
}
