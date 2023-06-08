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

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestGenericHTTP_FetchDataPoints(t *testing.T) {
	testCases := []struct {
		name           string
		pairs          []any
		options        TickGenericHTTPOptions
		expectedResult []datapoint.Point
		expectedURLs   []string
	}{
		{
			name:  "simple test",
			pairs: []any{value.Pair{Base: "BTC", Quote: "USD"}},
			options: TickGenericHTTPOptions{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) map[any]datapoint.Point {
					return map[any]datapoint.Point{
						value.Pair{Base: "BTC", Quote: "USD"}: {
							Value: value.Tick{
								Pair:      value.Pair{Base: "BTC", Quote: "USD"},
								Price:     bn.Float(1000),
								Volume24h: bn.Float(100),
							},
							Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
					}
				},
			},
			expectedResult: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "BTC", Quote: "USD"},
						Price:     bn.Float(1000),
						Volume24h: bn.Float(100),
					},
					Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
			expectedURLs: []string{"/?base=BTC&quote=USD"},
		},
		{
			name:  "one url for all pairs",
			pairs: []any{value.Pair{Base: "BTC", Quote: "USD"}, value.Pair{Base: "ETH", Quote: "USD"}},
			options: TickGenericHTTPOptions{
				URL: "/dataPoints",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) map[any]datapoint.Point {
					return map[any]datapoint.Point{
						value.Pair{Base: "BTC", Quote: "USD"}: {
							Value: value.Tick{
								Pair:      value.Pair{Base: "BTC", Quote: "USD"},
								Price:     bn.Float(1000),
								Volume24h: bn.Float(100),
							},
							Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
						value.Pair{Base: "ETH", Quote: "USD"}: {
							Value: value.Tick{
								Pair:      value.Pair{Base: "ETH", Quote: "USD"},
								Price:     bn.Float(2000),
								Volume24h: bn.Float(200),
							},
							Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
					}
				},
			},
			expectedResult: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "BTC", Quote: "USD"},
						Price:     bn.Float(1000),
						Volume24h: bn.Float(100),
					},
					Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "ETC", Quote: "USD"},
						Price:     bn.Float(2000),
						Volume24h: bn.Float(200),
					},
					Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
			},
			expectedURLs: []string{"/dataPoints"},
		},
		{
			name:  "one url per pair",
			pairs: []any{value.Pair{Base: "BTC", Quote: "USD"}, value.Pair{Base: "ETH", Quote: "USD"}},
			options: TickGenericHTTPOptions{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) map[any]datapoint.Point {
					if len(pairs) != 1 {
						t.Fatal("expected one pair")
					}
					switch pairs[0].String() {
					case "BTC/USD":
						return map[any]datapoint.Point{
							value.Pair{Base: "BTC", Quote: "USD"}: {
								Value: value.Tick{
									Pair:      value.Pair{Base: "BTC", Quote: "USD"},
									Price:     bn.Float(1000),
									Volume24h: bn.Float(100),
								},
								Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
							},
						}
					case "ETH/USD":
						return map[any]datapoint.Point{
							value.Pair{Base: "ETH", Quote: "USD"}: {
								Value: value.Tick{
									Pair:      value.Pair{Base: "ETC", Quote: "USD"},
									Price:     bn.Float(2000),
									Volume24h: bn.Float(200),
								},
								Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
							},
						}
					}
					return nil
				},
			},
			expectedResult: []datapoint.Point{
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "BTC", Quote: "USD"},
						Price:     bn.Float(1000),
						Volume24h: bn.Float(100),
					},
					Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
				},
				{
					Value: value.Tick{
						Pair:      value.Pair{Base: "ETC", Quote: "USD"},
						Price:     bn.Float(2000),
						Volume24h: bn.Float(200),
					},
					Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
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

			// Create the data.
			tt.options.URL = server.URL + tt.options.URL
			gh, err := NewTickGenericHTTP(tt.options)
			require.NoError(t, err)

			// Test the data.
			points, err := gh.FetchDataPoints(context.Background(), tt.pairs)
			require.NoError(t, err)
			require.Len(t, requests, len(tt.expectedURLs))
			for i, url := range tt.expectedURLs {
				assert.Equal(t, url, requests[i].URL.String())
			}
			_ = points
			/*
				for i, dataPoint := range points {
					assert.Equal(t, tt.expectedResult[i].Pair, dataPoint.Pair)
					assert.Equal(t, tt.expectedResult[i].Price, dataPoint.Price)
					assert.Equal(t, tt.expectedResult[i].Volume24h, dataPoint.Volume24h)
					assert.Equal(t, tt.expectedResult[i].Time, dataPoint.Time)
				}

			*/
		})
	}
}
