//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"

	gethRPC "github.com/ethereum/go-ethereum/rpc"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "RPCSPLITTER"

// maxBlocksBehind is the number of blocks behind the median of the block
// numbers reported by the endpoints that determines the lowest block number
// that can be used.
//
// For example, if 5 endpoints report the following block numbers: 9, 9, 9, 6, 4
// then the median block number is 9 but the lowest block number we can use is 6
// because 9-3 is 6, so only the response from the endpoint that reported 4 as
// the block number will be discarded.
const maxBlocksBehind = 3

type rpcCaller interface {
	Call(result interface{}, method string, args ...interface{}) error
}

// handler is an RPC proxy server. It merges multiple RPC endpoints into one.
type handler struct {
	rpc *gethRPC.Server // rpc is an RPC server.
	cli []rpcClient     // cli is a list of RPC clients.
	eth *rpcETHAPI      // eth implements procedures with the "eth_" prefix.
	net *rpcNETAPI      // net implements procedures with the "net_" prefix.
	log log.Logger
}

type rpcClient struct {
	rpcCaller
	endpoint string
}

type rpcETHAPI struct {
	handler *handler
}

type rpcNETAPI struct {
	handler *handler
}

func NewHandler(endpoints []string, log log.Logger) (http.Handler, error) {
	var clients []rpcClient
	for _, e := range endpoints {
		c, err := gethRPC.Dial(e)
		if err != nil {
			return nil, err
		}
		clients = append(clients, rpcClient{
			rpcCaller: c,
			endpoint:  e,
		})
	}
	return newHandlerWithClients(clients, log)
}

func newHandlerWithClients(clients []rpcClient, log log.Logger) (http.Handler, error) {
	h := &handler{
		rpc: gethRPC.NewServer(),
		cli: make([]rpcClient, len(clients)),
		log: log.WithField("tag", LoggerTag),
	}
	eth := &rpcETHAPI{handler: h}
	net := &rpcNETAPI{handler: h}
	h.eth = eth
	h.net = net
	for n, c := range clients {
		h.cli[n] = c
	}
	if err := h.rpc.RegisterName("eth", eth); err != nil {
		return nil, err
	}
	if err := h.rpc.RegisterName("net", net); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.rpc.ServeHTTP(rw, req)
}

// BlockNumber implements the "eth_blockNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) BlockNumber() (interface{}, error) {
	return useMedianDist(
		r.handler.call((*numberType)(nil), "eth_blockNumber"), r.handler.minReq(),
		maxBlocksBehind,
	)
}

// GetBlockByHash implements the "eth_getBlockByHash" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GetBlockByHash(blockHash hashType, obj bool) (interface{}, error) {
	return useMostCommon(
		r.handler.call(blockTypeNilPtr(obj), "eth_getBlockByHash", blockHash, obj),
		r.handler.minReq(),
	)
}

// GetBlockByNumber implements the "eth_getBlockByNumber" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetBlockByNumber(blockNumber numberType, obj bool) (interface{}, error) {
	return useMostCommon(
		r.handler.call(blockTypeNilPtr(obj), "eth_getBlockByNumber", blockNumber, obj),
		r.handler.minReq(),
	)
}

// GetTransactionByHash implements the "eth_getTransactionByHash" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetTransactionByHash(txHash hashType) (interface{}, error) {
	return useMostCommon(
		r.handler.call((*transactionType)(nil), "eth_getTransactionByHash", txHash),
		r.handler.minReq(),
	)
}

// GetTransactionCount implements the "eth_getTransactionCount" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetTransactionCount(addr addressType, blockID blockIDType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*numberType)(nil), "eth_getTransactionCount", addr, blockNumber),
		r.handler.minReq(),
	)
}

// GetTransactionReceipt implements the "eth_getTransactionReceipt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) GetTransactionReceipt(txHash hashType) (interface{}, error) {
	return useMostCommon(
		r.handler.call((*transactionReceiptType)(nil), "eth_getTransactionReceipt", txHash),
		r.handler.minReq(),
	)
}

// TODO: eth_getBlockTransactionCountByHash
// TODO: eth_getBlockTransactionCountByNumber
// TODO: eth_getTransactionByBlockHashAndIndex
// TODO: eth_getTransactionByBlockNumberAndIndex

// SendRawTransaction implements the "eth_sendRawTransaction" call.
//
// It returns the most common response.
func (r *rpcETHAPI) SendRawTransaction(data bytesType) (interface{}, error) {
	return useMostCommon(
		r.handler.call((*hashType)(nil), "eth_sendRawTransaction", data),
		r.handler.minReq(),
	)
}

// GetBalance implements the "eth_getBalance" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetBalance(addr addressType, blockID blockIDType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*numberType)(nil), "eth_getBalance", addr, blockNumber),
		r.handler.minReq(),
	)
}

// GetCode implements the "eth_getCode" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetCode(addr addressType, blockID blockIDType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*bytesType)(nil), "eth_getCode", addr, blockNumber),
		r.handler.minReq(),
	)
}

// GetStorageAt implements the "eth_getStorageAt" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetStorageAt(data addressType, pos numberType, blockID blockIDType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*hashType)(nil), "eth_getStorageAt", data, pos, blockNumber),
		r.handler.minReq(),
	)
}

// TODO: eth_accounts
// TODO: eth_getProof

// Call implements the "eth_call" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) Call(args jsonType, blockID blockIDType, overrides *jsonType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*bytesType)(nil), "eth_call", args, blockNumber, overrides),
		r.handler.minReq(),
	)
}

// GetLogs implements the "eth_getLogs" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) GetLogs(logFilter logFilterType) (interface{}, error) {
	if logFilter.FromBlock != nil {
		blockNumber, err := r.handler.taggedBlockToNumber(*logFilter.FromBlock)
		if err != nil {
			return nil, err
		}
		*logFilter.FromBlock = blockIDType(blockNumber)
	}
	if logFilter.ToBlock != nil {
		blockNumber, err := r.handler.taggedBlockToNumber(*logFilter.ToBlock)
		if err != nil {
			return nil, err
		}
		*logFilter.ToBlock = blockIDType(blockNumber)
	}
	return useMostCommon(
		r.handler.call((*[]logType)(nil), "eth_getLogs", logFilter),
		r.handler.minReq(),
	)
}

// TODO: eth_protocolVersion

// GasPrice implements the "eth_gasPrice" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) GasPrice() (interface{}, error) {
	return useMedian(
		r.handler.call((*numberType)(nil), "eth_gasPrice"),
		r.handler.minReq(),
	)
}

// EstimateGas implements the "eth_estimateGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
//
// If the block number is set to "latest" or "pending", it will be replaced by
// the block number returned by the BlockNumber method. The "earliest" tag is
// not supported.
func (r *rpcETHAPI) EstimateGas(args jsonType, blockID blockIDType) (interface{}, error) {
	blockNumber, err := r.handler.taggedBlockToNumber(blockID)
	if err != nil {
		return nil, err
	}
	return useMedian(
		r.handler.call((*numberType)(nil), "eth_estimateGas", args, blockNumber),
		r.handler.minReq(),
	)
}

// FeeHistory implements the "eth_feeHistory" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
func (r *rpcETHAPI) FeeHistory(count numberType, newestBlockID blockIDType, percentiles jsonType) (interface{}, error) {
	blockToNumber, err := r.handler.taggedBlockToNumber(newestBlockID)
	if err != nil {
		return nil, err
	}
	return useMostCommon(
		r.handler.call((*feeHistoryType)(nil), "eth_feeHistory", count, blockToNumber, percentiles),
		r.handler.minReq(),
	)
}

// MaxPriorityFeePerGas implements the "eth_maxPriorityFeePerGas" call.
//
// The number returned by this method is the median of all numbers returned
// by the endpoints.
func (r *rpcETHAPI) MaxPriorityFeePerGas() (interface{}, error) {
	return useMedian(
		r.handler.call((*numberType)(nil), "eth_maxPriorityFeePerGas"),
		r.handler.minReq(),
	)
}

// ChainId implements the "eth_chainId" call.
//
// It returns the most common response that occurred at least as many times as
// specified in the minReq method.
//nolint:revive,stylecheck
func (r *rpcETHAPI) ChainId() (interface{}, error) {
	return useMostCommon(
		r.handler.call((*numberType)(nil), "eth_chainId"),
		r.handler.minReq(),
	)
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
// specified in the minReq method.
func (r *rpcNETAPI) Version() (interface{}, error) {
	return useMostCommon(
		r.handler.call((*jsonType)(nil), "net_version"),
		r.handler.minReq(),
	)
}

// call executes RPC on all endpoints and returns a slice with all results.
//
// The typ argument must be an empty pointer with a type to which the results
// will be converted.
func (h *handler) call(typ interface{}, method string, params ...interface{}) (res []interface{}) {
	h.removeTrailingNilParams(&params)

	// Send request to all endpoints.
	ch := make(chan interface{}, len(h.cli))
	wg := &sync.WaitGroup{}
	wg.Add(len(h.cli))
	for _, cli := range h.cli {
		cli := cli
		go func() {
			defer wg.Done()
			ch <- h.callOne(cli, typ, method, params...)
		}()
	}

	// Wait for responses.
	wg.Wait()
	for len(ch) > 0 {
		res = append(res, <-ch)
	}
	return res
}

// call executes RPC on a single endpoint.
//
// The typ argument must be an empty pointer with a type to which the results
// will be converted.
func (h *handler) callOne(cli rpcClient, typ interface{}, method string, params ...interface{}) (out interface{}) {
	var err error
	var val interface{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %s", r)
		}
		if err != nil {
			h.log.
				WithField("endpoint", cli.endpoint).
				WithField("method", method).
				WithError(err).
				Error("Call error")
			// Errors are returned the same way as other values. The out
			// variable is a named return parameter.
			out = err
		}
	}()
	h.log.
		WithField("endpoint", cli.endpoint).
		WithField("method", method).
		Info("Call")
	// typ is a nil pointer, and it is shared by multiple goroutines, so it
	// cannot be used directly with cli.Call.
	val = reflect.New(reflect.TypeOf(typ).Elem()).Interface()
	err = cli.Call(val, method, params...)
	out = val
	return
}

// removeTrailingNilParams removes trailing nil parameters from the params
// slice. Some RPC servers do not like null parameters and will return a
// "bad request" error if they occur.
func (h *handler) removeTrailingNilParams(params *[]interface{}) {
	for i := len(*params) - 1; i >= 0; i-- {
		if len(*params)-1 == i && isNil((*params)[i]) {
			*params = (*params)[0:i]
		}
	}
}

// taggedBlockToNumber returns a block number for tagged blocks. This is
// necessary because different RPC endpoints may convert tags to different
// block numbers.
func (h *handler) taggedBlockToNumber(blockID blockIDType) (numberType, error) {
	if !blockID.IsTag() {
		return numberType(blockID), nil
	}
	if blockID.IsEarliest() {
		// The earliest block will be completely different on different
		// endpoints. It is impossible to reliably support it.
		return numberType{}, errors.New("earliest tag is not supported")
	}
	// The latest and pending blocks are handled in the same way.
	bn, err := h.eth.BlockNumber()
	if err != nil {
		return numberType{}, err
	}
	return *(bn.(*numberType)), nil
}

// minReq returns a number indicating how many times the same response
// must be returned by different endpoints to be considered valid.
func (h *handler) minReq() int {
	l := len(h.cli)
	if l <= 2 {
		return l
	}
	return l - 1
}

type rpcErrors []error

func (e rpcErrors) Error() string {
	switch len(e) {
	case 0:
		return "unknown error"
	case 1:
		return e[0].Error()
	default:
		s := strings.Builder{}
		s.WriteString("the following errors occurred: ")
		s.WriteString("[")
		for n, err := range e {
			s.WriteString(err.Error())
			if n < len(e)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteString("]")
		return s.String()
	}
}

// addError adds an error to an error slice. If errs is not an error slice it
// will be converted into one. If there is already an error with the same
// message in the slice, it will not be added.
func addError(errs error, err error, prepend bool) error {
	if errs, ok := errs.(rpcErrors); ok {
		msg := err.Error()
		for _, e := range errs {
			if e.Error() == msg {
				return errs
			}
		}
		if prepend {
			return append(rpcErrors{err}, errs...)
		}
		return append(errs, err)
	}
	if errs == nil {
		return rpcErrors{err}
	}
	return addError(rpcErrors{errs}, err, prepend)
}

// useMostCommon compares all responses returned from RPC endpoints and chooses
// the one that was repeated at least as many times as indicated by the minReq
// arg. Errors in the slice are not counted as responses and will be returned
// as one error if no valid response can be found.
func useMostCommon(s []interface{}, minReq int) (interface{}, error) {
	var err error
	// Count the number of occurrences of each item by comparing each item
	// in the slice with every other item. The result is stored in a map,
	// where the key is the item itself and the value is the number of
	// occurrences.
	occurs := map[interface{}]int{}
	maxOccurs := 0
	for _, a := range s {
		if e, ok := a.(error); ok {
			err = addError(err, e, false)
			continue
		}
		f := false
		for b := range occurs {
			if compare(a, b) {
				f = true
				break
			}
		}
		if f {
			continue
		}
		// Count occurrences of the `a` item.
		for _, b := range s {
			if compare(a, b) {
				occurs[a]++
				if occurs[a] > maxOccurs {
					maxOccurs = occurs[a]
				}
			}
		}
	}
	// Check if there are enough occurrences of the most common item.
	if maxOccurs < minReq {
		return nil, addError(
			err,
			errors.New("RPC servers returned different responses"),
			true,
		)
	}
	// Find the item with the maximum number of occurrences.
	var res interface{}
	for v, n := range occurs {
		if n == maxOccurs {
			if res != nil {
				// If `res` is not nil it means, that there are multiple items
				// that occurred `maxOccurs` times. In this case, we cannot
				// determine which one should be chosen.
				return nil, addError(
					err,
					errors.New("RPC servers returned different responses"),
					true,
				)
			}
			res = v
			// We do not want to "break" here because we still have to check
			// it there are no more items with the same number of occurrences.
		}
	}
	return res, nil
}

// useMedian calculates the median value of all numberType items in the given
// slice. There must be at least minReq items of type numberType in the slice,
// otherwise an error is returned.
func useMedian(s []interface{}, minReq int) (*numberType, error) {
	// Collect errors from responses.
	var err error
	for _, v := range s {
		if e, ok := v.(error); ok {
			err = addError(err, e, false)
		}
	}
	// Filter out anything that is not a numberType.
	var sn []*numberType
	for _, v := range s {
		if v, ok := v.(*numberType); ok {
			sn = append(sn, v)
		}
	}
	if len(sn) < minReq {
		err = addError(err, errors.New("not enough responses from RPC servers"), true)
		return nil, err
	}
	// Calculate the median.
	sort.Slice(sn, func(i, j int) bool {
		return sn[i].Big().Cmp(sn[j].Big()) < 0
	})
	if len(sn)%2 == 0 {
		m := len(s) / 2
		bx := sn[m-1].Big()
		by := sn[m].Big()
		return (*numberType)(new(big.Int).Div(new(big.Int).Add(bx, by), big.NewInt(2))), nil
	}
	return sn[len(sn)/2], nil
}

// useMedianDist works similarly to the useMedian function, but instead of
// median, it will return the lowest value that is greater than or equal to
// median-distance.
func useMedianDist(s []interface{}, minReq int, distance int64) (*numberType, error) {
	m, err := useMedian(s, minReq)
	if err != nil {
		return nil, err
	}
	// Filter out anything that is not a numberType.
	var sn []*numberType
	for _, v := range s {
		if v, ok := v.(*numberType); ok {
			sn = append(sn, v)
		}
	}
	// Calculate results.
	bx := m.Big()
	bl := new(big.Int).Sub(m.Big(), big.NewInt(distance))
	for _, n := range sn {
		bn := n.Big()
		if bn.Cmp(bl) >= 0 && bn.Cmp(bx) < 0 {
			bx = bn
		}
	}
	return (*numberType)(bx), nil
}

// blockTypeNilPtr returns a nil pointer to blockTxObjectsType if obj is set
// to true or blockTxHashesType otherwise.
func blockTypeNilPtr(obj bool) interface{} {
	if obj {
		return (*blockTxObjectsType)(nil)
	}
	return (*blockTxHashesType)(nil)
}

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
