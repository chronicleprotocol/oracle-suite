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

package provider

import (
	"fmt"
	"strings"
	"time"
)

// Provider provides prices for asset pairs.
type Provider interface {
	// Models describes price models which are used to calculate prices.
	// If no pairs are specified, models for all pairs are returned.
	Models(pairs ...Pair) (map[Pair]*Model, error)

	// Price returns a Price for the given pair.
	Price(pair Pair) (*Price, error)

	// Prices returns prices for the given pairs. If no pairs are specified,
	// prices for all pairs are returned.
	Prices(pairs ...Pair) (map[Pair]*Price, error)

	// Pairs returns all pairs.
	Pairs() ([]Pair, error)
}

// Pair represents an asset pair.
type Pair struct {
	Base  string
	Quote string
}

// NewPair returns a new Pair for given string. The string must be formatted
// as "BASE/QUOTE".
func NewPair(s string) (Pair, error) {
	var p Pair
	err := p.UnmarshalText([]byte(s))
	return p, err
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (p *Pair) UnmarshalText(text []byte) error {
	ss := strings.Split(string(text), "/")
	if len(ss) != 2 {
		return fmt.Errorf("invalid pair format: %s, expected BASE/QUOTE", text)
	}
	p.Base = strings.ToUpper(ss[0])
	p.Quote = strings.ToUpper(ss[1])
	return nil
}

// NewPairs returns a Pair slice for given strings. Pair strings must be
// formatted as "BASE/QUOTE".
func NewPairs(s ...string) ([]Pair, error) {
	var r []Pair
	for _, p := range s {
		pr, err := NewPair(p)
		if err != nil {
			return nil, err
		}
		r = append(r, pr)
	}
	return r, nil
}

func (p Pair) Empty() bool {
	return p.Base == "" && p.Quote == ""
}

func (p Pair) Equal(c Pair) bool {
	return p.Base == c.Base && p.Quote == c.Quote
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

// Model is a simplified representation of a model which is used to calculate
// asset pair prices. The main purpose of this structure is to help the end
// user to understand how prices are derived and calculated.
//
// This structure is purely informational. The way it is used depends on
// a specific implementation.
type Model struct {
	// Type is used to differentiate between model types.
	Type string
	// Parameters is a optional list of model's parameters.
	Parameters map[string]string
	// Pair is a asset pair for which this model returns a price.
	Pair Pair
	// Models is a list of sub models used to calculate price.
	Models []*Model
}

// Price represents price for a single pair. If the Price was calculated
// indirectly it will also contain all prices used to calculate the price.
type Price struct {
	Type       string
	Parameters map[string]string
	Pair       Pair
	Price      float64
	Bid        float64
	Ask        float64
	Volume24h  float64
	Time       time.Time
	Prices     []*Price
	Error      string
}

type PriceHook interface {
	Check(map[Pair]*Price) error
}
