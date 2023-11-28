//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

package origin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
)

const DeGateLoggerTag = "DEGATE_ORIGIN"

type assetPairs []AssetPair

func (m assetPairs) byPair(p value.Pair) int {
	for index, pair := range m {
		baseIndex := pair.IndexOf(p.Base)
		quoteIndex := pair.IndexOf(p.Quote)
		if baseIndex >= 0 && quoteIndex >= 0 && baseIndex < quoteIndex {
			return index
		}
	}
	return -1
}

// Structure of DeGate response
// Copied from DeGate Golang sdk
// https://github.com/degatedev/degate-sdk-golang/blob/master/degate/binance/response.go#L3
type degateBaseResponse struct {
	Code int `json:"code"`
}

// Copied from DeGate Golang SDK
// https://github.com/degatedev/degate-sdk-golang/blob/master/degate/model/model.go#L5
type degateTokenInfo struct {
	ID int `json:"id"`
	//Chain           string `json:"chain"`
	//Code            string `json:"code"`
	Symbol string `json:"symbol"`
	//Decimals        int32  `json:"decimals"`
	//IsTrustedToken  bool   `json:"is_trusted_token"`
	//IsQuotableToken bool   `json:"is_quotable_token"`
	//IsGasToken      bool   `json:"is_gas_token"`
	//IsListToken     bool   `json:"is_list_token"`
	//Active          bool   `json:"active"`
	//ShowDecimal     int32  `json:"show_decimal"`
	//Priority        uint64 `json:"priority"`
}

type degateTokensResponse struct {
	degateBaseResponse
	Data []*degateTokenInfo `json:"data"`
}

// Copied from DeGate Golang SDk
// https://github.com/degatedev/degate-sdk-golang/blob/master/degate/binance/market.go#L39
type degateTicker struct {
	//BaseTokenID        uint64 `json:"base_token_id"`
	//QuoteTokenID       uint64 `json:"quote_token_id"`
	//PriceChange        string `json:"price_change"`
	//PriceChangePercent string `json:"price_change_percent"`
	//WeightedAvgPrice   string `json:"weighted_avg_price"`
	//PrevClosePrice     string `json:"prev_close_price"`
	LastPrice string `json:"last_price"`
	//LastQty            string `json:"last_qty"`
	//BidPrice           string `json:"bid_price"`
	//BidQty             string `json:"bid_qty"`
	//AskPrice           string `json:"ask_price"`
	//AskQty             string `json:"ask_qty"`
	//OpenPrice          string `json:"open_price"`
	//HighPrice          string `json:"high_price"`
	//LowPrice           string `json:"low_price"`
	Volume string `json:"volume"`
	//QuoteVolume     string `json:"quote_volume"`
	//OpenTime        int64  `json:"open_time"`
	//CloseTime       int64  `json:"close_time"`
	//FirstID         string `json:"first_id"`
	//LastID          string `json:"last_id"`
	//Count           uint64 `json:"count"`
	//WeekHighPrice   string `json:"week_high_price"`
	//WeekLowPrice    string `json:"week_low_price"`
	//BaseTokenPrice  string `json:"base_token_price"`
	//QuoteTokenPrice string `json:"quote_token_price"`
	//MakerFee        string `json:"maker_fee"`
	//TakerFee        string `json:"taker_fee"`
	//PairID          uint64 `json:"pair_id"`
}

type degateTickerResponse struct {
	degateBaseResponse
	Data *degateTicker `json:"data"`
}

type DeGateConfig struct {
	Endpoint string
	Client   *http.Client

	// Available pairs
	Pairs  assetPairs
	Logger log.Logger
}

type DeGate struct {
	endpoint   string
	client     *http.Client
	pairs      assetPairs
	tokenIdMap map[string]int
	logger     log.Logger
}

func NewDeGate(config DeGateConfig) (*DeGate, error) {
	if config.Logger == nil {
		config.Logger = null.New()
	}
	if config.Client == nil {
		config.Client = &http.Client{}
	}

	return &DeGate{
		endpoint:   config.Endpoint,
		client:     config.Client,
		pairs:      config.Pairs,
		tokenIdMap: make(map[string]int),
		logger:     config.Logger.WithField("degate", DeGateLoggerTag),
	}, nil
}

func (d *DeGate) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	// Fetch token id if not found in cache
	var tokens []string
	for _, pair := range pairs {
		if _, ok := d.tokenIdMap[pair.Base]; !ok {
			tokens = append(tokens, pair.Base)
		}
		if _, ok := d.tokenIdMap[pair.Quote]; !ok {
			tokens = append(tokens, pair.Quote)
		}
	}
	if len(tokens) > 0 {
		if err := d.fetchTokenIds(ctx, tokens); err != nil {
			return nil, err
		}
	}

	for _, pair := range pairs {
		index := d.pairs.byPair(pair)
		if index < 0 {
			points[pair] = datapoint.Point{Error: fmt.Errorf("unsupported pair: %s", pair.String())}
			continue
		}

		ticker, err := d.fetchTicker24(ctx, d.tokenIdMap[pair.Base], d.tokenIdMap[pair.Quote])
		if err != nil || ticker == nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf("failed in fetching ticker24(%s): %v", pair.String(), err)}
			continue
		}

		points[pair] = datapoint.Point{
			Value: value.NewTick(pair, ticker.LastPrice, ticker.Volume),
			Time:  time.Now(),
		}
	}

	return points, nil
}

// fetchTokenIds fetches the token ids in the cache and return error
// Reference: https://github.com/degatedev/degate-sdk-golang/blob/master/degate/spot/exchange_client.go#L164
func (d *DeGate) fetchTokenIds(ctx context.Context, tokens []string) error {
	symbols := strings.Join(tokens, ",")

	url := d.endpoint + "/order-book-api/exchange/tokenList"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	query := request.URL.Query()
	query.Add("symbols", symbols)
	request.URL.RawQuery = query.Encode()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("use-trade-key", "0")

	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var body bytes.Buffer
	_, err = io.Copy(&body, response.Body)
	if err != nil {
		return err
	}

	tokensResponse := &degateTokensResponse{}
	err = json.Unmarshal(body.Bytes(), tokensResponse)
	if err != nil {
		return err
	}

	for _, token := range tokensResponse.Data {
		d.tokenIdMap[strings.ToUpper(token.Symbol)] = token.ID
	}
	return nil
}

// fetchTicker24 fetches the 24hr ticker price change statistics
// Reference: https://github.com/degatedev/degate-sdk-golang/blob/master/degate/spot/market_client.go#L181
func (d *DeGate) fetchTicker24(ctx context.Context, baseTokenID, quoteTokenID int) (*degateTicker, error) {
	url := d.endpoint + "/order-book-ws-api/ticker"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	query := request.URL.Query()
	query.Add("base_token_id", strconv.Itoa(baseTokenID))
	query.Add("quote_token_id", strconv.Itoa(quoteTokenID))
	request.URL.RawQuery = query.Encode()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("use-trade-key", "0")

	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var body bytes.Buffer
	_, err = io.Copy(&body, response.Body)
	if err != nil {
		return nil, err
	}

	tickerResponse := &degateTickerResponse{}
	err = json.Unmarshal(body.Bytes(), tickerResponse)
	if err != nil {
		return nil, err
	}
	return tickerResponse.Data, nil
}
