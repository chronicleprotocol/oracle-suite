package feed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	dataMocks "github.com/chronicleprotocol/oracle-suite/pkg/data/mocks"
	ethereumMocks "github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) Supports(point data.Point) bool {
	args := m.Called(point)
	return args.Bool(0)
}

func (m *mockHandler) Handle(model string, point data.Point) (*messages.Event, error) {
	args := m.Called(model, point)
	return args.Get(0).(*messages.Event), args.Error(1)
}

type pointValue struct {
	value string
}

func (p pointValue) Print() string {
	return p.value
}

func TestFeeder_Broadcast(t *testing.T) {
	tests := []struct {
		name             string
		dataModels       []string
		mocks            func(*dataMocks.Provider, *mockHandler, *ethereumMocks.Key)
		asserts          func(t *testing.T, events []*messages.Event)
		expectedMessages int
	}{
		{
			name:       "valid data point",
			dataModels: []string{"AAABBB"},
			mocks: func(p *dataMocks.Provider, h *mockHandler, s *ethereumMocks.Key) {
				point := data.Point{
					Value: pointValue{value: "foo"},
					Time:  time.Unix(100, 0),
				}
				p.On("DataPoints", mock.Anything, []string{"AAABBB"}).Return(
					map[string]data.Point{"AAABBB": point},
					nil,
				)
				p.On("DataPoint", mock.Anything, "AAABBB").Return(
					point,
					nil,
				)
				h.On("Supports", point).Return(true)
				h.On("Handle", "AAABBB", point).Return(
					&messages.Event{Type: "event"},
					nil,
				)
			},
			asserts: func(t *testing.T, events []*messages.Event) {
				require.Len(t, events, 1)
				require.Equal(t, "event", events[0].Type)
			},
			expectedMessages: 1,
		},
		{
			name:       "missing data model",
			dataModels: []string{"AAABBB", "CCCDDD"},
			mocks: func(p *dataMocks.Provider, h *mockHandler, s *ethereumMocks.Key) {
				point := data.Point{
					Value: pointValue{value: "foo"},
					Time:  time.Unix(100, 0),
				}
				p.On("DataPoints", mock.Anything, []string{"AAABBB", "CCCDDD"}).Return(
					map[string]data.Point{"AAABBB": point},
					nil,
				)
				p.On("DataPoint", mock.Anything, "AAABBB").Return(
					point,
					nil,
				)
				p.On("DataPoint", mock.Anything, "CCCDDD").Return(
					data.Point{},
					errors.New("not found"),
				)
				h.On("Supports", point).Return(true)

				// Even if one of the data models is missing, the other one should be processed.
				h.On("Handle", "AAABBB", point).Return(
					&messages.Event{Type: "event"},
					nil,
				)
			},
			asserts: func(t *testing.T, events []*messages.Event) {
				require.Len(t, events, 1)
				require.Equal(t, "event", events[0].Type)
			},
			expectedMessages: 1,
		},
		{
			name:       "unsupported data point",
			dataModels: []string{"AAABBB"},
			mocks: func(p *dataMocks.Provider, h *mockHandler, s *ethereumMocks.Key) {
				point := data.Point{
					Value: pointValue{value: "foo"},
					Time:  time.Unix(100, 0),
				}
				p.On("DataPoints", mock.Anything, []string{"AAABBB"}).Return(
					map[string]data.Point{"AAABBB": point},
					nil,
				)
				p.On("DataPoint", mock.Anything, "AAABBB").Return(
					point,
					nil,
				)
				h.On("Supports", point).Return(false)
			},
			asserts: func(t *testing.T, events []*messages.Event) {
				require.Len(t, events, 0)
			},
			expectedMessages: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
			defer ctxCancel()

			// Prepare mocks.
			dataProvider := &dataMocks.Provider{}
			pointHandler := &mockHandler{}
			signer := &ethereumMocks.Key{}

			ticker := timeutil.NewTicker(0)
			localTransport := local.New([]byte("test"), 0, map[string]transport.Message{
				messages.EventV1MessageName: (*messages.Event)(nil),
			})

			// Prepare mocks.
			tt.mocks(dataProvider, pointHandler, signer)

			// Start feeder.
			feeder, err := New(Config{
				DataModels:   tt.dataModels,
				DataProvider: dataProvider,
				Handlers:     []DataPointHandler{pointHandler},
				Transport:    localTransport,
				Interval:     ticker,
			})
			require.NoError(t, err)
			require.NoError(t, localTransport.Start(ctx))
			require.NoError(t, feeder.Start(ctx))
			defer func() {
				ctxCancel()
				<-feeder.Wait()
				<-localTransport.Wait()
			}()

			ticker.Tick()

			// Get messages.
			var events []*messages.Event
			msgCh := localTransport.Messages(messages.EventV1MessageName)
			for len(events) < tt.expectedMessages {
				select {
				case msg := <-msgCh:
					events = append(events, msg.Message.(*messages.Event))
				}
			}

			// Asserts.
			tt.asserts(t, events)
		})
	}
}

func TestFeeder_Start(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer ctxCancel()

	dataProvider := &dataMocks.Provider{}
	localTransport := local.New([]byte("test"), 0, map[string]transport.Message{})
	_ = localTransport.Start(ctx)
	defer func() {
		<-localTransport.Wait()
	}()

	gho, err := New(Config{
		DataModels:   []string{},
		DataProvider: dataProvider,
		Transport:    localTransport,
		Interval:     timeutil.NewTicker(time.Second),
	})
	require.NoError(t, err)
	require.Error(t, gho.Start(nil)) // Start without context should fail.
	require.NoError(t, gho.Start(ctx))
	require.Error(t, gho.Start(ctx)) // Second start should fail.
	ctxCancel()
}
