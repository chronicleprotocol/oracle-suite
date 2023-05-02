package origin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/pricenext/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/interpolate"
)

const GenericHTTPLoggerTag = "GENERIC_HTTP_ORIGIN"

type HTTPCallback func(ctx context.Context, pairs []provider.Pair, data io.Reader) []provider.Tick

type GenericHTTPOptions struct {
	// URL is an GenericHTTP endpoint that returns JSON data. It may contain
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

	// Headers is a set of GenericHTTP headers that are sent with each request.
	Headers http.Header

	// Callback is a function that is used to parse the response body.
	Callback HTTPCallback

	// Client is an GenericHTTP client that is used to fetch data from the
	// GenericHTTP endpoint. If nil, http.DefaultClient is used.
	Client *http.Client

	// Logger is an GenericHTTP logger that is used to log errors. If nil,
	// null logger is used.
	Logger log.Logger
}

// GenericHTTP is a generic GenericHTTP price provider that can fetch prices from
// an GenericHTTP endpoint. The callback function is used to parse the response body.
type GenericHTTP struct {
	url      string
	client   *http.Client
	headers  http.Header
	callback HTTPCallback
	logger   log.Logger
}

// NewGenericHTTP creates a new GenericHTTP instance.
func NewGenericHTTP(opts GenericHTTPOptions) (*GenericHTTP, error) {
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
	return &GenericHTTP{
		url:      opts.URL,
		client:   opts.Client,
		headers:  opts.Headers,
		callback: opts.Callback,
		logger:   opts.Logger.WithField("tag", GenericHTTPLoggerTag),
	}, nil
}

// FetchTicks implements the Origin interface.
func (g *GenericHTTP) FetchTicks(ctx context.Context, pairs []provider.Pair) []provider.Tick {
	var ticks []provider.Tick
	for url, pairs := range g.group(pairs) {
		g.logger.
			WithFields(log.Fields{
				"url":   url,
				"pairs": pairs,
			}).
			Debug("HTTP request")

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
