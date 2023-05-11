package origin

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestTick_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		dataPoint     Tick
		expectError   bool
		errorContains string
	}{
		{
			name: "valid tick",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Price:     bn.Float(1000),
				Volume24h: bn.Float(100),
			},
			expectError: false,
		},
		{
			name: "pair is not set",
			dataPoint: Tick{
				Price:     bn.Float(1000),
				Volume24h: bn.Float(100),
			},
			expectError:   true,
			errorContains: "pair is not set",
		},
		{
			name: "price is nil",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Volume24h: bn.Float(100),
			},
			expectError:   true,
			errorContains: "price is nil",
		},
		{
			name: "price is zero",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Price:     bn.Float(0),
				Volume24h: bn.Float(100),
			},
			expectError:   true,
			errorContains: "price is zero or negative",
		},
		{
			name: "price is negative",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Price:     bn.Float(-1000),
				Volume24h: bn.Float(100),
			},
			expectError:   true,
			errorContains: "price is zero or negative",
		},
		{
			name: "price is infinite",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Price:     bn.Float(math.Inf(1)),
				Volume24h: bn.Float(100),
			},
			expectError:   true,
			errorContains: "price is infinite",
		},
		{
			name: "volume is negative",
			dataPoint: Tick{
				Pair:      Pair{Base: "BTC", Quote: "USD"},
				Price:     bn.Float(1000),
				Volume24h: bn.Float(-100),
			},
			expectError:   true,
			errorContains: "volume is negative",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dataPoint.Validate()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPair(t *testing.T) {
	testCases := []struct {
		name          string
		pairStr       string
		expectedBase  string
		expectedQuote string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid pair",
			pairStr:       "BTC/USD",
			expectedBase:  "BTC",
			expectedQuote: "USD",
		},
		{
			name:          "valid pair lowercase",
			pairStr:       "btc/usd",
			expectedBase:  "BTC",
			expectedQuote: "USD",
		},
		{
			name:          "invalid pair",
			pairStr:       "BTC-USD",
			expectError:   true,
			errorContains: "pair must be formatted as BASE/QUOTE",
		},
		{
			name:          "empty pair",
			pairStr:       "",
			expectError:   true,
			errorContains: "pair must be formatted as BASE/QUOTE",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := PairFromString(tt.pairStr)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBase, pair.Base)
				assert.Equal(t, tt.expectedQuote, pair.Quote)

				// Test String() method
				assert.Equal(t, fmt.Sprintf("%s/%s", tt.expectedBase, tt.expectedQuote), pair.String())

				// Test Invert() method
				invertedPair := pair.Invert()
				assert.Equal(t, tt.expectedQuote, invertedPair.Base)
				assert.Equal(t, tt.expectedBase, invertedPair.Quote)

				// Test Equal() method
				assert.True(t, pair.Equal(Pair{Base: tt.expectedBase, Quote: tt.expectedQuote}))
			}
		})
	}
}
