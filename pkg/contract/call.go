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
	"context"
	"fmt"

	goethABI "github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

// Callable provides the data required to call a contract.
type Callable interface {
	// Address returns the address of the contract to call.
	Address() types.Address

	// CallData returns the encoded call data.
	CallData() ([]byte, error)
}

// Decoder decodes the data returned by the contract call.
type Decoder interface {
	DecodeTo([]byte, any) error
}

// TypedDecoder decodes the data returned by the contract call.
type TypedDecoder[T any] interface {
	Decode([]byte) (T, error)
}

// Caller can perform a call to a contract and decode the result.
type Caller interface {
	Address() types.Address
	Client() rpc.RPC
	Call(ctx context.Context, number types.BlockNumber, res any) error
	Gas(ctx context.Context, number types.BlockNumber) (uint64, error)
}

// TypedCaller can perform a call to a contract and decode the result.
type TypedCaller[T any] interface {
	Address() types.Address
	Client() rpc.RPC
	Call(ctx context.Context, number types.BlockNumber) (T, error)
	Gas(ctx context.Context, number types.BlockNumber) (uint64, error)
}

// Transactor can send a transaction to a contract.
type Transactor interface {
	Address() types.Address
	Client() rpc.RPC
	SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error)
}

type DecodeCallable interface {
	Callable
	Decoder
}

// SelfCaller is a Callable that can perform a call by itself.
type SelfCaller interface {
	Callable
	Caller
	Decoder
}

// SelfTransactableCaller is a Callable that can perform a call or send a
// transaction by itself.
type SelfTransactableCaller interface {
	Callable
	Caller
	Transactor
	Decoder
}

// TypedSelfCaller is a Callable that can perform a call by itself.
type TypedSelfCaller[T any] interface {
	Callable
	TypedCaller[T]
	TypedDecoder[T]
	Decoder
}

// TypedSelfTransactableCaller is a Callable that can perform a call or send a
// transaction by itself.
type TypedSelfTransactableCaller[T any] interface {
	Callable
	TypedCaller[T]
	Transactor
	Decoder
}

// CallOpts are the options for New*Call functions.
type CallOpts struct {
	// Client is the RPC client to use when performing the call or sending
	// the transaction.
	Client rpc.RPC

	// Address is the address of the contract to call or send a transaction to.
	Address types.Address

	// Method is the ABI method used to encode the call data and decode the
	// result.
	Method *goethABI.Method

	// Arguments are the arguments of the contract call or transaction.
	Arguments []any

	// Decoder is an optional decoder that will be used to decode the result
	// returned by the contract call.
	Decoder func(*goethABI.Method, []byte, any) error
}

// Call is a contract call.
//
// Using this type instead of performing the call directly allows to choose
// if the call should be executed immediately or passed as an argument to
// another function.
type Call struct {
	client  rpc.RPC
	address types.Address
	method  *goethABI.Method
	args    []any
	decoder func(*goethABI.Method, []byte, any) error
}

// TransactableCall works like Call but can be also used to send a transaction.
type TransactableCall struct {
	call
}

// TypedCall is a Call with a typed result.
type TypedCall[T any] struct {
	call
}

// TypedTransactableCall is a TransactableCall with a typed result.
type TypedTransactableCall[T any] struct {
	transactableCall
}

// NewCall creates a new Call instance.
func NewCall(opts CallOpts) *Call {
	return &Call{
		client:  opts.Client,
		address: opts.Address,
		method:  opts.Method,
		args:    opts.Arguments,
		decoder: opts.Decoder,
	}
}

// NewTransactableCall creates a new TransactableCall instance.
func NewTransactableCall(opts CallOpts) *TransactableCall {
	return &TransactableCall{
		call: Call{
			client:  opts.Client,
			address: opts.Address,
			method:  opts.Method,
			args:    opts.Arguments,
			decoder: opts.Decoder,
		},
	}
}

// NewTypedCall creates a new TypedCall instance.
func NewTypedCall[T any](opts CallOpts) *TypedCall[T] {
	return &TypedCall[T]{
		call: Call{
			client:  opts.Client,
			address: opts.Address,
			method:  opts.Method,
			args:    opts.Arguments,
			decoder: opts.Decoder,
		},
	}
}

// NewTypedTransactableCall creates a new TypedTransactableCall instance.
func NewTypedTransactableCall[T any](opts CallOpts) *TypedTransactableCall[T] {
	return &TypedTransactableCall[T]{
		transactableCall: TransactableCall{
			call: Call{
				client:  opts.Client,
				address: opts.Address,
				method:  opts.Method,
				args:    opts.Arguments,
				decoder: opts.Decoder,
			},
		},
	}
}

// Client implements the HasClient interface.
func (c *Call) Client() rpc.RPC {
	return c.client
}

// Address implements the Callable interface.
func (c *Call) Address() types.Address {
	return c.address
}

// CallData implements the Callable interface.
func (c *Call) CallData() ([]byte, error) {
	return c.method.EncodeArgs(c.args...)
}

// DecodeTo implements the Decoder interface.
func (c *Call) DecodeTo(data []byte, res any) error {
	if res == nil {
		return nil
	}
	if c.decoder != nil {
		return c.decoder(c.method, data, res)
	}
	switch c.method.Outputs().Size() {
	case 0:
		return nil
	case 1:
		if err := c.method.DecodeValues(data, res); err != nil {
			return fmt.Errorf("%s failed: %w", c.method.Name(), err)
		}
	default:
		if err := c.method.DecodeValue(data, res); err != nil {
			return fmt.Errorf("%s failed: %w", c.method.Name(), err)
		}
	}
	return nil
}

// Call executes the call and decodes the result into res.
func (c *Call) Call(ctx context.Context, number types.BlockNumber, res any) error {
	calldata, err := c.method.EncodeArgs(c.args...)
	if err != nil {
		return fmt.Errorf("%s failed: %w", c.method.Name(), err)
	}
	data, _, err := c.client.Call(
		ctx,
		types.Call{To: &c.address, Input: calldata},
		number,
	)
	if err != nil {
		return fmt.Errorf("%s failed: %w", c.method.Name(), err)
	}
	if c.method.Outputs().Size() == 0 {
		return nil
	}
	return c.DecodeTo(data, res)
}

// Gas returns the estimated gas usage of the call.
func (c *Call) Gas(ctx context.Context, number types.BlockNumber) (uint64, error) {
	calldata, err := c.method.EncodeArgs(c.args...)
	if err != nil {
		return 0, fmt.Errorf("%s failed: %w", c.method.Name(), err)
	}
	return c.client.EstimateGas(ctx, types.Call{To: &c.address, Input: calldata}, number)
}

// SendTransaction sends a call as a transaction.
func (t *TransactableCall) SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error) {
	calldata, err := t.method.EncodeArgs(t.args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s failed: %w", t.method.Name(), err)
	}
	txHash, txCpy, err := t.client.SendTransaction(
		ctx,
		types.Transaction{Call: types.Call{To: &t.address, Input: calldata}},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s failed: %w", t.method.Name(), err)
	}
	return txHash, txCpy, nil
}

// Decode decodes the result of the call.
func (t *TypedCall[T]) Decode(data []byte) (T, error) {
	var res T
	if err := t.call.DecodeTo(data, &res); err != nil {
		return res, err
	}
	return res, nil
}

// Call executes the call and returns the decoded result.
func (t *TypedCall[T]) Call(ctx context.Context, number types.BlockNumber) (T, error) {
	var res T
	if err := t.call.Call(ctx, number, &res); err != nil {
		return res, err
	}
	return res, nil
}

// Decode decodes the result of the call.
func (t *TypedTransactableCall[T]) Decode(data []byte) (T, error) {
	var res T
	if err := t.call.DecodeTo(data, &res); err != nil {
		return res, err
	}
	return res, nil
}

// Call executes the call and returns the decoded result.
func (t *TypedTransactableCall[T]) Call(ctx context.Context, number types.BlockNumber) (T, error) {
	var res T
	if err := t.call.Call(ctx, number, &res); err != nil {
		return res, err
	}
	return res, nil
}

// Create private aliases to allow embedding without exposing the methods:

type call = Call
type transactableCall = TransactableCall
