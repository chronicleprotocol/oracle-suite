package origin

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/interpolate"
)

type HTTPCallback func(ctx context.Context, pairs []provider.Pair, data io.Reader) []provider.Tick

// GenericHTTP is a generic GenericHTTP price provider that can fetch prices from
// an GenericHTTP endpoint. The callback function is used to parse the response body.
type GenericHTTP struct {
	// url is an GenericHTTP endpoint that returns JSON data.
	url string

	// client is an GenericHTTP client that is used to fetch data from the GenericHTTP endpoint.
	client *http.Client

	// headers is a set of GenericHTTP headers that are sent with each request.
	headers http.Header

	// callback is a function that is used to parse the response body.
	callback HTTPCallback
}

// NewGenericHTTP creates a new GenericHTTP instance.
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
func NewGenericHTTP(client *http.Client, header http.Header, url string, cb HTTPCallback) (*GenericHTTP, error) {
	if client == nil {
		client = http.DefaultClient
	}
	return &GenericHTTP{
		url:      url,
		client:   client,
		headers:  header,
		callback: cb,
	}, nil
}

// FetchTicks implements the Origin interface.
func (g *GenericHTTP) FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick {
	var ticks []provider.Tick
	for url, pairs := range g.group(pairs) {
		// Perform GenericHTTP request.
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			ticks = append(ticks, withError(pairs, err)...)
			continue
		}
		req.Header = g.headers
		req = req.WithContext(ctx)

		// Execute GenericHTTP request.
		res, err := g.client.Do(req)
		if err != nil {
			ticks = append(ticks, withError(pairs, err)...)
			continue
		}
		defer res.Body.Close()

		// Run callback function.
		ticks = append(ticks, g.callback(ctx, pairs, res.Body)...)
	}
	return ticks
}

// group interpolates the URL by substituting the base and quote, and then
// groups the resulting pairs by the interpolated URL.
func (g *GenericHTTP) group(pairs []provider.Pair) map[string][]provider.Pair {
	pairMap := make(map[string][]provider.Pair)
	parsedURL := interpolate.Parse(g.url)
	bases := make([]string, 0, len(pairs))
	quotes := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		bases = append(bases, pair.Base)
		quotes = append(quotes, pair.Quote)
	}
	for _, pair := range pairs {
		url := parsedURL.Interpolate(func(variable interpolate.Variable) string {
			switch variable.Name {
			case "lcbase":
				return strings.ToLower(pair.Base)
			case "ucbase":
				return strings.ToUpper(pair.Base)
			case "lcquote":
				return strings.ToLower(pair.Quote)
			case "ucquote":
				return strings.ToUpper(pair.Quote)
			case "lcbases":
				return strings.ToLower(strings.Join(bases, ","))
			case "ucbases":
				return strings.ToUpper(strings.Join(bases, ","))
			case "lcquotes":
				return strings.ToLower(strings.Join(quotes, ","))
			case "ucquotes":
				return strings.ToUpper(strings.Join(quotes, ","))
			default:
				return variable.Default
			}
		})
		pairMap[url] = append(pairMap[url], pair)
	}
	return pairMap
}
