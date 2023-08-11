package mocks

import (
	"context"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/mock"
)

type Recoverer struct {
	mock.Mock
}

func (r *Recoverer) RecoverHash(hash types.Hash, sig types.Signature) (*types.Address, error) {
	args := r.Called(hash, sig)
	return args.Get(0).(*types.Address), args.Error(1)
}

func (r *Recoverer) RecoverMessage(data []byte, sig types.Signature) (*types.Address, error) {
	args := r.Called(data, sig)
	return args.Get(0).(*types.Address), args.Error(1)
}

func (r *Recoverer) RecoverTransaction(tx *types.Transaction) (*types.Address, error) {
	args := r.Called(tx)
	return args.Get(0).(*types.Address), args.Error(1)
}

func (r *Recoverer) Supports(_ context.Context, data datapoint.Point) bool {
	return true
}

func (r *Recoverer) Recover(_ context.Context, _ string, p datapoint.Point, _ types.Signature) (*types.Address, error) {
	return types.MustAddressFromHexPtr(p.Meta["addr"].(string)), nil
}

var _ crypto.Recoverer = (*Recoverer)(nil)
