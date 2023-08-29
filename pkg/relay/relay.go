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
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/relay/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

const LoggerTag = "RELAY"

type MedianContract interface {
	Val(ctx context.Context) (*bn.DecFixedPointNumber, error)
	Age(ctx context.Context) (time.Time, error)
	Wat(ctx context.Context) (string, error)
	Bar(ctx context.Context) (int, error)
	Poke(ctx context.Context, vals []contract.MedianVal) (*types.Hash, *types.Transaction, error)
}

type ScribeContract interface {
	Wat(ctx context.Context) (string, error)
	Bar(ctx context.Context) (int, error)
	Feeds(ctx context.Context) ([]types.Address, []uint8, error)
	Read(ctx context.Context) (*bn.DecFixedPointNumber, time.Time, error)
	Poke(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData) (*types.Hash, *types.Transaction, error)
}

type OpScribeContract interface {
	ScribeContract
	OpPoke(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData, ecdsaData types.Signature) (*types.Hash, *types.Transaction, error)
}

// Relay is a service that relays data points to the blockchain.
type Relay struct {
	ctx    context.Context
	waitCh chan error
	log    log.Logger

	medians   []*medianWorker
	scribes   []*scribeWorker
	opScribes []*opScribeWorker
}

// Config is the configuration for the Relay.
type Config struct {
	// Medians is the list of median contracts configured for the relay.
	Medians []ConfigMedian

	// Scribes is the list of scribe contracts configured for the relay.
	Scribes []ConfigScribe

	// OptimisticScribes is the list of optimistic scribe contracts configured
	// for the relay.
	OptimisticScribes []ConfigOptimisticScribe

	// Logger is a current logger interface used by the Feed.
	// If nil, null logger will be used.
	Logger log.Logger
}

type ConfigMedian struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// DataPointStore is the store used to retrieve data points.
	DataPointStore *store.Store

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

	// Ticker notifies the relay to check if an update is required.
	Ticker *timeutil.Ticker
}

type ConfigScribe struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// MuSigStore is the store used to retrieve MuSig signatures.
	MuSigStore *MuSigStore

	// DataModel is the name of the data model that is used to update
	// the Scribe contract.
	DataModel string

	// ContractAddress is the address of the Scribe contract.
	ContractAddress types.Address

	// FeedAddresses is the list of feed addresses that are allowed to
	// update the Scribe contract.
	FeedAddresses []types.Address

	// Spread is the minimum calcSpread between the oracle price and new
	// price required to send update.
	Spread float64

	// Expiration is the minimum time difference between the last oracle
	// update on the Scribe contract and current time required to send
	// update.
	Expiration time.Duration

	// Ticker notifies the relay to check if an update is required.
	Ticker *timeutil.Ticker
}

type ConfigOptimisticScribe struct {
	// Client is the RPC client used to interact with the blockchain.
	Client rpc.RPC

	// MuSigStore is the store used to retrieve MuSig signatures.
	MuSigStore *MuSigStore

	// DataModel is the name of the data model that is used to update
	// the OptimisticScribe contract.
	DataModel string

	// ContractAddress is the address of the OptimisticScribe contract.
	ContractAddress types.Address

	// FeedAddresses is the list of feed addresses that are allowed to
	// update the Scribe contract.
	FeedAddresses []types.Address

	// Spread is the minimum calcSpread between the oracle price and new
	// price required to send update.
	Spread float64

	// Expiration is the minimum time difference between the last oracle
	// update on the Scribe contract and current time required to send
	// update.
	Expiration time.Duration

	// Ticker notifies the relay to check if an update is required.
	Ticker *timeutil.Ticker
}

// New creates a new Relay instance.
func New(cfg Config) (*Relay, error) {
	if cfg.Logger == nil {
		cfg.Logger = null.New()
	}
	logger := cfg.Logger.WithField("tag", LoggerTag)
	r := &Relay{
		waitCh: make(chan error),
		log:    logger,
	}
	for _, m := range cfg.Medians {
		r.medians = append(r.medians, &medianWorker{
			log:            logger,
			dataPointStore: m.DataPointStore,
			feedAddresses:  m.FeedAddresses,
			contract:       contract.NewMedian(m.Client, m.ContractAddress),
			dataModel:      m.DataModel,
			spread:         m.Spread,
			expiration:     m.Expiration,
			ticker:         m.Ticker,
		})
	}
	for _, s := range cfg.Scribes {
		r.scribes = append(r.scribes, &scribeWorker{
			log:        logger,
			muSigStore: s.MuSigStore,
			contract:   contract.NewScribe(s.Client, s.ContractAddress),
			dataModel:  s.DataModel,
			spread:     s.Spread,
			expiration: s.Expiration,
			ticker:     s.Ticker,
		})
	}
	for _, s := range cfg.OptimisticScribes {
		r.opScribes = append(r.opScribes, &opScribeWorker{
			log:        logger,
			muSigStore: s.MuSigStore,
			contract:   contract.NewOpScribe(s.Client, s.ContractAddress),
			dataModel:  s.DataModel,
			spread:     s.Spread,
			expiration: s.Expiration,
			ticker:     s.Ticker,
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
	for _, w := range m.medians {
		go w.workerRoutine(ctx)
	}
	for _, w := range m.scribes {
		go w.workerRoutine(ctx)
	}
	for _, w := range m.opScribes {
		go w.workerRoutine(ctx)
	}
	go m.contextCancelHandler()
	return nil
}

// Wait implements the supervisor.Service interface.
func (m *Relay) Wait() <-chan error {
	return m.waitCh
}

func (m *Relay) contextCancelHandler() {
	defer func() { close(m.waitCh) }()
	defer m.log.Info("Stopped")
	<-m.ctx.Done()
}
