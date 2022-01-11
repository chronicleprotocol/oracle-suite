package ethereum

import (
	"context"
	"errors"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const SignatureKey = "ethereum"
const SignerKey = "ethereum_signer"

// Signer signs Ethereum log messages using Ethereum signature.
type Signer struct {
	signer ethereum.Signer
	types  []string
}

// NewSigner returns a new instance of the Signer struct.
func NewSigner(signer ethereum.Signer, types []string) *Signer {
	return &Signer{signer: signer, types: types}
}

// Start implements the Signer interface.
func (l *Signer) Start(_ context.Context) error {
	return nil
}

// Sign implements the Signer interface.
func (l *Signer) Sign(event *messages.Event) (bool, error) {
	supports := false
	for _, t := range l.types {
		if t == event.Type {
			supports = true
			break
		}
	}
	if !supports {
		return false, nil
	}
	h, ok := event.Data["hash"]
	if !ok {
		return false, errors.New("missing hash field")
	}
	s, err := l.signer.Signature(h)
	if err != nil {
		return false, err
	}
	event.Data[SignerKey] = l.signer.Address().Bytes()
	event.Signatures[SignatureKey] = s.Bytes()
	return true, nil
}
