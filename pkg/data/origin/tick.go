package origin

import (
	"fmt"
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Tick contains a price, volume and other information for a given asset pair
// at a given time.
//
// Before using this data, you should check if it is valid by calling
// Tick.Validate() method.
type Tick struct {
	// Pair is an asset pair for which this price is calculated.
	Pair Pair

	// Price is a price for the given asset pair.
	// Depending on the provider implementation, this price can be
	// a last trade price, an average of bid and ask prices, etc.
	//
	// Price is always non-nil if there is no error.
	Price *bn.FloatNumber

	// Volume24h is a 24h volume for the given asset pair presented in the
	// base currency.
	//
	// May be nil if the provider does not provide volume.
	Volume24h *bn.FloatNumber
}

// Print implements the data.Value interface.
func (t Tick) Print() string {
	return fmt.Sprintf("Pair=%s, Price=%s, Volume24h=%s", t.Pair, t.Price, t.Volume24h)
}

// Validate returns an error if the tick is invalid.
func (t Tick) Validate() error {
	if t.Pair.Empty() {
		return fmt.Errorf("pair is not set")
	}
	if t.Price == nil {
		return fmt.Errorf("price is nil")
	}
	if t.Price.Sign() <= 0 {
		return fmt.Errorf("price is zero or negative")
	}
	if t.Price.IsInf() {
		return fmt.Errorf("price is infinite")
	}
	if t.Volume24h != nil && t.Volume24h.Sign() < 0 {
		return fmt.Errorf("volume is negative")
	}
	return nil
}

// Number implements the data.NumericValue interface.
func (t Tick) Number() *bn.FloatNumber {
	return t.Price
}

func (t Tick) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"pair":%q,"price":%q,"volume24h":%q}`, t.Pair, t.Price, t.Volume24h)), nil
}

// Pair represents an asset pair.
type Pair struct {
	Base  string
	Quote string
}

// PairFromString returns a new Pair for given string.
// The string must be formatted as "BASE/QUOTE".
func PairFromString(s string) (p Pair, err error) {
	return p, p.UnmarshalText([]byte(s))
}

// MarshalText implements encoding.TextMarshaler interface.
func (p Pair) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (p *Pair) UnmarshalText(text []byte) error {
	ss := strings.Split(string(text), "/")
	if len(ss) != 2 {
		return fmt.Errorf("pair must be formatted as BASE/QUOTE, got %q", string(text))
	}
	p.Base = strings.ToUpper(ss[0])
	p.Quote = strings.ToUpper(ss[1])
	return nil
}

// Empty returns true if the pair is empty.
// Pair is considered empty if either base or quote is empty.
func (p Pair) Empty() bool {
	return p.Base == "" || p.Quote == ""
}

// Equal returns true if the pair is equal to the given pair.
func (p Pair) Equal(c Pair) bool {
	return p.Base == c.Base && p.Quote == c.Quote
}

// Invert returns an inverted pair.
// For example, if the pair is "BTC/USD", then the inverted pair is "USD/BTC".
func (p Pair) Invert() Pair {
	return Pair{
		Base:  p.Quote,
		Quote: p.Base,
	}
}

// String returns a string representation of the pair.
func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

func queryToPairs(query []any) ([]Pair, bool) {
	pairs := make([]Pair, len(query))
	for i, q := range query {
		switch q := q.(type) {
		case Pair:
			pairs[i] = q
		default:
			return nil, false
		}
	}
	return pairs, true
}
