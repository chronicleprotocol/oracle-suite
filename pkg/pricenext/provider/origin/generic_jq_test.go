package origin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestNewGenericJQ(t *testing.T) {
	t.Run("empty URL", func(t *testing.T) {
		_, err := NewGenericJQ(GenericJQOptions{
			URL:    "",
			Query:  ".price",
			Logger: null.New(),
		})
		assert.EqualError(t, err, "url cannot be empty")
	})
	t.Run("empty query", func(t *testing.T) {
		_, err := NewGenericJQ(GenericJQOptions{
			URL:    "https://example.com",
			Query:  "",
			Logger: null.New(),
		})
		assert.EqualError(t, err, "query must be specified")
	})
	t.Run("invalid query", func(t *testing.T) {
		_, err := NewGenericJQ(GenericJQOptions{
			URL:    "https://example.com",
			Query:  "invalid jq",
			Logger: null.New(),
		})
		assert.Error(t, err)
	})
	t.Run("valid options", func(t *testing.T) {
		_, err := NewGenericJQ(GenericJQOptions{
			URL:    "https://example.com",
			Query:  ".price",
			Logger: null.New(),
		})
		assert.NoError(t, err)
	})
}

func TestGenericJQ_FetchTicks(t *testing.T) {
	testCases := []struct {
		name             string
		query            string
		responseBody     string
		expectedResult   []provider.Tick
		skipVolumeAssert bool
		skipTimeAssert   bool
	}{
		{
			name:         "price, volume and time",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "2023-05-02T12:34:56Z"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "price only",
			query:        "{price: .price}",
			responseBody: `{"price": "1000", "volume": "100", "time": "2023-05-02T12:34:56Z"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: nil,
					Time:      time.Now(),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "single array result",
			query:        ".[] | {price: .price}",
			responseBody: `[{"price": "1000"}]`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: nil,
					Time:      time.Now(),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "multiple array results",
			query:        ".[] | {price: .price}",
			responseBody: `[{"price": "1000"}, {"price": "2000"}]`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     nil,
					Volume24h: nil,
					Time:      time.Now(),
					Error:     fmt.Errorf("multiple results from JQ query"),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "variables $ucbase and $ucquote",
			query:        ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .price}",
			responseBody: `[{"price": "1000", "symbol": "BTCUSD"}, {"price": "2000", "symbol": "ETHUSD"}]`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: nil,
					Time:      time.Now(),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "variables $lcbase and $lcquote",
			query:        ".[] | select(.symbol == ($lcbase + $lcquote)) | {price: .price}",
			responseBody: `[{"price": "1000", "symbol": "btcusd"}, {"price": "2000", "symbol": "ethusd"}]`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: nil,
					Time:      time.Now(),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "invalid JSON",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "2023-05-02T12:34:56Z"`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     nil,
					Volume24h: nil,
					Time:      time.Now(),
					Error:     errors.New("unexpected EOF"),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "empty response",
			query:        ".",
			responseBody: ``,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     nil,
					Volume24h: nil,
					Time:      time.Now(),
					Error:     errors.New("EOF"),
				},
			},
			skipTimeAssert: true,
		},
		{
			name:         "time.RFC3339",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "2023-05-02T12:34:56Z"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.RFC3339Nano",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "2023-05-02T12:34:56.123456789Z"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 123456789, time.UTC),
				},
			},
		},
		{
			name:         "time.RFC1123",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tue, 02 May 2023 12:34:56 UTC"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.RFC1123Z",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tue, 02 May 2023 12:34:56 +0000"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.FixedZone("+0000", 0)),
				},
			},
		},
		{
			name:         "time.RFC822",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "02 May 23 12:34 UTC"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 0, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.RFC822Z",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "02 May 23 12:34 +0000"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 0, 0, time.FixedZone("+0000", 0)),
				},
			},
		},
		{
			name:         "time.RFC850",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tuesday, 02-May-23 12:34:56 UTC"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.ANSIC",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tue May  2 12:34:56 2023"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.UnixDate",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tue May  2 12:34:56 UTC 2023"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
		},
		{
			name:         "time.RubyDate",
			query:        ".",
			responseBody: `{"price": "1000", "volume": "100", "time": "Tue May 02 12:34:56 +0000 2023"}`,
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.FixedZone("+0000", 0)),
				},
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, tt.responseBody)
			}))
			defer server.Close()

			// Create a new GenericJQ provider.
			gjq, err := NewGenericJQ(GenericJQOptions{
				URL:    server.URL,
				Query:  tt.query,
				Logger: null.New(),
			})
			require.NoError(t, err)

			// Test the provider.
			pairs := []provider.Pair{{Base: "BTC", Quote: "USD"}}
			ticks := gjq.FetchTicks(context.Background(), pairs)
			for i, tick := range ticks {
				assert.Equal(t, tt.expectedResult[i].Pair, tick.Pair)
				assert.Equal(t, tt.expectedResult[i].Price, tick.Price)
				if !tt.skipVolumeAssert {
					assert.Equal(t, tt.expectedResult[i].Volume24h, tick.Volume24h)
				}
				if !tt.skipTimeAssert {
					assert.Equal(t, tt.expectedResult[i].Time, tick.Time)
				}
				if tt.expectedResult[i].Error != nil {
					assert.EqualError(t, tick.Error, tt.expectedResult[i].Error.Error())
				} else {
					assert.NoError(t, tick.Error)
				}
			}
		})
	}
}
