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
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

type mockLogger struct {
	LevelFn      func() log.Level
	WithFieldFn  func(key string, value any) log.Logger
	WithFieldsFn func(fields log.Fields) log.Logger
	WithErrorFn  func(err error) log.Logger
	WithAdviceFn func(advice string) log.Logger
	DebugFn      func(args ...any)
	InfoFn       func(args ...any)
	WarnFn       func(args ...any)
	ErrorFn      func(args ...any)
	PanicFn      func(args ...any)
}

func newMockLogger(t *testing.T) *mockLogger {
	ml := &mockLogger{}
	ml.reset(t)
	return ml
}

func (m *mockLogger) reset(t *testing.T) {
	m.LevelFn = func() log.Level { return log.Debug }
	m.WithFieldFn = func(key string, value any) log.Logger { return m }
	m.WithFieldsFn = func(fields log.Fields) log.Logger { return m }
	m.WithErrorFn = func(err error) log.Logger { return m }
	m.WithAdviceFn = func(advice string) log.Logger { return m }
	m.DebugFn = func(args ...any) { assert.FailNow(t, "unexpected call to Debug") }
	m.InfoFn = func(args ...any) { assert.FailNow(t, "unexpected call to Info") }
	m.WarnFn = func(args ...any) { assert.FailNow(t, "unexpected call to Warn") }
	m.ErrorFn = func(args ...any) { assert.FailNow(t, "unexpected call to Error") }
	m.PanicFn = func(args ...any) { assert.FailNow(t, "unexpected call to Panic") }
}

func (m *mockLogger) Level() log.Level {
	return m.LevelFn()
}

func (m *mockLogger) WithField(key string, value any) log.Logger {
	return m.WithFieldFn(key, value)
}

func (m *mockLogger) WithFields(fields log.Fields) log.Logger {
	return m.WithFieldsFn(fields)
}

func (m *mockLogger) WithError(err error) log.Logger {
	return m.WithErrorFn(err)
}

func (m *mockLogger) WithAdvice(advice string) log.Logger {
	return m.WithAdviceFn(advice)
}

func (m *mockLogger) Debug(args ...any) {
	m.DebugFn(args...)
}

func (m *mockLogger) Info(args ...any) {
	m.InfoFn(args...)

}

func (m *mockLogger) Warn(args ...any) {
	m.WarnFn(args...)
}

func (m *mockLogger) Error(args ...any) {
	m.ErrorFn(args...)
}

func (m *mockLogger) Panic(args ...any) {
	m.PanicFn(args...)
}

type mockTransport struct {
	BroadcastFn func(topic string, message transport.Message) error
	MessagesFn  func(topic string) <-chan transport.ReceivedMessage
}

func newMockTransport(t *testing.T) *mockTransport {
	mt := &mockTransport{}
	mt.reset(t)
	return mt
}

func (m *mockTransport) reset(t *testing.T) {
	m.BroadcastFn = func(topic string, message transport.Message) error {
		assert.FailNow(t, "unexpected call to Broadcast")
		return nil
	}
	m.MessagesFn = func(topic string) <-chan transport.ReceivedMessage {
		assert.FailNow(t, "unexpected call to Messages")
		return nil
	}
}

func (m *mockTransport) Start(ctx context.Context) error {
	return nil
}

func (m *mockTransport) Wait() <-chan error {
	return nil
}

func (m *mockTransport) Broadcast(topic string, message transport.Message) error {
	return m.BroadcastFn(topic, message)
}

func (m *mockTransport) Messages(topic string) <-chan transport.ReceivedMessage {
	return m.MessagesFn(topic)
}

type mockMedianContract struct {
	AddressFn func() types.Address
	ValFn     func(ctx context.Context) (*bn.DecFixedPointNumber, error)
	AgeFn     func(ctx context.Context) (time.Time, error)
	BarFn     func(ctx context.Context) (int, error)
	WatFn     func(ctx context.Context) (string, error)
	PokeFn    func(ctx context.Context, vals []contract.MedianVal) (*types.Hash, *types.Transaction, error)
}

func newMockMedianContract(t *testing.T) *mockMedianContract {
	mc := &mockMedianContract{}
	mc.reset(t)
	return mc
}

func (m *mockMedianContract) reset(t *testing.T) {
	m.AddressFn = func() types.Address {
		assert.FailNow(t, "unexpected call to Address")
		return types.Address{}
	}
	m.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
		assert.FailNow(t, "unexpected call to Val")
		return nil, nil
	}
	m.AgeFn = func(ctx context.Context) (time.Time, error) {
		assert.FailNow(t, "unexpected call to Age")
		return time.Time{}, nil
	}
	m.BarFn = func(ctx context.Context) (int, error) {
		assert.FailNow(t, "unexpected call to Bar")
		return 0, nil
	}
	m.WatFn = func(ctx context.Context) (string, error) {
		assert.FailNow(t, "unexpected call to Wat")
		return "", nil
	}
	m.PokeFn = func(ctx context.Context, vals []contract.MedianVal) (*types.Hash, *types.Transaction, error) {
		assert.FailNow(t, "unexpected call to Poke")
		return nil, nil, nil
	}
}

func (m *mockMedianContract) Address() types.Address {
	return m.AddressFn()
}

func (m *mockMedianContract) Val(ctx context.Context) (*bn.DecFixedPointNumber, error) {
	return m.ValFn(ctx)
}

func (m *mockMedianContract) Age(ctx context.Context) (time.Time, error) {
	return m.AgeFn(ctx)
}

func (m *mockMedianContract) Bar(ctx context.Context) (int, error) {
	return m.BarFn(ctx)
}

func (m *mockMedianContract) Wat(ctx context.Context) (string, error) {
	return m.WatFn(ctx)
}

func (m *mockMedianContract) Poke(ctx context.Context, vals []contract.MedianVal) (*types.Hash, *types.Transaction, error) {
	return m.PokeFn(ctx, vals)
}

// Assume mockScribeContract, mockMuSigStore, and mockTicker are similar to the mock structures used in previous tests.
type mockScribeContract struct {
	AddressFn func() types.Address
	WatFn     func(ctx context.Context) (string, error)
	BarFn     func(ctx context.Context) (int, error)
	FeedsFn   func(ctx context.Context) ([]types.Address, []uint8, error)
	ReadFn    func(ctx context.Context) (contract.PokeData, error)
	PokeFn    func(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData) (*types.Hash, *types.Transaction, error)
}

func newMockScribeContract(t *testing.T) *mockScribeContract {
	sc := &mockScribeContract{}
	sc.reset(t)
	return sc
}

func (m *mockScribeContract) reset(t *testing.T) {
	m.AddressFn = func() types.Address {
		assert.FailNow(t, "unexpected call to Address")
		return types.Address{}
	}
	m.WatFn = func(ctx context.Context) (string, error) {
		assert.FailNow(t, "unexpected call to Wat")
		return "", nil
	}
	m.BarFn = func(ctx context.Context) (int, error) {
		assert.FailNow(t, "unexpected call to Bar")
		return 0, nil
	}
	m.FeedsFn = func(ctx context.Context) ([]types.Address, []uint8, error) {
		assert.FailNow(t, "unexpected call to Feeds")
		return nil, nil, nil
	}
	m.ReadFn = func(ctx context.Context) (contract.PokeData, error) {
		assert.FailNow(t, "unexpected call to Read")
		return contract.PokeData{}, nil
	}
	m.PokeFn = func(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData) (*types.Hash, *types.Transaction, error) {
		assert.FailNow(t, "unexpected call to Poke")
		return nil, nil, nil
	}
}

func (m *mockScribeContract) Address() types.Address {
	return m.AddressFn()
}

func (m *mockScribeContract) Wat(ctx context.Context) (string, error) {
	return m.WatFn(ctx)
}

func (m *mockScribeContract) Bar(ctx context.Context) (int, error) {
	return m.BarFn(ctx)
}

func (m *mockScribeContract) Feeds(ctx context.Context) ([]types.Address, []uint8, error) {
	return m.FeedsFn(ctx)
}

func (m *mockScribeContract) Read(ctx context.Context) (contract.PokeData, error) {
	return m.ReadFn(ctx)
}

func (m *mockScribeContract) Poke(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData) (*types.Hash, *types.Transaction, error) {
	return m.PokeFn(ctx, pokeData, schnorrData)
}

type mockOpScribeContract struct {
	mockScribeContract
	OpPokeFn func(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData, ecdsaData types.Signature) (*types.Hash, *types.Transaction, error)
}

func newMockOpScribeContract(t *testing.T) *mockOpScribeContract {
	sc := &mockOpScribeContract{}
	sc.reset(t)
	return sc
}

func (m *mockOpScribeContract) reset(t *testing.T) {
	m.mockScribeContract.reset(t)
	m.OpPokeFn = func(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData, ecdsaData types.Signature) (*types.Hash, *types.Transaction, error) {
		assert.FailNow(t, "unexpected call to OpPoke")
		return nil, nil, nil
	}
}

func (m *mockOpScribeContract) OpPoke(ctx context.Context, pokeData contract.PokeData, schnorrData contract.SchnorrData, ecdsaData types.Signature) (*types.Hash, *types.Transaction, error) {
	return m.OpPokeFn(ctx, pokeData, schnorrData, ecdsaData)
}

type mockDataPointProvider struct {
	LatestFromFn func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error)
	LatestFn     func(ctx context.Context, model string) (map[types.Address]store.StoredDataPoint, error)
}

func newMockDataPointProvider(t *testing.T) *mockDataPointProvider {
	dp := &mockDataPointProvider{}
	dp.reset(t)
	return dp
}

func (m *mockDataPointProvider) reset(t *testing.T) {
	m.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
		assert.FailNow(t, "unexpected call to LatestFrom")
		return store.StoredDataPoint{}, false, nil
	}
	m.LatestFn = func(ctx context.Context, model string) (map[types.Address]store.StoredDataPoint, error) {
		assert.FailNow(t, "unexpected call to Latest")
		return nil, nil
	}
}

func (m *mockDataPointProvider) LatestFrom(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
	return m.LatestFromFn(ctx, from, model)
}

func (m *mockDataPointProvider) Latest(ctx context.Context, model string) (map[types.Address]store.StoredDataPoint, error) {
	return m.LatestFn(ctx, model)
}

type mockSignatureProvider struct {
	SignaturesByDataModelFn func(model string) []*messages.MuSigSignature
}

func newMockSignatureProvider(t *testing.T) *mockSignatureProvider {
	sp := &mockSignatureProvider{}
	sp.reset(t)
	return sp
}

func (m *mockSignatureProvider) reset(t *testing.T) {
	m.SignaturesByDataModelFn = func(model string) []*messages.MuSigSignature {
		assert.FailNow(t, "unexpected call to SignaturesByDataModel")
		return nil
	}
}

func (m *mockSignatureProvider) SignaturesByDataModel(model string) []*messages.MuSigSignature {
	return m.SignaturesByDataModelFn(model)
}
