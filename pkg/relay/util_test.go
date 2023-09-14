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
		new, old *bn.DecFloatPointNumber
		expected float64
	}{
		{"calculateSpread(150, 100)", bn.DecFloatPoint(150), bn.DecFloatPoint(100), 50},
		{"calculateSpread(50, 100)", bn.DecFloatPoint(50), bn.DecFloatPoint(100), 50},
		{"calculateSpread(100, 100)", bn.DecFloatPoint(100), bn.DecFloatPoint(100), 0},
		{"calculateSpread(100, 0)", bn.DecFloatPoint(100), bn.DecFloatPoint(0), math.Inf(1)},
		{"calculateSpread(-100, 0)", bn.DecFloatPoint(-100), bn.DecFloatPoint(0), math.Inf(1)},
		{"calculateSpread(0, 0)", bn.DecFloatPoint(0), bn.DecFloatPoint(0), math.Inf(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSpread(tt.new, tt.old)
			assert.Equal(t, tt.expected, got)
		})
	}
}
