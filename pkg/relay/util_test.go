package relay

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

func TestCalculateSpread(t *testing.T) {
	tests := []struct {
		name     string
		new, old *bn.DecFixedPointNumber
		expected float64
	}{
		{"calculateSpread(150, 100)", bn.DecFixedPoint(150, 18), bn.DecFixedPoint(100, 18), 50},
		{"calculateSpread(50, 100)", bn.DecFixedPoint(50, 18), bn.DecFixedPoint(100, 18), 50},
		{"calculateSpread(100, 100)", bn.DecFixedPoint(100, 18), bn.DecFixedPoint(100, 18), 0},
		{"calculateSpread(100, 0)", bn.DecFixedPoint(100, 18), bn.DecFixedPoint(0, 18), math.Inf(1)},
		{"calculateSpread(-100, 0)", bn.DecFixedPoint(-100, 18), bn.DecFixedPoint(0, 18), math.Inf(1)},
		{"calculateSpread(0, 0)", bn.DecFixedPoint(0, 18), bn.DecFixedPoint(0, 18), math.Inf(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSpread(tt.new, tt.old)
			assert.Equal(t, tt.expected, got)
		})
	}
}
