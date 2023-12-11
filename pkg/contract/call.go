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
	// Address returns the address of the contract to call.
	Address() types.Address

	// Client returns the RPC client to use when performing the call.
	Client() rpc.RPC

	// Call executes the call and decodes the result into res.
	Call(ctx context.Context, number types.BlockNumber, res any) error

	// Gas returns the estimated gas usage of the call.
	Gas(ctx context.Context, number types.BlockNumber) (uint64, error)
}

// TypedCaller can perform a call to a contract and decode the result.
type TypedCaller[T any] interface {
	// Address returns the address of the contract to call.
	Address() types.Address

	// Client returns the RPC client to use when performing the call.
	Client() rpc.RPC

	// Call executes the call and decodes the result into res.
	Call(ctx context.Context, number types.BlockNumber) (T, error)

	// Gas returns the estimated gas usage of the call.
	Gas(ctx context.Context, number types.BlockNumber) (uint64, error)
}

// Transactor can send a transaction to a contract.
type Transactor interface {
	// Address returns the address of the contract to send a transaction to.
	Address() types.Address

	// Client returns the RPC client to use when sending the transaction.
	Client() rpc.RPC

	// SendTransaction sends a call as a transaction.
	SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error)

	// Gas returns the estimated gas usage of the transaction.
	Gas(ctx context.Context, number types.BlockNumber) (uint64, error)
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

	// Encoder is an optional encoder that will be used to encode the
	// arguments of the contract call.
	Encoder func() ([]byte, error)

	// Decoder is an optional decoder that will be used to decode the result
	// returned by the contract call.
	Decoder func([]byte, any) error

	// ErrorDecoder is an optional decoder that will be used to decode the
	// error returned by the contract call.
	ErrorDecoder func(err error) error
}

// Call is a contract call.
//
// Using this type instead of performing the call directly allows to choose
// if the call should be executed immediately or passed as an argument to
// another function.
type Call struct {
	client       rpc.RPC
	address      types.Address
	encoder      func() ([]byte, error)
	decoder      func([]byte, any) error
	errorDecoder func(err error) error
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

// NewCallEncoder creates a new encoder for the given method and arguments.
//
// It can be used to create a CallOpts.Encoder.
func NewCallEncoder(method *goethABI.Method, args ...any) func() ([]byte, error) {
	return func() ([]byte, error) {
		return method.EncodeArgs(args...)
	}
}

// NewCallDecoder creates a new decoder for the given method and result.
//
// It can be used to create a CallOpts.Decoder.
func NewCallDecoder(method *goethABI.Method) func(data []byte, res any) error {
	return func(data []byte, res any) error {
		switch method.Outputs().Size() {
		case 0:
			return nil
		case 1:
			if err := method.DecodeValues(data, res); err != nil {
				return fmt.Errorf("failed to decode result: %w", err)
			}
		default:
			if err := method.DecodeValue(data, res); err != nil {
				return fmt.Errorf("%s failed: %w", method.Name(), err)
			}
		}
		return nil
	}
}

// NewContractErrorDecoder creates a new decoder that can handle errors
// defined in the given contract.
//
// It can be used to create a CallOpts.ErrorDecoder.
func NewContractErrorDecoder(contract *goethABI.Contract) func(err error) error {
	return contract.HandleError
}

// NewCall creates a new Call instance.
func NewCall(opts CallOpts) *Call {
	if opts.Encoder == nil {
		opts.Encoder = defEncoder
	}
	if opts.Decoder == nil {
		opts.Decoder = defDecoder
	}
	if opts.ErrorDecoder == nil {
		opts.ErrorDecoder = defErrorDecoder
	}
	return &Call{
		client:       opts.Client,
		address:      opts.Address,
		encoder:      opts.Encoder,
		decoder:      opts.Decoder,
		errorDecoder: opts.ErrorDecoder,
	}
}

// NewTransactableCall creates a new TransactableCall instance.
func NewTransactableCall(opts CallOpts) *TransactableCall {
	return &TransactableCall{call: *NewCall(opts)}
}

// NewTypedCall creates a new TypedCall instance.
func NewTypedCall[T any](opts CallOpts) *TypedCall[T] {
	return &TypedCall[T]{call: *NewCall(opts)}
}

// NewTypedTransactableCall creates a new TypedTransactableCall instance.
func NewTypedTransactableCall[T any](opts CallOpts) *TypedTransactableCall[T] {
	return &TypedTransactableCall[T]{transactableCall: TransactableCall{call: *NewCall(opts)}}
}

// Client implements the Caller, TypedCaller, and Transactor interface.
func (c *Call) Client() rpc.RPC {
	return c.client
}

// Address implements the Callable, Caller, TypedCaller, and Transactor interface.
func (c *Call) Address() types.Address {
	return c.address
}

// CallData implements the Callable interface.
func (c *Call) CallData() ([]byte, error) {
	return c.encoder()
}

// DecodeTo implements the Decoder interface.
func (c *Call) DecodeTo(data []byte, res any) error {
	if res == nil {
		return nil
	}
	if c.decoder != nil {
		return c.decoder(data, res)
	}
	return nil
}

// Call executes the call and decodes the result into res.
func (c *Call) Call(ctx context.Context, number types.BlockNumber, res any) error {
	callData, err := c.encoder()
	if err != nil {
		return err
	}
	data, _, err := c.client.Call(ctx, types.Call{To: &c.address, Input: callData}, number)
	if err != nil {
		return c.errorDecoder(err)
	}
	return c.DecodeTo(data, res)
}

// Gas returns the estimated gas usage of the call.
func (c *Call) Gas(ctx context.Context, number types.BlockNumber) (uint64, error) {
	callData, err := c.encoder()
	if err != nil {
		return 0, err
	}
	gas, err := c.client.EstimateGas(ctx, types.Call{To: &c.address, Input: callData}, number)
	if err != nil {
		return 0, c.errorDecoder(err)
	}
	return gas, nil
}

// SendTransaction sends a call as a transaction.
func (t *TransactableCall) SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error) {
	callData, err := t.encoder()
	if err != nil {
		return nil, nil, err
	}
	txHash, txCpy, err := t.client.SendTransaction(
		ctx,
		types.Transaction{Call: types.Call{To: &t.address, Input: callData}},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send transaction: %w", err)
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

var defEncoder = func() ([]byte, error) { return nil, nil }
var defDecoder = func([]byte, any) error { return nil }
var defErrorDecoder = func(err error) error {
	if err := goethABI.Panic.HandleError(err); err != nil {
		return err
	}
	if err := goethABI.Revert.HandleError(err); err != nil {
		return err
	}
	return err
}

// Private aliases to allow embedding without exposing the methods:

type call = Call
type transactableCall = TransactableCall
