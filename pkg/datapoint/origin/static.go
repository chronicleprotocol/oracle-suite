package origin

import (
	"context"
	"fmt"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Static is an origin that returns the same value for all queries.
type Static struct{}

// NewStatic creates a new static origin.
func NewStatic() *Static {
	return &Static{}
}

// FetchDataPoints implements the data.Type interface.
func (s *Static) FetchDataPoints(_ context.Context, query []any) (map[any]datapoint.Point, error) {
	points := make(map[any]datapoint.Point, len(query))
	for _, q := range query {
		f := bn.Float(q)
		if f == nil {
			return nil, fmt.Errorf("invalid query: %T", q)
		}
		points[q] = datapoint.Point{
			Value: value.StaticValue{Value: f},
			Time:  time.Now(),
		}
	}
	return points, nil
}
