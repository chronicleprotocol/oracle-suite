package origin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/interpolate"
)

const TickGenericHTTPLoggerTag = "TICK_GENERIC_HTTP_ORIGIN"

type HTTPCallback func(ctx context.Context, pairs []value.Pair, data io.Reader) map[any]datapoint.Point

type TickGenericHTTPOptions struct {
	// URL is an TickGenericHTTP endpoint that returns JSON data. It may contain
	// the following variables:
	//   - ${lcbase} - lower case base asset
	//   - ${ucbase} - upper case base asset
	//   - ${lcquote} - lower case quote asset
	//   - ${ucquote} - upper case quote asset
	//   - ${lcbases} - lower case base assets joined by commas
	//   - ${ucbases} - upper case base assets joined by commas
	//   - ${lcquotes} - lower case quote assets joined by commas
	//   - ${ucquotes} - upper case quote assets joined by commas
	URL string

	// Headers is a set of TickGenericHTTP headers that are sent with each request.
	Headers http.Header

	// Callback is a function that is used to parse the response body.
	Callback HTTPCallback

	// Client is an TickGenericHTTP client that is used to fetch data from the
	// TickGenericHTTP endpoint. If nil, http.DefaultClient is used.
	Client *http.Client

	// Logger is an TickGenericHTTP logger that is used to log errors. If nil,
	// null logger is used.
	Logger log.Logger
}

// TickGenericHTTP is a generic http price provider that can fetch prices from
// an HTTP endpoint. The callback function is used to parse the response body.
type TickGenericHTTP struct {
	url      string
	client   *http.Client
	headers  http.Header
	callback HTTPCallback
	logger   log.Logger
}

// NewTickGenericHTTP creates a new TickGenericHTTP instance.
func NewTickGenericHTTP(opts TickGenericHTTPOptions) (*TickGenericHTTP, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if opts.Callback == nil {
		return nil, fmt.Errorf("callback cannot be nil")
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	if opts.Logger == nil {
		opts.Logger = null.New()
	}
	return &TickGenericHTTP{
		url:      opts.URL,
		client:   opts.Client,
		headers:  opts.Headers,
		callback: opts.Callback,
		logger:   opts.Logger.WithField("tag", TickGenericHTTPLoggerTag),
	}, nil
}

// FetchDataPoints implements the Origin interface.
func (g *TickGenericHTTP) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}
	points := make(map[any]datapoint.Point)
	for url, pairs := range g.group(pairs) {
		g.logger.
			WithFields(log.Fields{
				"url":   url,
				"pairs": pairs,
			}).
			Debug("HTTP request")

		// Perform TickGenericHTTP request.
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			for _, pair := range pairs {
				points[pair] = datapoint.Point{Error: err}
			}
			continue
		}
		req.Header = g.headers
		req = req.WithContext(ctx)

		// Execute TickGenericHTTP request.
		res, err := g.client.Do(req)
		if err != nil {
			for _, pair := range pairs {
				points[pair] = datapoint.Point{Error: err}
			}
			continue
		}
		defer res.Body.Close()

		// Run callback function.
		for pair, point := range g.callback(ctx, pairs, res.Body) {
			points[pair] = point
		}
	}
	return points, nil
}

// group interpolates the URL by substituting the base and quote, and then
// groups the resulting pairs by the interpolated URL.
func (g *TickGenericHTTP) group(pairs []value.Pair) map[string][]value.Pair {
	pairMap := make(map[string][]value.Pair)
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

func queryToPairs(query []any) ([]value.Pair, bool) {
	pairs := make([]value.Pair, len(query))
	for i, q := range query {
		switch q := q.(type) {
		case value.Pair:
			pairs[i] = q
		default:
			return nil, false
		}
	}
	return pairs, true
}
