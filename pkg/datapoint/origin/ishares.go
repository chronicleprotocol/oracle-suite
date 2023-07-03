package origin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/webscraper"
)

const ISharesLoggerTag = "ISHARES_ORIGIN"

type ISharesOptions struct {
	URL     string
	Headers http.Header
	Client  *http.Client
	Logger  log.Logger
}

type IShares struct {
	http   *TickGenericHTTP
	logger log.Logger
}

func NewIShares(opts ISharesOptions) (*IShares, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	if opts.Logger == nil {
		opts.Logger = null.New()
	}

	ishares := &IShares{}
	gh, err := NewTickGenericHTTP(TickGenericHTTPOptions{
		URL:      opts.URL,
		Headers:  opts.Headers,
		Callback: ishares.handle,
		Client:   opts.Client,
		Logger:   opts.Logger,
	})
	if err != nil {
		return nil, err
	}
	ishares.http = gh
	ishares.logger = opts.Logger.WithField("ishares", ISharesLoggerTag)
	return ishares, nil
}

// FetchDataPoints implements the Origin interface.
func (g *IShares) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	return g.http.FetchDataPoints(ctx, query)
}

func (g *IShares) handle(_ context.Context, pairs []value.Pair, body io.Reader) map[any]datapoint.Point {
	b, err := io.ReadAll(body)
	if err != nil {
		return fillDataPointsWithError(pairs, err)
	}

	points := make(map[any]datapoint.Point)
	for _, pair := range pairs {
		if pair.String() != "IBTA/USD" {
			points[pair] = datapoint.Point{Error: fmt.Errorf("unknown pair: %s", pair.String())}
			continue
		}

		// Scrape results
		w, err := webscraper.NewScraper().WithPreloadedDocFromBytes(b)
		if err != nil {
			return fillDataPointsWithError(pairs, err)
		}
		var convErrs []string
		err = w.Scrape("span.header-nav-data",
			func(e webscraper.Element) {
				txt := strings.ReplaceAll(e.Text, "\n", "")
				if strings.HasPrefix(txt, "USD ") {
					ntxt := strings.ReplaceAll(txt, "USD ", "")

					if price, e := strconv.ParseFloat(ntxt, 64); e == nil {
						tick := value.Tick{Pair: pair, Price: bn.Float(price)}
						points[pair] = datapoint.Point{
							Value: tick,
							Time:  time.Now(),
						}
					} else {
						convErrs = append(convErrs, e.Error())
					}
				}
			})
		if err != nil {
			return fillDataPointsWithError(pairs, err)
		}
		if len(convErrs) > 0 {
			err := errors.New(strings.Join(convErrs, ","))
			return fillDataPointsWithError(pairs, err)
		}
	}

	return points
}
