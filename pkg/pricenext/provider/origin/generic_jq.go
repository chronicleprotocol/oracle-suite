package origin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/itchyny/gojq"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// GenericJQ is a generic origin implementation that uses JQ to parse JSON data
// from an GenericHTTP endpoint.
type GenericJQ struct {
	http *GenericHTTP

	// query is a JQ query that is used to parse JSON data.
	query *gojq.Code
}

// NewGenericJQ creates a new GenericJQ instance.
//
// The client argument is an GenericHTTP client that is used to fetch data from the
// GenericHTTP endpoint.
//
// The url argument is an GenericHTTP endpoint that returns JSON data. It may contain
// the following variables:
//   - ${lcbase} - lower case base asset
//   - ${ucbase} - upper case base asset
//   - ${lcquote} - lower case quote asset
//   - ${ucquote} - upper case quote asset
//   - ${lcbases} - lower case base assets joined by commas
//   - ${ucbases} - upper case base assets joined by commas
//   - ${lcquotes} - lower case quote assets joined by commas
//   - ${ucquotes} - upper case quote assets joined by commas
//
// The jq argument is a JQ query that is used to parse JSON data. It must
// return a single value that will be used as a price or an object with the
// following fields:
//   - price - a price
//   - time - a timestamp (optional)
//   - volume - a 24h volume (optional)
//
// The JQ query may contain the following variables:
//   - $lcbase - lower case base asset
//   - $ucbase - upper case base asset
//   - $lcquote - lower case quote asset
//   - $ucquote - upper case quote asset
//
// Price and volume must be a string that can be parsed as a number or a number.
//
// Time must be a string that can be parsed as time or a number that represents
// a UNIX timestamp.
//
// If JQ query returns multiple values, the tick will be invalid.
func NewGenericJQ(client *http.Client, header http.Header, url string, query string) (*GenericJQ, error) {
	parsed, err := gojq.Parse(query)
	if err != nil {
		return nil, err
	}
	compiled, err := gojq.Compile(parsed, gojq.WithVariables([]string{
		"$lcbase",
		"$ucbase",
		"$lcquote",
		"$ucquote",
	}))
	if err != nil {
		return nil, err
	}
	jq := &GenericJQ{}
	gh, err := NewGenericHTTP(client, header, url, jq.handle)
	if err != nil {
		return nil, err
	}
	jq.http = gh
	jq.query = compiled
	return jq, nil
}

// FetchTicks implements the Origin interface.
func (g *GenericJQ) FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick {
	return g.http.FetchTicks(ctx, pairs)
}

func (g *GenericJQ) handle(ctx context.Context, pairs []provider.Pair, body io.Reader) []provider.Tick {
	var ticks []provider.Tick

	// Parse JSON data.
	var data any
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return withError(pairs, err)
	}

	// Run JQ query for each pair and parse the result.
	for _, pair := range pairs {
		tick := provider.Tick{
			Pair: pair,
			Time: time.Now(),
		}
		iter := g.query.RunWithContext(
			ctx,
			data,
			strings.ToLower(pair.Base),  // $lcbase
			strings.ToUpper(pair.Base),  // $ucbase
			strings.ToLower(pair.Quote), // $lcquote
			strings.ToUpper(pair.Quote), // $ucquote
		)
		v, ok := iter.Next()
		if !ok {
			tick.Error = fmt.Errorf("no result from JQ query")
			ticks = append(ticks, tick)
			continue
		}
		if err, ok := v.(error); ok {
			tick.Error = err
			ticks = append(ticks, tick)
			continue
		}
		if _, ok := iter.Next(); ok {
			tick.Error = fmt.Errorf("multiple results from JQ query")
			ticks = append(ticks, tick)
			continue
		}
		switch v := v.(type) {
		case map[string]any:
			for k, v := range v {
				switch k {
				case "price":
					tick.Price = bn.Float(v)
				case "volume":
					tick.Volume24h = bn.Float(v)
				case "time":
					if tm, ok := anyToTime(v); ok {
						tick.Time = tm
					}
				default:
					tick.Error = fmt.Errorf("unknown key in JQ result: %s", k)
				}
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			tick.Price = bn.Float(v)
		}
		ticks = append(ticks, tick)
	}
	return ticks
}

// anyToTime converts an arbitrary value to a time.Time.
func anyToTime(v any) (time.Time, bool) {
	switch v := v.(type) {
	case string:
		for _, layout := range []string{
			time.RFC3339,
			time.RFC3339Nano,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC822,
			time.RFC822Z,
			time.RFC850,
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
		} {
			t, err := time.Parse(layout, v)
			if err == nil {
				return t, true
			}
		}
	case int, int8, int16, int32, int64:
		return time.Unix(v.(int64), 0), true
	case uint, uint8, uint16, uint32, uint64:
		return time.Unix(int64(v.(uint64)), 0), true
	case float32, float64:
		return time.Unix(int64(v.(float64)), 0), true
	}
	return time.Time{}, false
}
