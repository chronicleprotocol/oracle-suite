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

package multicall

import (
	"context"
	"fmt"
	"reflect"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
)

// AggregatedCallables is a Callable that aggregates multiple Callables into
// a single call.
//
// The AggregatedCallables will call the multicall contract to aggregate the
// calls. If only a single call is aggregated, the call will be called directly.
//
// The AggregatedCallables will decode the result of the multicall contract
// and decode the result of the individual calls if they implement the Decoder
// interface.
//
// There is a special case for Callables that returns a zero address as their
// address. In this case, the Callable will not be included in the aggregated
// call, but the AggregatedCallables will try to decode the result of the
// Callable providing nil as the call result. This is useful mostly for
// testing purposes.
type AggregatedCallables struct {
	client    rpc.RPC
	calls     []contract.Callable
	allowFail bool
}

// AggregateCallables creates a new AggregatedCallables instance.
func AggregateCallables(client rpc.RPC, calls ...contract.Callable) *AggregatedCallables {
	return &AggregatedCallables{
		client: client,
		calls:  calls,
	}
}

// AllowFail allows the aggregated call to partially fail. If a call fails,
// the rest of the calls will still be executed.
func (a *AggregatedCallables) AllowFail() *AggregatedCallables {
	a.allowFail = true
	return a
}

// Address implements the contract.Callable interface.
func (a *AggregatedCallables) Address() types.Address {
	nonZeroCalls := a.nonZeroCalls()
	if len(nonZeroCalls) == 0 {
		return types.ZeroAddress
	}
	if len(nonZeroCalls) == 1 {
		return nonZeroCalls[0].Address()
	}
	return multicallAddress
}

// CallData implements the contract.Callable interface.
func (a *AggregatedCallables) CallData() ([]byte, error) {
	nonZeroCalls := a.nonZeroCalls()
	callsArg := make([]Call, len(nonZeroCalls))
	for i, c := range nonZeroCalls {
		callData, err := c.CallData()
		if err != nil {
			return nil, fmt.Errorf("unable to encode call %d: %w", i, err)
		}
		callsArg[i] = Call{
			Target:    c.Address(),
			CallData:  callData,
			AllowFail: a.allowFail,
		}
	}
	if len(callsArg) == 0 {
		return nil, nil
	}
	if len(callsArg) == 1 {
		return callsArg[0].CallData, nil
	}
	return multicallAbi.Methods["aggregate3"].EncodeArgs(callsArg)
}

// DecodeTo implements the contract.Decoder interface.
func (a *AggregatedCallables) DecodeTo(data []byte, res any) error {
	if res == nil {
		return nil
	}
	if len(a.calls) == 0 {
		return nil
	}
	resRefl := reflect.ValueOf(res)
	for resRefl.Kind() == reflect.Ptr {
		resRefl = resRefl.Elem()
	}
	if resRefl.Kind() == reflect.Interface && resRefl.IsNil() {
		return nil
	}
	if resRefl.Kind() != reflect.Slice {
		return fmt.Errorf("result must be a slice")
	}
	if resRefl.Len() == 0 {
		return nil
	}
	var multiCallRes []Result
	if data != nil {
		if len(a.nonZeroCalls()) == 1 {
			multiCallRes = []Result{{Data: data, Success: true}}
		} else {
			if err := multicallAbi.Methods["aggregate3"].DecodeValues(data, &multiCallRes); err != nil {
				return fmt.Errorf("unable to decode result: %w", err)
			}
		}
	}
	multiCallResIdx := 0
	for i, call := range a.calls {
		if i >= resRefl.Len() {
			continue
		}
		dec, ok := call.(contract.Decoder)
		if !ok {
			continue
		}
		if call.Address() == types.ZeroAddress {
			if err := safeDecode(resRefl.Index(i), dec, nil); err != nil {
				return fmt.Errorf("unable to decode element %d: %w", i, err)
			}
		} else {
			if multiCallResIdx >= len(multiCallRes) {
				continue
			}
			if err := safeDecode(resRefl.Index(i), dec, multiCallRes[multiCallResIdx].Data); err != nil {
				return fmt.Errorf("unable to decode element %d: %w", i, err)
			}
			multiCallResIdx++
		}
	}
	return nil
}

// Client implements the contract.Caller interface.
func (a *AggregatedCallables) Client() rpc.RPC {
	return a.client
}

// Call implements the contract.Caller interface.
func (a *AggregatedCallables) Call(ctx context.Context, number types.BlockNumber, res any) error {
	callData, err := a.CallData()
	if err != nil {
		return err
	}
	if callData == nil {
		return a.DecodeTo(nil, res)
	}
	address := a.Address()
	data, _, err := a.client.Call(ctx, types.Call{To: &address, Input: callData}, number)
	if err != nil {
		return fmt.Errorf("multicall failed: %w", err)
	}
	return a.DecodeTo(data, res)
}

// Gas implements the contract.Caller interface.
func (a *AggregatedCallables) Gas(ctx context.Context, number types.BlockNumber) (uint64, error) {
	callData, err := a.CallData()
	if err != nil {
		return 0, err
	}
	if callData == nil {
		return 0, nil
	}
	address := a.Address()
	return a.client.EstimateGas(ctx, types.Call{To: &address, Input: callData}, number)
}

// SendTransaction implements the contract.Transactor interface.
func (a *AggregatedCallables) SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error) {
	callData, err := a.CallData()
	if err != nil {
		return nil, nil, err
	}
	address := a.Address()
	return a.client.SendTransaction(ctx, types.Transaction{Call: types.Call{To: &address, Input: callData}})
}

func (a *AggregatedCallables) nonZeroCalls() []contract.Callable {
	var nonZeroCalls []contract.Callable
	for _, c := range a.calls {
		if c.Address() != types.ZeroAddress {
			nonZeroCalls = append(nonZeroCalls, c)
		}
	}
	return nonZeroCalls
}

func safeDecode(v reflect.Value, dec contract.Decoder, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	if v.Kind() == reflect.Ptr && v.IsNil() {
		if !v.CanSet() {
			return fmt.Errorf("unable to decode value to %s", v.Type())
		}
		v.Set(reflect.New(v.Type().Elem()))
	}
	return dec.DecodeTo(data, v.Interface())
}
