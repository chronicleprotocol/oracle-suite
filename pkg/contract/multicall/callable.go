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

func AggregateCallables(client rpc.RPC, calls ...contract.Callable) *AggregatedCallables {
	return &AggregatedCallables{client: client, calls: calls}
}

func (a *AggregatedCallables) AllowFail() *AggregatedCallables {
	a.allowFail = true
	return a
}

func (a *AggregatedCallables) Address() types.Address {
	if len(a.calls) == 1 {
		return a.calls[0].Address()
	}
	return multicallAddress
}

func (a *AggregatedCallables) CallData() ([]byte, error) {
	callsArg := make([]Call, 0, len(a.calls))
	for i, c := range a.calls {
		address := c.Address()
		if address == types.ZeroAddress {
			continue
		}
		callData, err := c.CallData()
		if err != nil {
			return nil, fmt.Errorf("unable to encode call %d: %w", i, err)
		}
		callsArg = append(callsArg, Call{
			Target:    c.Address(),
			CallData:  callData,
			AllowFail: a.allowFail,
		})
	}
	if len(callsArg) == 0 {
		return nil, nil
	}
	if len(callsArg) == 1 {
		return callsArg[0].CallData, nil
	}
	return multicallAbi.Methods["aggregate3"].EncodeArgs(callsArg)
}

func (a *AggregatedCallables) DecodeTo(bytes []byte, res any) error {
	if len(a.calls) == 0 {
		return nil
	}
	if len(a.calls) == 1 {
		if dec, ok := a.calls[0].(contract.Decoder); ok {
			return dec.DecodeTo(bytes, res)
		}
		return nil
	}
	var results []Result
	if bytes != nil {
		if err := multicallAbi.Methods["aggregate3"].DecodeValue(bytes, &results); err != nil {
			return fmt.Errorf("unable to decode result: %w", err)
		}
	}
	rv := reflect.ValueOf(res)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("result must be a slice")
	}
	resIdx := 0
	for i, call := range a.calls {
		if i >= rv.Len() {
			continue
		}
		if dec, ok := call.(contract.Decoder); ok {
			if call.Address() == types.ZeroAddress {
				if err := dec.DecodeTo(nil, rv.Index(i).Addr().Interface()); err != nil {
					return fmt.Errorf("unable to decode element %d: %w", i, err)
				}
				continue
			}
			if resIdx >= len(results) {
				continue
			}
			if !rv.Index(i).CanAddr() {
				return fmt.Errorf("unable to decode element %d: not addressable", i)
			}
			if err := dec.DecodeTo(results[resIdx].Data, rv.Index(i).Addr().Interface()); err != nil {
				return fmt.Errorf("unable to decode element %d: %w", i, err)
			}
		}
		resIdx++
	}
	return nil
}

func (a *AggregatedCallables) Client() rpc.RPC {
	return a.client
}

func (a *AggregatedCallables) Call(ctx context.Context, number types.BlockNumber, res any) error {
	callData, err := a.CallData()
	if err != nil {
		return err
	}
	if callData == nil {
		return a.DecodeTo(nil, res)
	}
	data, _, err := a.client.Call(
		ctx,
		types.Call{To: &multicallAddress, Input: callData},
		number,
	)
	if err != nil {
		return fmt.Errorf("call failed: %w", err)
	}
	return a.DecodeTo(data, res)
}

func (a *AggregatedCallables) Gas(ctx context.Context, number types.BlockNumber) (uint64, error) {
	callData, err := a.CallData()
	if err != nil {
		return 0, err
	}
	if callData == nil {
		return 0, nil
	}
	return a.client.EstimateGas(ctx, types.Call{To: &multicallAddress, Input: callData}, number)
}

func (a *AggregatedCallables) SendTransaction(ctx context.Context) (*types.Hash, *types.Transaction, error) {
	callData, err := a.CallData()
	if err != nil {
		return nil, nil, err
	}
	return a.client.SendTransaction(ctx, types.Transaction{Call: types.Call{To: &multicallAddress, Input: callData}})
}
