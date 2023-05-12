package feed

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/defiweb/go-eth/wallet"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/data/origin"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const priceMultiplier = 1e18

type TickHandler struct {
	signer wallet.Key
}

func NewTickHandler(signer wallet.Key) *TickHandler {
	return &TickHandler{signer: signer}
}

func (t *TickHandler) Supports(point data.Point) bool {
	_, ok := point.Value.(origin.Tick)
	return ok
}

func (t *TickHandler) Handle(model string, point data.Point) (*messages.Event, error) {
	tick, ok := point.Value.(origin.Tick)
	if !ok {
		return nil, fmt.Errorf("invalid tick type: %T", point.Value)
	}
	if err := point.Validate(); err != nil {
		return nil, fmt.Errorf("invalid point: %w", err)
	}
	hash := hashTick(model, tick.Price, point.Time)
	signature, err := t.signer.SignMessage(hash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign tick: %w", err)
	}
	return &messages.Event{
		Type:        "price_tick",
		ID:          hash.Bytes(),
		Index:       hash.Bytes(),
		EventDate:   point.Time,
		MessageDate: time.Now(),
		Data: map[string][]byte{
			"val": tick.Price.BigInt().Bytes(),
			"age": bn.Int(point.Time.Unix()).BigInt().Bytes(),
			"wat": []byte(model),
		},
		Signatures: map[string]messages.EventSignature{
			"ethereum": {
				Signer:    t.signer.Address().Bytes(),
				Signature: signature.Bytes(),
			},
		},
	}, nil
}

// hashTick is an equivalent of keccak256(abi.encodePacked(val, age, wat))) in Solidity.
func hashTick(model string, price *bn.FloatNumber, time time.Time) types.Hash {
	// Price:
	val := make([]byte, 32)
	price.Mul(priceMultiplier).BigInt().FillBytes(val)

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
