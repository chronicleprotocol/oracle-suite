package ethereum

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/geth"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestSigner_IgnoreUnsupportedType(t *testing.T) {
	msg := &messages.Event{Type: "foo"}
	signer := NewSigner(geth.NewSigner(nil), []string{"bar"})

	// If message is of different type, signer should do nothing:
	ok, err := signer.Sign(msg)
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestSigner_MissingHashField(t *testing.T) {
	msg := &messages.Event{Type: "foo"}
	signer := NewSigner(geth.NewSigner(nil), []string{"foo"})

	// If hash field is missing, an error must be returned:
	ok, err := signer.Sign(msg)
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestSigner_Sign(t *testing.T) {
	address := common.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	account, err := geth.NewAccount("./keystore", "test123", address)
	require.NoError(t, err)
	gethSigner := geth.NewSigner(account)
	msg := &messages.Event{Type: "foo", Data: map[string][]byte{"hash": common.HexToHash("f76b84eff86432f629ab567880256b50c8eb31cafaec58c5edb24d9b4c246470").Bytes()}}
	signer := NewSigner(gethSigner, []string{"foo"})

	ok, err := signer.Sign(msg)
	assert.True(t, ok)
	assert.NoError(t, err)

	// Verify if address in signer field is correct:
	assert.Equal(t, msg.Data[SignerKey], address.Bytes())

	// Verify signature:
	recovered, err := gethSigner.Recover(ethereum.SignatureFromBytes(msg.Signatures[SignatureKey]), msg.Data["hash"])
	require.NoError(t, err)
	assert.Equal(t, address, *recovered)
}
