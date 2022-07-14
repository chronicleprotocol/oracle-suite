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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

type Bitfinex struct {
	WorkerPool query.WorkerPool
	BaseURL    string
}

const bitfinexBaseURL = "https://api-pub.bitfinex.com"
const bitfinexURL = "%s/v2/tickers?symbols=%s"

func (b Bitfinex) Pool() query.WorkerPool {
	return b.WorkerPool
}

func (b Bitfinex) PullPrices(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: buildOriginURL(bitfinexURL, b.BaseURL, bitfinexBaseURL, b.localPairName(pairs...)),
	}
	res := b.WorkerPool.Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return b.parseResponse(pairs, res)
}

func (b *Bitfinex) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	var resp [][]interface{}
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}
	return b.mapResults(pairs, b.mapTickers(resp))
}

func (b *Bitfinex) mapResults(pairs []Pair, tickers map[string]bitfinexTicker) []FetchResult {
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		//nolint:gocritic
		if t, is := tickers[b.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else if t.Error != nil {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: t.Error,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.LastPrice,
					Ask:       t.Ask,
					Bid:       t.Bid,
					Volume24h: t.Volume,
					Timestamp: time.Now(),
				},
			})
		}
	}
	return results
}

func (b *Bitfinex) mapTickers(resp [][]interface{}) map[string]bitfinexTicker {
	tickers := make(map[string]bitfinexTicker)
	for _, tt := range resp {
		t := b.parseTicker(tt)
		tickers[t.Symbol] = t
	}
	return tickers
}

type bitfinexTicker struct {
	Symbol              string  //  [0]: SYMBOL
	Bid                 float64 //  [1]: BID
	BidSize             float64 //  [2]: BID_SIZE
	Ask                 float64 //  [3]: ASK
	AskSize             float64 //  [4]: ASK_SIZE
	DailyChange         float64 //  [5]: DAILY_CHANGE
	DailyChangeRelative float64 //  [6]: DAILY_CHANGE_RELATIVE
	LastPrice           float64 //  [7]: LAST_PRICE
	Volume              float64 //  [8]: VOLUME
	High                float64 //  [9]: HIGH
	Low                 float64 // [10]: LOW
	Error               error
}

//nolint:funlen,gocyclo
func (*Bitfinex) parseTicker(tt []interface{}) bitfinexTicker {
	var t bitfinexTicker
	crc := make(map[int]bool)
	for i, a := range tt {
		switch x := a.(type) {
		case string:
			t.Symbol = x
			if i != 0 {
				t.Error = errors.New("market symbol is not at index 0")
				return t
			}
			crc[i] = true
		case float64:
			switch i {
			case 1:
				t.Bid = x
			case 2:
				t.BidSize = x
			case 3:
				t.Ask = x
			case 4:
				t.AskSize = x
			case 5:
				t.DailyChange = x
			case 6:
				t.DailyChangeRelative = x
			case 7:
				t.LastPrice = x
			case 8:
				t.Volume = x
			case 9:
				t.High = x
			case 10:
				t.Low = x
			}
			crc[i] = true
		default:
			t.Error = fmt.Errorf("item at index %d is unexpexted (Type: %T)", i, x)
			return t
		}
	}
	expectedItemCount := 11
	if len(crc) > expectedItemCount {
		t.Error = fmt.Errorf("too many (%d) items", len(crc)+1)
		return t
	}
	for i := 0; i <= 10; i++ {
		if v, ok := crc[i]; !ok || !v {
			t.Error = fmt.Errorf("item at index %d is missing", i)
			return t
		}
	}
	return t
}

// TODO: move to aliases ?
//nolint:lll
const bitfinexConfig = `[[["AAA","TESTAAA"],["ABS","ABYSS"],["AIO","AION"],["ALG","ALGO"],["AMP","AMPL"],["AMPF0","AMPLF0"],["ATO","ATOM"],["BAB","BCH"],["BBB","TESTBBB"],["CNHT","CNHt"],["CSX","CS"],["CTX","CTXC"],["DAT","DATA"],["DOG","MDOGE"],["DRN","DRGN"],["DSH","DASH"],["DTX","DT"],["EDO","PNT"],["EUS","EURS"],["EUT","EURt"],["GSD","GUSD"],["IOS","IOST"],["IOT","IOTA"],["LBT","LBTC"],["LES","LEO-EOS"],["LET","LEO-ERC20"],["MIT","MITH"],["MNA","MANA"],["NCA","NCASH"],["OMN","OMNI"],["PAS","PASS"],["POY","POLY"],["QSH","QASH"],["QTM","QTUM"],["RBT","RBTC"],["REP","REP2"],["SCR","XD"],["SNG","SNGLS"],["SPK","SPANK"],["STJ","STORJ"],["TSD","TUSD"],["UDC","USDC"],["USK","USDK"],["UST","USDt"],["USTF0","USDt0"],["UTN","UTNP"],["VSY","VSYS"],["WBT","WBTC"],["XAUT","XAUt"],["XCH","XCHF"],["YGG","YEED"],["YYW","YOYOW"]]]`

func (*Bitfinex) localPairName(pairs ...Pair) string {
	c := bitfinexConfigMap([]byte(bitfinexConfig))
	var l []string
	for _, pair := range pairs {
		b, ok := c[pair.Base]
		if !ok {
			b = pair.Base
		}
		q, ok := c[pair.Quote]
		if !ok {
			q = pair.Quote
		}
		// Hack to treat prices with USD quotes from Bitfinex as if they were with USDT quotes
		if pair.Quote == "USDT" {
			q = "USD"
		}
		l = append(l, fmt.Sprintf("t%s%s", b, q))
	}
	return strings.Join(l, ",")
}

func bitfinexConfigMap(j []byte) map[string]string {
	var a [][][]string
	err := json.Unmarshal(j, &a)
	if err != nil {
		panic(err)
	}

	c := make(map[string]string)
	for _, v := range a[0] {
		c[strings.ToUpper(v[1])] = strings.ToUpper(v[0])
	}
	return c
}
