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
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/contract"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/chronicle"
	"github.com/chronicleprotocol/oracle-suite/pkg/contract/mock"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestMedian(t *testing.T) {
	testFeed1 := types.MustAddressFromHex("0x1111111111111111111111111111111111111111")
	testFeed2 := types.MustAddressFromHex("0x2222222222222222222222222222222222222222")
	testFeed3 := types.MustAddressFromHex("0x3333333333333333333333333333333333333333")
	mockLogger := newMockLogger(t)
	mockContract := newMockMedianContract(t)
	mockTransport := newMockTransport(t)
	mockStore := newMockDataPointProvider(t)

	mw := &median{
		contract:       mockContract,
		dataPointStore: mockStore,
		feedAddresses:  []types.Address{testFeed1, testFeed2, testFeed3},
		dataModel:      "ETH/USD",
		spread:         5,
		expiration:     10 * time.Minute,
		log:            mockLogger,
	}

	t.Run("above spread", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(1, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			if from == testFeed1 {
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		pokeCalled := false
		mockContract.PokeFn = func(vals []chronicle.MedianVal) contract.SelfTransactableCaller {
			pokeCalled = true
			assert.Equal(t, 1, len(vals))
			assert.Equal(t, bn.DecFixedPoint(110, chronicle.MedianPricePrecision).String(), vals[0].Val.String())
			assert.Equal(t, uint8(27), vals[0].V)
			assert.Equal(t, big.NewInt(1).String(), vals[0].R.String())
			assert.Equal(t, big.NewInt(2).String(), vals[0].S.String())
			return mock.NewCaller(t).MockAllowAllCalls()
		}

		mw.createRelayCall(ctx)
		assert.True(t, pokeCalled, "poke should have been called")
	})

	t.Run("within spread", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(1, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			if from == testFeed1 {
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(104)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		mw.createRelayCall(ctx)
	})

	t.Run("expired", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-15*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(1, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			if from == testFeed1 {
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(100)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		pokeCalled := false
		mockContract.PokeFn = func(vals []chronicle.MedianVal) contract.SelfTransactableCaller {
			pokeCalled = true
			return mock.NewCaller(t).MockAllowAllCalls()
		}

		mw.createRelayCall(ctx)
		assert.True(t, pokeCalled, "poke should have been called")
	})

	t.Run("median of 2", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(2, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			switch from {
			case testFeed1:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(100)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			case testFeed2:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed2,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(3), big.NewInt(4)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		pokeCalled := false
		mockContract.PokeFn = func(vals []chronicle.MedianVal) contract.SelfTransactableCaller {
			pokeCalled = true
			return mock.NewCaller(t).MockAllowAllCalls()
		}

		mw.createRelayCall(ctx)
		assert.True(t, pokeCalled, "poke should have been called")
	})

	t.Run("median of 3", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(2, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			switch from {
			case testFeed1:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(100)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			case testFeed2:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed2,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(3), big.NewInt(4)),
				}, true, nil
			case testFeed3:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed3,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(5), big.NewInt(6)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		pokeCalled := false
		mockContract.PokeFn = func(vals []chronicle.MedianVal) contract.SelfTransactableCaller {
			pokeCalled = true
			return mock.NewCaller(t).MockAllowAllCalls()
		}

		mw.createRelayCall(ctx)
		assert.True(t, pokeCalled, "poke should have been called")
	})

	t.Run("should use random data points", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-15*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(2, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}

		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			switch from {
			case testFeed1:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(100)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			case testFeed2:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(105)},
					},
					From:      testFeed2,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(3), big.NewInt(4)),
				}, true, nil
			case testFeed3:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed3,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(5), big.NewInt(6)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		maxTries := 1000
		usedPrices := make(map[string]int)
		for i := 0; i < maxTries; i++ {
			mockContract.PokeFn = func(vals []chronicle.MedianVal) contract.SelfTransactableCaller {
				if len(vals) != 2 {
					t.Fatal("poke should have been called with 2 values")
				}
				usedPrices[vals[0].Val.String()+vals[1].Val.String()]++
				return mock.NewCaller(t).MockAllowAllCalls()
			}
			mw.createRelayCall(ctx)
		}

		// This tests verifies that the random data points are used. Because it's based on random
		// values, it's possible that this test fails even though the code is correct, but the
		// probability of that happening is very low.
		assert.Greater(t, usedPrices["100105"], 100)
		assert.Greater(t, usedPrices["100110"], 100)
		assert.Greater(t, usedPrices["105110"], 100)
	})

	t.Run("broken data points", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(3, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.WarnFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockLogger.ErrorFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			switch from {
			case testFeed1:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      testFeed1,
					Signature: types.Signature{},
				}, true, nil
			case testFeed2:
				return store.StoredDataPoint{
					Model:     "ETH/USD",
					DataPoint: datapoint.Point{},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			case testFeed3:
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.Tick{Price: bn.DecFloatPoint(110)},
					},
					From:      types.Address{},
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		mw.createRelayCall(ctx)
	})

	t.Run("not a tick", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("ETH/USD", nil)
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(1, nil)
		}
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockLogger.WarnFn = func(args ...any) {}
		mockLogger.ErrorFn = func(args ...any) {}
		mockStore.LatestFromFn = func(ctx context.Context, from types.Address, model string) (store.StoredDataPoint, bool, error) {
			if from == testFeed1 {
				return store.StoredDataPoint{
					Model: "ETH/USD",
					DataPoint: datapoint.Point{
						Time:  time.Now(),
						Value: value.StaticValue{Value: bn.DecFloatPoint(110)},
					},
					From:      testFeed1,
					Signature: types.SignatureFromVRS(big.NewInt(27), big.NewInt(1), big.NewInt(2)),
				}, true, nil
			}
			return store.StoredDataPoint{}, false, nil
		}

		mw.createRelayCall(ctx)
	})

	t.Run("call error", func(t *testing.T) {
		mockLogger.reset(t)
		mockContract.reset(t)
		mockTransport.reset(t)

		errLogCalled := false
		mockLogger.InfoFn = func(args ...any) {}
		mockLogger.DebugFn = func(args ...any) {}
		mockLogger.ErrorFn = func(args ...any) { errLogCalled = true }

		ctx := context.Background()
		mockContract.ClientFn = func() rpc.RPC { return nil }
		mockContract.AddressFn = func() types.Address { return types.Address{} }
		mockContract.WatFn = func() contract.TypedSelfCaller[string] {
			return mock.NewTypedCaller[string](t).MockResult("", errors.New("error"))
		}
		mockContract.ValFn = func(ctx context.Context) (*bn.DecFixedPointNumber, error) {
			return bn.DecFixedPoint(100, chronicle.MedianPricePrecision), nil
		}
		mockContract.AgeFn = func() contract.TypedSelfCaller[time.Time] {
			return mock.NewTypedCaller[time.Time](t).MockResult(time.Now().Add(-1*time.Minute), nil)
		}
		mockContract.BarFn = func() contract.TypedSelfCaller[int] {
			return mock.NewTypedCaller[int](t).MockResult(1, nil)
		}

		mw.createRelayCall(ctx)
		assert.True(t, errLogCalled)
	})
}
