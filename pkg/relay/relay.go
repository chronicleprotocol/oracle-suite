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

package relay

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/multicall"
	datapointStore "github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	musigStore "github.com/chronicleprotocol/oracle-suite/pkg/musig/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/errutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/sliceutil"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const LoggerTag = "RELAY"

const (
	// baseGasUsage is the base gas usage for an Ethereum transaction.
	baseGasUsage = 21_000

	// gasUsageSoftCap is the soft cap for gas usage for an Ethereum transaction.
	// If the gas usage is above this value, no more transactions will be added
	// to the aggregate transaction.
	gasUsageSoftCap = 2_500_000

	// maxParallelCallProviders is the maximum number of call providers that
	// can be executed in parallel.
	maxParallelCallProviders = 8
)

type MedianContract interface {
	Client() rpc.RPC
	Address() types.Address
	Val(ctx context.Context) (*bn.DecFixedPointNumber, error)
	Age() contract.TypedSelfCaller[time.Time]
	Wat() contract.TypedSelfCaller[string]
	Bar() contract.TypedSelfCaller[int]
	Poke(vals []chronicle.MedianVal) contract.SelfTransactableCaller
}

type ScribeContract interface {
	Client() rpc.RPC
	Address() types.Address
	Read(ctx context.Context) (chronicle.PokeData, error)
	Wat() contract.TypedSelfCaller[string]
	Bar() contract.TypedSelfCaller[int]
	Feeds() contract.TypedSelfCaller[[]types.Address]
	Poke(pokeData chronicle.PokeData, schnorrData chronicle.SchnorrData) contract.SelfTransactableCaller
}

type OpScribeContract interface {
	ScribeContract
	ReadNext(ctx context.Context) (chronicle.PokeData, bool, error)
	OpPoke(pokeData chronicle.PokeData, schnorrData chronicle.SchnorrData, ecdsaData types.Signature) contract.SelfTransactableCaller
}

// callProvider provides a contract call that can be used to relay data to the
// contract.
type callProvider interface {
	// createRelayCall creates a callable that can be used to relay data to the
	// contract. It returns the gas estimate for the transaction and the callable.
	// If callable is nil, then there is no data to relay.
	createRelayCall(ctx context.Context) []relayCall
}

type relayCall struct {
	client      rpc.RPC
	address     types.Address
	callable    contract.Callable
	gasEstimate uint64
}

// Relay is a service that relays data to the blockchain.
type Relay struct {
	ctx       context.Context
	waitCh    chan error
	ticker    *timeutil.Ticker
	providers []callProvider
	log       log.Logger
}

// Config is the configuration for the Relay.
type Config struct {
	// Medians is the list of median contracts configuration.
	Medians []ConfigMedian

	// Scribes is the list of scribe contracts configuration.
	Scribes []ConfigScribe

	// OptimisticScribes is the list of scribe optimistic contracts configuration.
	OptimisticScribes []ConfigOptimisticScribe

	// Ticker notifies the relay to check if an update is required.
	Ticker *timeutil.Ticker

	// Logger is a current logger interface used by the Relay.
	// If nil, null logger will be used.
	Logger log.Logger
}

type ConfigMedian struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// DataPointStore is the store used to retrieve data points.
	DataPointStore datapointStore.DataPointProvider

	// DataModel is the name of the data model from which data points
	// are retrieved.
	DataModel string

	// ContractAddress is the address of the Median contract.
	ContractAddress types.Address

	// FeedAddresses is the list of feed addresses that are allowed to
	// update the Median contract.
	FeedAddresses []types.Address

	// Spread is the minimum spread between the oracle price and new
	// price required to send update.
	Spread float64

	// Expiration is the minimum time difference between the last oracle
	// update on the Median contract and current time required to send
	// update.
	Expiration time.Duration
}

type ConfigScribe struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// MuSigStore is the store used to retrieve MuSig signatures.
	MuSigStore musigStore.SignatureProvider

	// DataModel is the name of the data model that is used to update
	// the Scribe contract.
	DataModel string

	// ContractAddress is the address of the Scribe contract.
	ContractAddress types.Address

	// Spread is the minimum calcSpread between the oracle price and new
	// price required to send update.
	Spread float64

	// Expiration is the minimum time difference between the last oracle
	// update on the Scribe contract and current time required to send
	// update.
	Expiration time.Duration
}

type ConfigOptimisticScribe struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// MuSigStore is the store used to retrieve MuSig signatures.
	MuSigStore musigStore.SignatureProvider

	// DataModel is the name of the data model that is used to update
	// the OptimisticScribe contract.
	DataModel string

	// ContractAddress is the address of the OptimisticScribe contract.
	ContractAddress types.Address

	// Spread is the minimum calcSpread between the oracle price and new
	// price required to send regular update.
	Spread float64

	// Expiration is the minimum time difference between the last oracle
	// update on the Scribe contract and current time required to send
	// regular update.
	Expiration time.Duration

	// OptimisticSpread is the minimum time difference between the last oracle
	// update on the Scribe contract and current time required to send
	// optimistic update.
	OptimisticSpread float64

	// OptimisticExpiration is the minimum time difference between the last
	// oracle update on the Scribe contract and current time required to send
	// optimistic update.
	OptimisticExpiration time.Duration
}

// New creates a new Relay instance.
func New(cfg Config) (*Relay, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	logger := cfg.Logger.WithField("tag", LoggerTag)
	r := &Relay{
		waitCh: make(chan error),
		ticker: cfg.Ticker,
		log:    logger,
	}
	for _, s := range cfg.OptimisticScribes {
		contract := chronicle.NewOpScribe(s.Client, s.ContractAddress)
		r.providers = append(r.providers, &opScribe{
			scribe: scribe{
				contract:   contract,
				muSigStore: s.MuSigStore,
				dataModel:  s.DataModel,
				spread:     s.Spread,
				expiration: s.Expiration,
				log:        logger,
			},
			opContract:   contract,
			opSpread:     s.OptimisticSpread,
			opExpiration: s.OptimisticExpiration,
		})
	}
	for _, s := range cfg.Scribes {
		r.providers = append(r.providers, &scribe{
			contract:   chronicle.NewScribe(s.Client, s.ContractAddress),
			muSigStore: s.MuSigStore,
			dataModel:  s.DataModel,
			spread:     s.Spread,
			expiration: s.Expiration,
			log:        logger,
		})
	}
	for _, m := range cfg.Medians {
		r.providers = append(r.providers, &median{
			contract:       chronicle.NewMedian(m.Client, m.ContractAddress),
			dataPointStore: m.DataPointStore,
			feedAddresses:  m.FeedAddresses,
			dataModel:      m.DataModel,
			spread:         m.Spread,
			expiration:     m.Expiration,
			log:            logger,
		})
	}
	return r, nil
}

// Start implements the supervisor.Service interface.
func (m *Relay) Start(ctx context.Context) error {
	if m.ctx != nil {
		return errors.New("service can be started only once")
	}
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	m.log.Info("Starting")
	m.ctx = ctx
	go m.relayRoutine()
	go m.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (m *Relay) Wait() <-chan error {
	return m.waitCh
}

func (m *Relay) sendRelayTransactions() {
	for client, calls := range m.relayCalls() {
		// Note, that there is not need to create a separate branch for
		// a single call because MultiCall internally handles this case.
		call := multicall.AggregateCallables(client, calls...).AllowFail()
		txHash, tx, err := call.SendTransaction(m.ctx)
		if err != nil {
			if strings.Contains(err.Error(), "nonce too low") || strings.Contains(err.Error(), "replacement transaction underpriced") {
				m.log.
					WithError(err).
					WithFields(log.Fields{
						"txTo":              call.Address(),
						"txInput":           errutil.Ignore(call.CallData()),
						"contractAddresses": addressesFromCalls(calls),
					}).
					Info("Unable to send transaction, previous transaction is still pending")
				continue
			}
			m.log.
				WithError(err).
				WithFields(log.Fields{
					"txTo":              call.Address(),
					"txInput":           errutil.Ignore(call.CallData()),
					"contractAddresses": addressesFromCalls(calls),
				}).
				WithAdvice("Ignore if it is related to temporary network issues").
				Error("Failed to send transaction")
			continue
		}
		m.log.
			WithFields(log.Fields{
				"txHash":                 txHash,
				"txType":                 tx.Type,
				"txFrom":                 tx.From,
				"txTo":                   tx.To,
				"txChainId":              tx.ChainID,
				"txNonce":                tx.Nonce,
				"txGasPrice":             tx.GasPrice,
				"txGasLimit":             tx.GasLimit,
				"txMaxFeePerGas":         tx.MaxFeePerGas,
				"txMaxPriorityFeePerGas": tx.MaxPriorityFeePerGas,
				"contractAddresses":      addressesFromCalls(calls),
				"txInput":                hexutil.BytesToHex(tx.Input),
			}).
			Info("Relay transaction sent")
	}
}

func (m *Relay) relayCalls() map[rpc.RPC][]contract.Callable {
	var (
		mu        = sync.Mutex{}
		wg        = sync.WaitGroup{}
		limiter   = make(chan struct{}, maxParallelCallProviders)
		gasUsage  = make(map[rpc.RPC]uint64)
		contracts = make(map[rpc.RPC][]types.Address)
		calls     = make(map[rpc.RPC][]contract.Callable)
	)
	for _, u := range m.providers {
		go func(u callProvider) {
			defer wg.Done()
			defer func() { <-limiter }()
			limiter <- struct{}{}
			for _, c := range u.createRelayCall(m.ctx) {
				mu.Lock()
				if gasUsage[c.client] >= gasUsageSoftCap {
					// If the gas usage is above the soft cap, then do not
					// add any more transactions to the aggregate transaction.
					mu.Unlock()
					continue
				}
				if sliceutil.Contains(contracts[c.client], c.address) {
					// If there is already a transaction for the contract,
					// then do not add another one.
					mu.Unlock()
					continue
				}
				gasEstimate := c.gasEstimate + 700 //nolint:gomnd // 700 is the minimum gas usage for a call
				if gasEstimate > baseGasUsage {
					gasEstimate -= baseGasUsage
				}
				gasUsage[c.client] += gasEstimate
				contracts[c.client] = append(contracts[c.client], c.address)
				calls[c.client] = append(calls[c.client], c.callable)
				mu.Unlock()
			}
		}(u)
	}
	wg.Wait()
	return calls
}

func (m *Relay) relayRoutine() {
	m.ticker.Start(m.ctx)
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.ticker.TickCh():
			m.sendRelayTransactions()
		}
	}
}

func (m *Relay) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}

func addressesFromCalls(calls []contract.Callable) []types.Address {
	addresses := make([]types.Address, 0, len(calls))
	for _, c := range calls {
		addresses = append(addresses, c.Address())
	}
	return addresses
}
