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
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
)

// Handler is interface that all Origin API handlers should implement.
type Handler interface {
	// Fetch should implement making API request to origin URL and
	// collecting/parsing origin data.
	Fetch(pairs []Pair) []FetchResult
}

type ExchangeHandler interface {
	// PullPrices is similar to Handler.Fetch
	// but pairs will be already renamed based on given BaseExchangeHandler.symbolAliases
	PullPrices(pairs []Pair) []FetchResult
}

type BaseExchangeHandler struct {
	ExchangeHandler
	aliases SymbolAliases
}

func NewBaseExchangeHandler(handler ExchangeHandler, aliases SymbolAliases) *BaseExchangeHandler {
	return &BaseExchangeHandler{
		ExchangeHandler: handler,
		aliases:         aliases,
	}
}

func (h BaseExchangeHandler) Fetch(pairs []Pair) []FetchResult {
	if h.aliases == nil {
		return h.PullPrices(pairs)
	}

	var renamedPairs []Pair
	for _, pair := range pairs {
		renamedPairs = append(renamedPairs, h.aliases.replacePair(pair))
	}
	results := h.PullPrices(renamedPairs)

	// Reverting our replacement
	for i := range results {
		results[i].Price.Pair = h.aliases.revertPair(results[i].Price.Pair)
	}
	return results
}

type ContractAddresses map[string]string

func (c ContractAddresses) ByPair(p Pair) (string, bool, bool) {
	contract, ok := c[fmt.Sprintf("%s/%s", p.Base, p.Quote)]
	if !ok {
		contract, ok = c[fmt.Sprintf("%s/%s", p.Quote, p.Base)]
		return contract, true, ok
	}
	return contract, false, ok
}

func (c ContractAddresses) AddressByPair(pair Pair) (ethereum.Address, bool, error) {
	contract, inverted, ok := c.ByPair(pair)
	if !ok {
		return ethereum.Address{}, inverted, fmt.Errorf("failed to get contract address for pair: %s", pair.String())
	}
	return ethereum.HexToAddress(contract), inverted, nil
}

type SymbolAliases map[string]string

func (a SymbolAliases) replaceSymbol(symbol string) (string, bool) {
	replacement, ok := a[symbol]
	if !ok {
		return symbol, false
	}
	return replacement, ok
}

// revertSymbol reverts symbol replacement.
func (a SymbolAliases) revertSymbol(symbol string) string {
	for pre, post := range a {
		if symbol == post {
			return pre
		}
	}

	return symbol
}

func (a SymbolAliases) replacePair(pair Pair) Pair {
	base, baseOk := a.replaceSymbol(pair.Base)
	quote, quoteOk := a.replaceSymbol(pair.Quote)

	return Pair{Base: base, Quote: quote, baseReplaced: baseOk, quoteReplaced: quoteOk}
}

func (a SymbolAliases) revertPair(pair Pair) Pair {
	base := pair.Base
	if pair.baseReplaced {
		base = a.revertSymbol(pair.Base)
	}

	quote := pair.Quote
	if pair.quoteReplaced {
		quote = a.revertSymbol(pair.Quote)
	}

	return Pair{Base: base, Quote: quote, baseReplaced: false, quoteReplaced: false}
}

type Pair struct {
	Base          string
	Quote         string
	baseReplaced  bool
	quoteReplaced bool
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}
func (p Pair) Inverse() Pair {
	return Pair{
		Base:          p.Quote,
		Quote:         p.Base,
		baseReplaced:  p.quoteReplaced,
		quoteReplaced: p.baseReplaced,
	}
}

func (p Pair) Equal(c Pair) bool {
	return p.Base == c.Base && p.Quote == c.Quote
}

type Price struct {
	Pair      Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Timestamp time.Time
}

type FetchResult struct {
	Price Price
	Error error
}

func fetchResult(price Price) FetchResult {
	return FetchResult{
		Price: price,
		Error: nil,
	}
}

func fetchResultWithError(pair Pair, err error) FetchResult {
	return FetchResult{
		Price: Price{
			Pair:      pair,
			Timestamp: time.Now(),
		},
		Error: err,
	}
}

func fetchResultListWithErrors(pairs []Pair, err error) []FetchResult {
	r := make([]FetchResult, len(pairs))
	for i, pair := range pairs {
		r[i] = FetchResult{
			Price: Price{
				Pair:      pair,
				Timestamp: time.Now(),
			},
			Error: err,
		}
	}
	return r
}

type Set struct {
	list map[string]Handler
}

func NewSet(list map[string]Handler) *Set {
	return &Set{list: list}
}

func (e *Set) SetHandler(name string, handler Handler) {
	e.list[name] = handler
}

func (e *Set) Handlers() map[string]Handler {
	c := map[string]Handler{}
	for k, v := range e.list {
		c[k] = v
	}
	return c
}

// Fetch makes handler fetch using handlers from the Set structure.
func (e *Set) Fetch(originPairs map[string][]Pair) map[string][]FetchResult {
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(originPairs))

	frs := map[string][]FetchResult{}
	for origin, pairs := range originPairs {
		origin, pairs := origin, pairs
		handler, ok := e.list[origin]

		go func() {
			if !ok {
				mu.Lock()
				frs[origin] = fetchResultListWithErrors(
					pairs,
					fmt.Errorf("%w (%s)", ErrUnknownOrigin, origin),
				)
				mu.Unlock()
			} else {
				resp := handler.Fetch(pairs)
				mu.Lock()
				frs[origin] = append(frs[origin], resp...)
				mu.Unlock()
			}

			wg.Done()
		}()
	}

	wg.Wait()
	return frs
}

func DefaultOriginSet(pool query.WorkerPool) *Set {
	return NewSet(map[string]Handler{
		"binance":       NewBaseExchangeHandler(Binance{WorkerPool: pool}, nil),
		"bitfinex":      NewBaseExchangeHandler(Bitfinex{WorkerPool: pool}, nil),
		"bitstamp":      NewBaseExchangeHandler(Bitstamp{WorkerPool: pool}, nil),
		"bitthumb":      NewBaseExchangeHandler(BitThump{WorkerPool: pool}, nil),
		"bithumb":       NewBaseExchangeHandler(BitThump{WorkerPool: pool}, nil),
		"coinbase":      NewBaseExchangeHandler(CoinbasePro{WorkerPool: pool}, nil),
		"coinbasepro":   NewBaseExchangeHandler(CoinbasePro{WorkerPool: pool}, nil),
		"cryptocompare": NewBaseExchangeHandler(CryptoCompare{WorkerPool: pool}, nil),
		"ddex":          NewBaseExchangeHandler(Ddex{WorkerPool: pool}, nil),
		"folgory":       NewBaseExchangeHandler(Folgory{WorkerPool: pool}, nil),
		"ftx":           NewBaseExchangeHandler(Ftx{WorkerPool: pool}, nil),
		"gateio":        NewBaseExchangeHandler(Gateio{WorkerPool: pool}, nil),
		"gemini":        NewBaseExchangeHandler(Gemini{WorkerPool: pool}, nil),
		"hitbtc":        NewBaseExchangeHandler(Hitbtc{WorkerPool: pool}, nil),
		"huobi":         NewBaseExchangeHandler(Huobi{WorkerPool: pool}, nil),
		"kraken":        NewBaseExchangeHandler(Kraken{WorkerPool: pool}, nil),
		"kucoin":        NewBaseExchangeHandler(Kucoin{WorkerPool: pool}, nil),
		"loopring":      NewBaseExchangeHandler(Loopring{WorkerPool: pool}, nil),
		"okex":          NewBaseExchangeHandler(Okex{WorkerPool: pool}, nil),
		"okx":           NewBaseExchangeHandler(Okx{WorkerPool: pool}, nil),
		"upbit":         NewBaseExchangeHandler(Upbit{WorkerPool: pool}, nil),
	})
}

type singlePairOrigin interface {
	callOne(pair Pair) (*Price, error)
}

func callSinglePairOrigin(e singlePairOrigin, pairs []Pair) []FetchResult {
	crs := make([]FetchResult, 0)
	for _, pair := range pairs {
		price, err := e.callOne(pair)
		if err != nil {
			crs = append(crs, FetchResult{
				Price: Price{Pair: pair},
				Error: err,
			})
		} else {
			crs = append(crs, FetchResult{
				Price: *price,
				Error: err,
			})
		}
	}

	return crs
}

func validateResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrInvalidResponseStatus)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("bad response: %w", res.Error))
	}
	return nil
}

func buildOriginURL(template, configURL, defaultURL string, a ...interface{}) string {
	url := configURL
	if url == "" {
		url = defaultURL
	}

	replacement := []interface{}{url}
	replacement = append(replacement, a...)

	return fmt.Sprintf(template, replacement...)
}

func reduceEtherAverageFloat(r [][]byte) *big.Float {
	total := new(big.Float).SetInt64(0)
	for _, resp := range r {
		// TODO(jamesr) Always uint256, so even if resp is larger, truncate.
		// However, this assumes that we only care about the first 32 bytes.
		// You might want the last 32... perhaps revisit this.
		price := new(big.Int).SetBytes(resp[0:32])
		total = new(big.Float).Add(
			total,
			new(big.Float).Quo(new(big.Float).SetInt(price), new(big.Float).SetUint64(ether)),
		)
	}
	return new(big.Float).Quo(total, new(big.Float).SetUint64(uint64(len(r))))
}
