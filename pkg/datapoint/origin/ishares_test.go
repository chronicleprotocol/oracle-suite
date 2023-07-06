package origin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIShares_FetchDataPoints(t *testing.T) {
	testCases := []struct {
		name             string
		responseBody     string
		expectedResult   map[any]datapoint.Point
		skipVolumeAssert bool
		skipTimeAssert   bool
	}{
		{
			name: "Success",
			responseBody: `<html><body><li class="navAmount " data-col="fundHeader.fundNav.navAmount" data-path="">
<span class="header-nav-label navAmount">
NAV as of 09/Jan/2023
</span>
<span class="header-nav-data">
USD 5.43
</span>
<span class="header-info-bubble">
</span>
<br>
<span class="fiftyTwoWeekData">
52 WK: 5.11 - 5.37
</span>
</li></body></html>`,
			expectedResult: map[any]datapoint.Point{
				value.Pair{Base: "IBTA", Quote: "USD"}: {
					Value: value.Tick{
						Pair:  value.Pair{Base: "IBTA", Quote: "USD"},
						Price: bn.Float(5.43),
					},
					Time: time.Now(),
				},
			},
			skipTimeAssert: true,
		},
	}

	ctx := context.Background()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, tt.responseBody)
			}))
			defer server.Close()

			// Create IShare Origin
			ishares, err := NewIShares(ISharesConfig{
				URL:    server.URL,
				Logger: null.New(),
			})
			require.NoError(t, err)

			// Test the cases
			pairs := []any{value.Pair{Base: "IBTA", Quote: "USD"}}
			points, err := ishares.FetchDataPoints(ctx, pairs)
			require.NoError(t, err)

			for pair, dataPoint := range points {
				if tt.expectedResult[pair].Value != nil {
					assert.Equal(t, tt.expectedResult[pair].Value.(value.Tick).Pair, dataPoint.Value.(value.Tick).Pair)
					assert.Equal(t, tt.expectedResult[pair].Value.(value.Tick).Price, dataPoint.Value.(value.Tick).Price)
					if !tt.skipVolumeAssert {
						assert.Equal(t, tt.expectedResult[pair].Value.(value.Tick).Volume24h, dataPoint.Value.(value.Tick).Volume24h)
					}
				} else {
					assert.Nil(t, dataPoint.Value)
				}
				if !tt.skipTimeAssert {
					assert.Equal(t, tt.expectedResult[pair].Time, dataPoint.Time)
				}
				if tt.expectedResult[pair].Error != nil {
					assert.EqualError(t, dataPoint.Error, tt.expectedResult[pair].Error.Error())
				} else {
					assert.NoError(t, dataPoint.Error)
				}
			}
		})
	}
}
