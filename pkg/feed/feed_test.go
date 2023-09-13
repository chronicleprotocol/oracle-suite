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

package feed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	dataMocks "github.com/chronicleprotocol/oracle-suite/pkg/datapoint/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

var (
	testSignature = types.MustSignatureFromHex("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00")
	testAddress   = types.MustAddressFromHex("0x00112233445566778899aabbccddeeff00112233")
)

type mockSigner struct{}

func (s mockSigner) Supports(_ context.Context, data datapoint.Point) bool {
	_, ok := data.Value.(value.StaticValue)
	return ok
}

func (s mockSigner) Sign(_ context.Context, _ string, _ datapoint.Point) (*types.Signature, error) {
	return &testSignature, nil
}

func (s mockSigner) Recover(_ context.Context, _ string, _ datapoint.Point, signature types.Signature) (*types.Address, error) {
	if signature != testSignature {
		return nil, errors.New("invalid signature")
	}
	return &testAddress, nil
}

func TestFeed_Broadcast(t *testing.T) {
	tests := []struct {
		name             string
		dataModels       []string
		mocks            func(*dataMocks.Provider)
		asserts          func(t *testing.T, dataPoints []*messages.DataPoint)
		expectedMessages int
	}{
		{
			name:       "valid data point",
			dataModels: []string{"AAABBB"},
			mocks: func(p *dataMocks.Provider) {
				point := datapoint.Point{
					Value: value.StaticValue{Value: bn.DecFloatPoint(42)},
					Time:  time.Unix(100, 0),
				}
				p.On("ModelNames", mock.Anything).Return(
					[]string{"AAABBB"},
				)
				p.On("DataPoints", mock.Anything, []string{"AAABBB"}).Return(
					map[string]datapoint.Point{"AAABBB": point},
					nil,
				)
				p.On("DataPoint", mock.Anything, "AAABBB").Return(
					point,
					nil,
				)
			},
			asserts: func(t *testing.T, dataPoints []*messages.DataPoint) {
				require.Len(t, dataPoints, 1)
				assert.Equal(t, "AAABBB", dataPoints[0].Model)
				assert.Equal(t, "42", dataPoints[0].Point.Value.Print())
				assert.Equal(t, time.Unix(100, 0), dataPoints[0].Point.Time)
				assert.Equal(t, testSignature, dataPoints[0].ECDSASignature)
			},
			expectedMessages: 1,
		},
		{
			name:       "missing data model",
			dataModels: []string{"AAABBB", "CCCDDD"},
			mocks: func(p *dataMocks.Provider) {
				point := datapoint.Point{
					Value: value.StaticValue{Value: bn.DecFloatPoint(42)},
					Time:  time.Unix(100, 0),
				}
				p.On("ModelNames", mock.Anything).Return(
					[]string{"AAABBB", "CCCDDD"},
				)
				p.On("DataPoints", mock.Anything, []string{"AAABBB", "CCCDDD"}).Return(
					map[string]datapoint.Point{"AAABBB": point},
					nil,
				)
				p.On("DataPoint", mock.Anything, "AAABBB").Return(
					point,
					nil,
				)
				p.On("DataPoint", mock.Anything, "CCCDDD").Return(
					datapoint.Point{},
					errors.New("not found"),
				)
			},
			asserts: func(t *testing.T, dataPoints []*messages.DataPoint) {
				require.Len(t, dataPoints, 1)
			},
			expectedMessages: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer ctxCancel()

			// Setup test environment.
			ticker := timeutil.NewTicker(0)
			dataProvider := &dataMocks.Provider{}
			localTransport := local.New([]byte("test"), 0, map[string]transport.Message{
				messages.DataPointV1MessageName: (*messages.DataPoint)(nil),
			})

			// Prepare mocks.
			tt.mocks(dataProvider)

			// Start feed.
			feed, err := New(Config{
				DataModels:   tt.dataModels,
				DataProvider: dataProvider,
				Signers:      []datapoint.Signer{mockSigner{}},
				Transport:    localTransport,
				Interval:     ticker,
			})
			require.NoError(t, err)
			require.NoError(t, localTransport.Start(ctx))
			require.NoError(t, feed.Start(ctx))
			defer func() {
				ctxCancel()
				<-feed.Wait()
				<-localTransport.Wait()
			}()

			// Wait for services to start.
			time.Sleep(time.Millisecond * 100)

			// Trigger a tick manually to get the first message.
			ticker.Tick()

			// Get messages.
			var dataPoints []*messages.DataPoint
			msgCh := localTransport.Messages(messages.DataPointV1MessageName)
			for len(dataPoints) < tt.expectedMessages {
				select {
				case msg := <-msgCh:
					dataPoints = append(dataPoints, msg.Message.(*messages.DataPoint))
				}
			}

			// Check that the broadcasted messages meet the expectations.
			tt.asserts(t, dataPoints)
		})
	}
}

func TestFeed_Start(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer ctxCancel()

	// Setup the test environment.
	dataProvider := &dataMocks.Provider{}
	localTransport := local.New([]byte("test"), 0, map[string]transport.Message{})
	_ = localTransport.Start(ctx)
	defer func() {
		<-localTransport.Wait()
	}()

	// Wait for the transport to start.
	time.Sleep(time.Millisecond * 100)

	// Create a new feed.
	feed, err := New(Config{
		DataModels:   []string{"AAABBB"},
		Signers:      []datapoint.Signer{mockSigner{}},
		DataProvider: dataProvider,
		Transport:    localTransport,
		Interval:     timeutil.NewTicker(time.Second),
	})
	require.NoError(t, err)

	// Try to start the feed without a context, which should fail.
	require.Error(t, feed.Start(nil))

	// Try to start the feed with a context, which should be successful.
	require.NoError(t, feed.Start(ctx))

	// Try to start the feed a second time, which should fail.
	require.Error(t, feed.Start(ctx))

	ctxCancel()
}
