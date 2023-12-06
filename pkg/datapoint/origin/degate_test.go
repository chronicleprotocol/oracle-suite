package origin

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
)

func TestDeGate_FetchDataPoints(t *testing.T) {
	testCases := []struct {
		name              string
		pair              value.Pair
		tokenListResponse string
		ticker24Response  string
		expectedResult    map[value.Pair]datapoint.Point
	}{
		{
			name:              "Success",
			pair:              value.Pair{Base: "USDM", Quote: "USDC"},
			tokenListResponse: "{\"code\":0,\"data\":[{\"id\":2,\"base_token_id\":0,\"quote_token_id\":0,\"chain\":\"ETH\",\"code\":\"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48\",\"symbol\":\"USDC\",\"name\":\"USD Coin\",\"icon\":\"https://mainnet-cdn.degate.com/token/USDC.png\",\"decimals\":6,\"is_trusted_token\":true,\"is_quotable_token\":true,\"is_gas_token\":true,\"is_list_token\":true,\"active\":true,\"is_dynamic\":false,\"is_transfer_deposit\":true,\"show_decimal\":8,\"priority\":999900,\"list_priority\":999800,\"gas_priority\":999800,\"is_black\":false},{\"id\":58,\"base_token_id\":0,\"quote_token_id\":0,\"chain\":\"ETH\",\"code\":\"0x59d9356e565ab3a36dd77763fc0d87feaf85508c\",\"symbol\":\"USDM\",\"name\":\"Mountain Protocol USD\",\"icon\":\"https://v1-mainnet-cdn.degate.com/files/token/usdm_0x59d9356e565ab3a36dd77763fc0d87feaf85508c1697432749980.jpg\",\"decimals\":18,\"is_trusted_token\":false,\"is_quotable_token\":true,\"is_gas_token\":true,\"is_list_token\":true,\"active\":true,\"is_dynamic\":false,\"is_transfer_deposit\":true,\"show_decimal\":8,\"priority\":999750,\"list_priority\":999450,\"gas_priority\":0,\"is_black\":false}]}",
			ticker24Response:  "{\"code\":0,\"data\":{\"base_token_id\":58,\"base_token\":{\"token_id\":58,\"chain\":\"ETH\",\"code\":\"0x59d9356e565ab3a36dd77763fc0d87feaf85508c\",\"symbol\":\"USDM\",\"name\":\"Mountain Protocol USD\",\"icon\":\"https://v1-mainnet-cdn.degate.com/files/token/usdm_0x59d9356e565ab3a36dd77763fc0d87feaf85508c1697432749980.jpg\",\"decimals\":18,\"volume\":\"\",\"show_decimals\":8,\"is_quotable_token\":true,\"is_gas_token\":true,\"is_list_token\":true,\"active\":true,\"is_trusted_token\":false,\"is_dynamic\":false,\"is_transfer_deposit\":true,\"priority\":999750,\"list_priority\":999450,\"gas_priority\":0,\"is_black\":false},\"quote_token_id\":2,\"quote_token\":{\"token_id\":2,\"chain\":\"ETH\",\"code\":\"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48\",\"symbol\":\"USDC\",\"name\":\"USD Coin\",\"icon\":\"https://mainnet-cdn.degate.com/token/USDC.png\",\"decimals\":6,\"volume\":\"\",\"show_decimals\":8,\"is_quotable_token\":true,\"is_gas_token\":true,\"is_list_token\":true,\"active\":true,\"is_trusted_token\":true,\"is_dynamic\":false,\"is_transfer_deposit\":true,\"priority\":999900,\"list_priority\":999800,\"gas_priority\":999800,\"is_black\":false},\"pair_id\":35,\"price_change\":\"0.0001\",\"price_change_percent\":\"0.01\",\"weighted_avg_price\":\"0.9999\",\"prev_close_price\":\"1\",\"last_price\":\"1\",\"last_qty\":\"104.48\",\"bid_price\":\"0.9999\",\"bid_qty\":\"51887.3128\",\"ask_price\":\"1\",\"ask_qty\":\"476819.133\",\"open_price\":\"0.9999\",\"high_price\":\"1\",\"low_price\":\"0.9999\",\"volume\":\"1104.477\",\"quote_volume\":\"1104.377\",\"open_time\":0,\"close_time\":1701103573084,\"first_id\":\"53919893700859292120094714781079606621102837344689782890071465132034\",\"last_id\":\"53919893700859345896047460393090271041309254551297433092689301078071\",\"count\":2,\"week_high_price\":\"1\",\"week_low_price\":\"0.9999\",\"base_token_price\":\"1\",\"quote_token_price\":\"1\",\"base_token_risk_price\":\"1\",\"quote_token_risk_price\":\"1\",\"is_stable\":true,\"maker_fee\":\"0\",\"taker_fee\":\"1\",\"amount\":\"1104.3770\"}}",
			expectedResult: map[value.Pair]datapoint.Point{
				value.Pair{Base: "USDM", Quote: "USDC"}: {
					Value: value.NewTick(value.Pair{Base: "USDM", Quote: "USDC"}, 1, 1104.477),
				},
			},
		},
	}

	ctx := context.Background()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "/tokenList") {
					fmt.Fprintln(w, tt.tokenListResponse)
				} else {
					fmt.Fprintln(w, tt.ticker24Response)
				}
			}))
			defer server.Close()

			// Create DeGate Origin
			degate, err := NewDeGate(DeGateConfig{
				Endpoint: server.URL,
			})
			require.NoError(t, err)

			pairs := []any{tt.pair}
			points, err := degate.FetchDataPoints(ctx, pairs)

			for pair, dataPoint := range points {
				valuePair := pair.(value.Pair)
				if tt.expectedResult[valuePair].Error != nil {
					assert.EqualError(t, dataPoint.Error, tt.expectedResult[valuePair].Error.Error())
				} else {
					assert.NoError(t, dataPoint.Error)
				}
				if tt.expectedResult[valuePair].Value != nil {
					assert.Equal(t, tt.expectedResult[valuePair].Value.(value.Tick).Pair, dataPoint.Value.(value.Tick).Pair)
					assert.Equal(t, tt.expectedResult[valuePair].Value.(value.Tick).Price.String(), dataPoint.Value.(value.Tick).Price.String())
					assert.Equal(t, tt.expectedResult[valuePair].Value.(value.Tick).Volume24h.String(), dataPoint.Value.(value.Tick).Volume24h.String())
				} else {
					assert.Nil(t, dataPoint.Value)
				}
			}
		})
	}
}
