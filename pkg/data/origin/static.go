package origin

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/data"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// StaticValue is a numeric value obtained from a static origin.
type StaticValue struct {
	Value *bn.FloatNumber
}

// Number implements the data.NumericValue interface.
func (s StaticValue) Number() *bn.FloatNumber {
	return s.Value
}

// Print implements the data.Value interface.
func (s StaticValue) Print() string {
	return s.Value.String()
}

// Static is an origin that returns the same value for all queries.
type Static struct{}

// NewStatic creates a new static origin.
func NewStatic() *Static {
	return &Static{}
}

// FetchDataPoints implements the data.Type interface.
func (s *Static) FetchDataPoints(_ context.Context, query []any) (map[any]data.Point, error) {
	points := make(map[any]data.Point, len(query))
	for _, q := range query {
		switch n := q.(type) {
		case *big.Int, *big.Float, *bn.FloatNumber, *bn.IntNumber,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			points[q] = data.Point{
				Value: StaticValue{Value: bn.Float(n)},
				Time:  time.Now(),
			}
		default:
			return nil, fmt.Errorf("invalid query type: %T", q)
		}

	}
	return points, nil
}
