package origin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestGenericHTTP_FetchTicks(t *testing.T) {
	testCases := []struct {
		name           string
		pairs          []provider.Pair
		options        GenericHTTPOptions
		expectedResult []provider.Tick
		expectedURLs   []string
	}{
		{
			name:  "simple test",
			pairs: []provider.Pair{{Base: "BTC", Quote: "USD"}},
			options: GenericHTTPOptions{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []provider.Pair, data io.Reader) []provider.Tick {
					return []provider.Tick{
						{
							Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
							Price:     bn.Float(1000),
							Volume24h: bn.Float(100),
							Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
					}
				},
			},
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
			expectedURLs: []string{"/?base=BTC&quote=USD"},
		},
		{
			name:  "one url for all pairs",
			pairs: []provider.Pair{{Base: "BTC", Quote: "USD"}, {Base: "ETH", Quote: "USD"}},
			options: GenericHTTPOptions{
				URL: "/ticks",
				Callback: func(ctx context.Context, pairs []provider.Pair, data io.Reader) []provider.Tick {
					return []provider.Tick{
						{
							Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
							Price:     bn.Float(1000),
							Volume24h: bn.Float(100),
							Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
						{
							Pair:      provider.Pair{Base: "ETH", Quote: "USD"},
							Price:     bn.Float(2000),
							Volume24h: bn.Float(200),
							Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
					}
				},
			},
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
				{
					Pair:      provider.Pair{Base: "ETH", Quote: "USD"},
					Price:     bn.Float(2000),
					Volume24h: bn.Float(200),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
			expectedURLs: []string{"/ticks"},
		},
		{
			name:  "one url per pair",
			pairs: []provider.Pair{{Base: "BTC", Quote: "USD"}, {Base: "ETH", Quote: "USD"}},
			options: GenericHTTPOptions{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []provider.Pair, data io.Reader) []provider.Tick {
					if len(pairs) != 1 {
						t.Fatal("expected one pair")
					}
					switch pairs[0].String() {
					case "BTC/USD":
						return []provider.Tick{
							{
								Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
								Price:     bn.Float(1000),
								Volume24h: bn.Float(100),
								Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
							},
						}
					case "ETH/USD":
						return []provider.Tick{
							{
								Pair:      provider.Pair{Base: "ETH", Quote: "USD"},
								Price:     bn.Float(2000),
								Volume24h: bn.Float(200),
								Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
							},
						}
					}
					return nil
				},
			},
			expectedResult: []provider.Tick{
				{
					Pair:      provider.Pair{Base: "BTC", Quote: "USD"},
					Price:     bn.Float(1000),
					Volume24h: bn.Float(100),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
				{
					Pair:      provider.Pair{Base: "ETH", Quote: "USD"},
					Price:     bn.Float(2000),
					Volume24h: bn.Float(200),
					Time:      time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
			expectedURLs: []string{"/?base=BTC&quote=USD", "/?base=ETH&quote=USD"},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server.
			var requests []*http.Request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r)
			}))
			defer server.Close()

			// Create the provider.
			tt.options.URL = server.URL + tt.options.URL
			gh, err := NewGenericHTTP(tt.options)
			require.NoError(t, err)

			// Test the provider.
			ticks := gh.FetchTicks(context.Background(), tt.pairs)
			require.Len(t, requests, len(tt.expectedURLs))
			for i, url := range tt.expectedURLs {
				assert.Equal(t, url, requests[i].URL.String())
			}
			for i, tick := range ticks {
				assert.Equal(t, tt.expectedResult[i].Pair, tick.Pair)
				assert.Equal(t, tt.expectedResult[i].Price, tick.Price)
				assert.Equal(t, tt.expectedResult[i].Volume24h, tick.Volume24h)
				assert.Equal(t, tt.expectedResult[i].Time, tick.Time)
			}
		})
	}
}
