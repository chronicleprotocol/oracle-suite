package origin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
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
		options        TickGenericHTTPConfig
		expectedResult map[any]datapoint.Point
		expectedURLs   []string
	}{
		{
			name:  "simple test",
			pairs: []any{value.Pair{Base: "BTC", Quote: "USD"}},
			options: TickGenericHTTPConfig{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) (map[any]datapoint.Point, error) {
					return map[any]datapoint.Point{
						value.Pair{Base: "BTC", Quote: "USD"}: {
							Value: value.Tick{
								Pair:      value.Pair{Base: "BTC", Quote: "USD"},
								Price:     bn.Float(1000),
								Volume24h: bn.Float(100),
							},
							Time: time.Date(2023, 5, 2, 12, 34, 56, 0, time.UTC),
						},
					}, nil
				},
			},
			expectedResult: map[any]datapoint.Point{
				value.Pair{Base: "BTC", Quote: "USD"}: {
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
			options: TickGenericHTTPConfig{
				URL: "/dataPoints",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) (map[any]datapoint.Point, error) {
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
					}, nil
				},
			},
			expectedResult: map[any]datapoint.Point{
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
			},
			expectedURLs: []string{"/dataPoints"},
		},
		{
			name:  "one url per pair",
			pairs: []any{value.Pair{Base: "BTC", Quote: "USD"}, value.Pair{Base: "ETH", Quote: "USD"}},
			options: TickGenericHTTPConfig{
				URL: "/?base=${ucbase}&quote=${ucquote}",
				Callback: func(ctx context.Context, pairs []value.Pair, body io.Reader) (map[any]datapoint.Point, error) {
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
						}, nil
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
						}, nil
					}
					return nil, nil
				},
			},
			expectedResult: map[any]datapoint.Point{
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
			// Because of changing order in `requests` sometimes, the assertion of comparing urls may have a break.
			sort.Slice(requests, func(i, j int) bool {
				return requests[i].URL.String() < requests[j].URL.String()
			})
			for i, url := range tt.expectedURLs {
				assert.Equal(t, url, requests[i].URL.String())
			}
			for i, dataPoint := range points {
				assert.Equal(t, tt.expectedResult[i].Value.Print(), dataPoint.Value.Print())
				assert.Equal(t, tt.expectedResult[i].Time, dataPoint.Time)
			}
		})
	}
}
