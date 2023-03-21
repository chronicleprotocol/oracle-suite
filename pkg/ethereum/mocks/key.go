package mocks

import (
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"
)

type Key struct {
	mock.Mock
}

func (k *Key) Address() types.Address {
	args := k.Called()
	return args.Get(0).(types.Address)
}

func (k *Key) SignHash(hash types.Hash) (*types.Signature, error) {
	args := k.Called(hash)
	return args.Get(0).(*types.Signature), args.Error(1)
}

func (k *Key) SignMessage(data []byte) (*types.Signature, error) {
	args := k.Called(data)
	return args.Get(0).(*types.Signature), args.Error(1)
}

func (k *Key) SignTransaction(tx *types.Transaction) error {
	args := k.Called(tx)
	return args.Error(0)
}

func (k *Key) VerifyHash(hash types.Hash, sig types.Signature) bool {
	args := k.Called(hash, sig)
	return args.Bool(0)
}

func (k *Key) VerifyMessage(data []byte, sig types.Signature) bool {
	args := k.Called(data, sig)
	return args.Bool(0)
}
