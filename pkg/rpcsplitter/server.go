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

package rpcsplitter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	gethRPC "github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/rpcsplitter/types"
)

const LoggerTag = "RPCSPLITTER"

const defaultTotalTimeout = 10 * time.Second
const defaultGracefulTimeout = 1 * time.Second

type caller interface {
	CallContext(ctx context.Context, result any, method string, args ...any) error
}

// server is an RPC proxy server. It merges multiple RPC endpoints into one.
type server struct {
	rpc *gethRPC.Server // rpc is an RPC server.
	eth *rpcETHAPI      // eth implements procedures with the "eth_" prefix.
	net *rpcNETAPI      // net implements procedures with the "net_" prefix.
	log log.Logger

	// List of endpoint callers.
	callers map[string]caller
	// Total timeout for all endpoints.
	totalTimeout time.Duration
	// Timeout for slower endpoints, when it exceeds, request will be canceled
	// if there is enough responses.
	gracefulTimeout time.Duration

	// Resolvers used to convert multiple responses into a single response:
	defaultResolver     *defaultResolver
	gasValueResolver    *gasValueResolver
	blockNumberResolver *blockNumberResolver
}

type rpcETHAPI struct {
	handler *server
}

type rpcNETAPI struct {
	handler *server
}

func NewServer(opts ...Option) (http.Handler, error) {
	h := &server{
		rpc:     gethRPC.NewServer(),
		callers: map[string]caller{},
	}
	eth := &rpcETHAPI{handler: h}
	net := &rpcNETAPI{handler: h}
	h.eth = eth
	h.net = net
	if err := h.rpc.RegisterName("eth", eth); err != nil {
		return nil, err
	}
	if err := h.rpc.RegisterName("net", net); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err := opt(h)
		if err != nil {
			return nil, fmt.Errorf("rpc-splitter error: unable to apply option: %w", err)
		}
	}
	if h.log == nil {
		h.log = null.New()
	}
	if h.callers == nil {
		return nil, fmt.Errorf("rpc-splitter error: WithEndpoints option is required")
	}
	if h.defaultResolver == nil || h.gasValueResolver == nil || h.blockNumberResolver == nil {
		return nil, fmt.Errorf("rpc-splitter error: WithRequirements option is required")
	}
	if h.totalTimeout == 0 {
		h.totalTimeout = defaultTotalTimeout
	}
	if h.gracefulTimeout == 0 {
		h.gracefulTimeout = defaultGracefulTimeout
	}
	h.log = h.log.WithField("tag", LoggerTag)
	return h, nil
}

func (s *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.rpc.ServeHTTP(rw, req)
}

// BlockNumber implements the "eth_blockNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) BlockNumber() (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Number{}
	err := r.handler.call(ctx, r.handler.blockNumberResolver, res, "eth_blockNumber")

	return res, err
}

// GetBlockByHash implements the "eth_getBlockByHash" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GetBlockByHash(blockHash types.Hash, obj bool) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	var res any
	switch obj {
	case true:
		res = &types.BlockTxObjects{}
	case false:
		res = &types.BlockTxHashes{}
	}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getBlockByHash", blockHash, obj)

	return res, err
}

// GetBlockByNumber implements the "eth_getBlockByNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) GetBlockByNumber(blockNumber types.Number, obj bool) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	var res any
	switch obj {
	case true:
		res = &types.BlockTxObjects{}
	case false:
		res = &types.BlockTxHashes{}
	}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getBlockByNumber", blockNumber, obj)

	return res, err
}

// GetTransactionByHash implements the "eth_getTransactionByHash" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) GetTransactionByHash(txHash types.Hash) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Transaction{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getTransactionByHash", txHash)

	return res, err
}

// GetTransactionCount implements the "eth_getTransactionCount" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetTransactionCount(addr types.Address, blockID types.BlockNumber) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Number{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getTransactionCount", addr, blockNumber)

	return res, err
}

// GetTransactionReceipt implements the "eth_getTransactionReceipt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) GetTransactionReceipt(txHash types.Hash) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.TransactionReceiptType{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getTransactionReceipt", txHash)

	return res, err
}

// TODO: eth_getBlockTransactionCountByHash
// TODO: eth_getBlockTransactionCountByNumber
// TODO: eth_getTransactionByBlockHashAndIndex
// TODO: eth_getTransactionByBlockNumberAndIndex

// SendRawTransaction implements the "eth_sendRawTransaction" call.
//
// It returns the most common response.
func (r *rpcETHAPI) SendRawTransaction(data types.Bytes) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Hash{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_sendRawTransaction", data)

	return res, err
}

// GetBalance implements the "eth_getBalance" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetBalance(addr types.Address, blockID types.BlockNumber) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Number{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getBalance", addr, blockNumber)

	return res, err
}

// GetCode implements the "eth_getCode" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetCode(addr types.Address, blockID types.BlockNumber) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Bytes{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getCode", addr, blockNumber)

	return res, err
}

// GetStorageAt implements the "eth_getStorageAt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetStorageAt(data types.Address, pos types.Number, blockID types.BlockNumber) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Hash{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getStorageAt", data, pos, blockNumber)

	return res, err
}

// TODO: eth_accounts
// TODO: eth_getProof

// Call implements the "eth_call" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) Call(args Any, blockID types.BlockNumber, overrides *Any) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Bytes{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_call", args, blockNumber, overrides)

	return res, err
}

// GetLogs implements the "eth_getLogs" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetLogs(logFilter types.FilterLogsQuery) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	if logFilter.FromBlock != nil {
		blockNumber, err := r.handler.taggedBlockToNumber(ctx, *logFilter.FromBlock)
		if err != nil {
			return nil, err
		}
		*logFilter.FromBlock = blockNumber
	}
	if logFilter.ToBlock != nil {
		blockNumber, err := r.handler.taggedBlockToNumber(ctx, *logFilter.ToBlock)
		if err != nil {
			return nil, err
		}
		*logFilter.ToBlock = blockNumber
	}
	res := &[]types.Log{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_getLogs", logFilter)

	return res, err
}

// TODO: eth_protocolVersion

// GasPrice implements the "eth_gasPrice" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GasPrice() (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Number{}
	err := r.handler.call(ctx, r.handler.gasValueResolver, res, "eth_gasPrice")

	return res, err
}

// EstimateGas implements the "eth_estimateGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) EstimateGas(args Any, blockID types.BlockNumber) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, blockID)
	if err != nil {
		return nil, err
	}
	res := &types.Number{}
	err = r.handler.call(ctx, r.handler.gasValueResolver, res, "eth_estimateGas", args, blockNumber)

	return res, err
}

// FeeHistory implements the "eth_feeHistory" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) FeeHistory(count types.Number, newestBlockID types.BlockNumber, percentiles Any) (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	blockNumber, err := r.handler.taggedBlockToNumber(ctx, newestBlockID)
	if err != nil {
		return nil, err
	}
	res := &types.FeeHistory{}
	err = r.handler.call(ctx, r.handler.defaultResolver, res, "eth_feeHistory", count, blockNumber, percentiles)

	return res, err
}

// MaxPriorityFeePerGas implements the "eth_maxPriorityFeePerGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) MaxPriorityFeePerGas() (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Number{}
	err := r.handler.call(ctx, r.handler.gasValueResolver, res, "eth_maxPriorityFeePerGas")

	return res, err
}

// ChainId implements the "eth_chainId" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcETHAPI) ChainId() (any, error) { //nolint:revive,stylecheck
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &types.Number{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "eth_chainId")

	return res, err
}

// TODO: eth_getUncleByBlockNumberAndIndex
// TODO: eth_getUncleByBlockHashAndIndex
// TODO: eth_getUncleCountByBlockHash
// TODO: eth_getUncleCountByBlockNumber
// TODO: eth_getFilterChanges
// TODO: eth_getFilterLogs
// TODO: eth_newBlockFilter
// TODO: eth_newFilter
// TODO: eth_newPendingTransactionFilter
// TODO: eth_uninstallFilter

// Version implements the "net_version" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minRes method.
func (r *rpcNETAPI) Version() (any, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), r.handler.totalTimeout)
	defer ctxCancel()

	res := &Any{}
	err := r.handler.call(ctx, r.handler.defaultResolver, res, "net_version")

	return res, err
}

// taggedBlockToNumber returns a block number for tagged blocks. This is
// necessary because different RPC endpoints may convert tags to different
// block numbers.
func (s *server) taggedBlockToNumber(ctx context.Context, blockID types.BlockNumber) (types.BlockNumber, error) {
	if len(s.callers) == 1 {
		return blockID, nil
	}
	if !blockID.IsTag() {
		return blockID, nil
	}
	if blockID.IsEarliest() {
		// The earliest block will be completely different on different
		// endpoints. It is impossible to reliably support it.
		return types.BlockNumber{}, errors.New("earliest tag is not supported")
	}
	// The latest and pending blocks are handled in the same way.
	res := &types.Number{}
	err := s.call(ctx, s.blockNumberResolver, res, "eth_blockNumber")
	if err != nil {
		return types.BlockNumber{}, err
	}
	return types.BlockNumber(*res), nil
}

// call executes RPC on all endpoints with the given arguments. If the context is
// canceled before the call has successfully returned, call returns immediately.
//
// The result must be a pointer with a proper type.
//
//nolint:funlen
func (s *server) call(
	ctx context.Context,
	resolver resolver,
	result any,
	method string,
	args ...any,
) error {

	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer")
	}

	// Recover from panics.
	defer func() {
		if r := recover(); r != nil {
			s.log.
				WithField("method", method).
				// WithField("args", args).
				WithError(fmt.Errorf("panic: %s", r)).
				Error("Panic")
		}
	}()

	// Send request to all endpoints.
	ch := make(chan any, len(s.callers))
	rt := reflect.TypeOf(result).Elem()
	for n, c := range s.callers {
		n, c := n, c
		go func() {
			t := time.Now()
			var res any
			var err error
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %s", r)
				}
				switch {
				case err != nil:
					s.log.
						WithField("name", n).
						WithField("method", method).
						// WithField("args", args).
						WithField("duration", time.Since(t)).
						WithError(err).
						Debug("Call error")
					ch <- err
				default:
					s.log.
						WithField("name", n).
						WithField("method", method).
						// WithField("args", args).
						WithField("duration", time.Since(t)).
						Debug("Call")
					ch <- res
				}
			}()
			res = reflect.New(rt).Interface()
			err = c.CallContext(ctx, res, method, removeTrailingNilArgs(args)...)
		}()
	}
	// Wait for response. The following code will wait for the above requests
	// to complete, but if gracefulTimeout exceeds and there are enough
	// responses to return a valid response, then the context will be canceled
	// and the response returned.
	t := time.NewTimer(s.gracefulTimeout)
	defer t.Stop()
	var rs []any
	for {
		wait := true
		select {
		case r := <-ch:
			rs = append(rs, r)
		case <-t.C:
			wait = false
		}
		if len(rs) == len(s.callers) {
			wait = false
		}
		if !wait {
			res, err := resolver.resolve(rs)
			switch {
			case err == nil:
				reflect.ValueOf(result).Elem().Set(reflect.ValueOf(res).Elem())
				return nil
			case len(rs) >= len(s.callers):
				return err
			}
		}
	}
}

// removeTrailingNilArgs removes trailing nil parameters from the params
// slice. Some RPC servers do not like null parameters and will return a
// "bad request" error if they occur.
func removeTrailingNilArgs(params []any) []any {
	for i := len(params) - 1; i >= 0; i-- {
		if isNil(params[i]) {
			continue
		}
		return params[0 : i+1]
	}
	return nil
}

func isNil(v any) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

// Any stores an argument in its raw form, and it is passed to the RPC server
// unmodified.
type Any struct{ j any }

// MarshalJSON implements the json.Marshaler interface.
func (t Any) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.j)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Any) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.j)
}
