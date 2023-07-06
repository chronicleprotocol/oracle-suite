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

type ISharesConfig struct {
	URL     string
	Headers http.Header
	Client  *http.Client
	Logger  log.Logger
}

type IShares struct {
	http   *TickGenericHTTP
	logger log.Logger
}

func NewIShares(config ISharesConfig) (*IShares, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	ishares := &IShares{}
	gh, err := NewTickGenericHTTP(TickGenericHTTPOptions{
		URL:      config.URL,
		Headers:  config.Headers,
		Callback: ishares.handle,
		Client:   config.Client,
		Logger:   config.Logger,
	})
	if err != nil {
		return nil, err
	}
	ishares.http = gh
	ishares.logger = config.Logger.WithField("ishares", ISharesLoggerTag)
	return ishares, nil
}

// FetchDataPoints implements the Origin interface.
func (g *IShares) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	return g.http.FetchDataPoints(ctx, query)
}

func (g *IShares) handle(_ context.Context, pairs []value.Pair, body io.Reader) (map[any]datapoint.Point, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http, %w", err)
	}

	points := make(map[any]datapoint.Point)
	for _, pair := range pairs {
		if pair.String() != "IBTA/USD" {
			points[pair] = datapoint.Point{Error: fmt.Errorf("unknown pair: %s", pair.String())}
			continue
		}
	}

	for _, pair := range pairs {
		if pair.String() != "IBTA/USD" {
			continue
		} // IBTA/USD

		// Scrape results
		w, err := webscraper.NewScraper().WithPreloadedDocFromBytes(b)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			return points, err
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
			points[pair] = datapoint.Point{Error: err}
			return points, err
		}
		if len(convErrs) > 0 {
			err := errors.New(strings.Join(convErrs, ","))
			points[pair] = datapoint.Point{Error: err}
			return points, err
		}
	}

	return points, nil
}
