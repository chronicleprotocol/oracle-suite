package signer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Hash for the AAABBB asset pair, with the price set to 42 and the age to 1605371361:
var priceHash = "0xc678b27c20ef30b95452d8d61f8f3916899717692d8a01c595971035b25a00ff"

func TestTick_Supports(t *testing.T) {
	t.Run("supported data point", func(t *testing.T) {
		k := &mocks.Key{}
		r := &mocks.Recoverer{}
		s := NewTick(k, r)
		assert.True(t, s.Supports(context.Background(), datapoint.Point{Value: value.Tick{}}))
	})
	t.Run("unsupported data point", func(t *testing.T) {
		k := &mocks.Key{}
		r := &mocks.Recoverer{}
		s := NewTick(k, r)
		assert.False(t, s.Supports(context.Background(), datapoint.Point{Value: value.StaticValue{}}))
	})
}

func TestTick_Sign(t *testing.T) {
	k := &mocks.Key{}
	r := &mocks.Recoverer{}
	s := NewTick(k, r)

	expSig := types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65))
	k.On("SignMessage", hexutil.MustHexToBytes(priceHash)).Return(expSig, nil).Once()

	retSig, err := s.Sign(context.Background(), "AAABBB", datapoint.Point{
		Value: value.Tick{
			Pair:      value.Pair{Base: "AAA", Quote: "BBB"},
			Price:     bn.Float(42),
			Volume24h: bn.Float(0),
		},
		Time:      time.Unix(1605371361, 0),
		SubPoints: nil,
		Meta:      nil,
		Error:     nil,
	})
	require.NoError(t, err)

	assert.Equal(t, *expSig, *retSig)
}

func TestTick_Recover(t *testing.T) {
	k := &mocks.Key{}
	r := &mocks.Recoverer{}
	s := NewTick(k, r)

	msgSig := types.MustSignatureFromBytesPtr(bytes.Repeat([]byte{0xAA}, 65))
	expAddr := types.MustAddressFromHexPtr("0x1234567890123456789012345678901234567890")
	r.On("RecoverMessage", hexutil.MustHexToBytes(priceHash), *msgSig).Return(expAddr, nil).Once()

	retAddr, err := s.Recover(context.Background(), "AAABBB", datapoint.Point{
		Value: value.Tick{
			Pair:      value.Pair{Base: "AAA", Quote: "BBB"},
			Price:     bn.Float(42),
			Volume24h: bn.Float(0),
		},
		Time:      time.Unix(1605371361, 0),
		SubPoints: nil,
		Meta:      nil,
		Error:     nil,
	}, *msgSig)
	require.NoError(t, err)

	assert.Equal(t, *expAddr, *retAddr)
}

func TestHashTick(t *testing.T) {
	assert.Equal(t, priceHash, hashTick("AAABBB", bn.Float(42), time.Unix(1605371361, 0)).String())
}
