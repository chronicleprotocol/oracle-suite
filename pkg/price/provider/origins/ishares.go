//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package origins

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/webscraper"
)

const iSharesBaseURL = "https://ishares.com"
const ibtaURL = "%s/uk/individual/en/products/287340/ishares-treasury-bond-1-3yr-ucits-etf?switchLocale=y&siteEntryPassthrough=true" //nolint:lll

// Origin handler
type IShares struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

func (o IShares) localPairName(pair Pair) string {
	return pair.Base + pair.Quote
}

func (o IShares) Pool() query.WorkerPool {
	return o.WorkerPool
}

func (o IShares) PullPrices(pairs []Pair) []FetchResult {
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		switch o.localPairName(pair) {
		case "IBTAUSD":
			// Do query via pool
			req := &query.HTTPRequest{
				URL: buildOriginURL(ibtaURL, o.BaseURL, iSharesBaseURL),
			}
			res := o.WorkerPool.Query(req)
			if res == nil {
				return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
			}
			if res.Error != nil {
				return fetchResultListWithErrors(pairs, res.Error)
			}
			// Scrape results
			w, err := webscraper.NewScraper().WithPreloadedDocFromBytes(res.Body)
			if err != nil {
				return fetchResultListWithErrors(pairs, err)
			}
			var convErrs []string
			err = w.Scrape("span.header-nav-data",
				func(e webscraper.Element) {
					txt := strings.ReplaceAll(e.Text, "\n", "")
					if strings.HasPrefix(txt, "USD ") {
						ntxt := strings.ReplaceAll(txt, "USD ", "")
						if price, e := strconv.ParseFloat(ntxt, 64); e == nil {
							results = append(results, FetchResult{
								Price: Price{
									Pair:      pair,
									Price:     price,
									Timestamp: time.Now(),
								},
							})
						} else {
							convErrs = append(convErrs, e.Error())
						}
					}
				})
			if err != nil {
				return fetchResultListWithErrors(pairs, err)
			}
			if len(convErrs) > 0 {
				return fetchResultListWithErrors(pairs, errors.New(strings.Join(convErrs, ",")))
			}
		default:
			return fetchResultListWithErrors(pairs, ErrUnknownPair)
		}
	}
	return results
}
