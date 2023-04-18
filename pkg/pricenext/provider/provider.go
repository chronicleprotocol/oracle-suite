package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

// Provider provides prices for asset pairs.
type Provider interface {
	// ModelNames returns a list of supported price models.
	ModelNames(ctx context.Context) []string

	// Tick returns a price for the given asset pair.
	Tick(ctx context.Context, model string) (Tick, error)

	// Ticks returns prices for the given asset pairs.
	Ticks(ctx context.Context, models ...string) (map[string]Tick, error)

	// Model returns a price model for the given asset pair.
	Model(ctx context.Context, model string) (*Model, error)

	// Models describes price models which are used to calculate prices.
	// If no pairs are specified, models for all pairs are returned.
	Models(ctx context.Context, models ...string) (map[string]*Model, error)
}

// Model is a simplified representation of a model which is used to calculate
// asset pair prices. The main purpose of this structure is to help the end
// user to understand how prices are derived and calculated.
//
// This structure is purely informational. The way it is used depends on
// a specific implementation.
type Model struct {
	// Type is a model type, e.g. "origin", "median", etc.
	Type string

	// Meta is an optional metadata for the model.
	Meta Meta

	// Pair is an asset pair for which this model returns a price.
	Pair Pair

	// Models is a list of sub models used to calculate price.
	Models []*Model
}

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

	// Time is the time of the price (usually the time of the last trade)
	// reported by the provider or, if not available, the time when the price
	// was obtained.
	Time time.Time

	// Meta is an optional metadata for the price.
	Meta Meta

	// Warning is an optional error which occurred during obtaining the price
	// but does not affect the price calculation process.
	//
	// To check if the price is valid, use Tick.Validate() method.
	Warning error

	// Error is an optional error which occurred during obtaining the price.
	// If error is not nil, then the price is invalid and should not be used.
	//
	// To check if the price is valid, use Tick.Validate() method.
	Error error
}

// Validate returns an error if the tick is invalid.
func (t Tick) Validate() error {
	if t.Error != nil {
		return t.Error
	}
	if t.Pair.Base == "" {
		return fmt.Errorf("base is empty")
	}
	if t.Pair.Quote == "" {
		return fmt.Errorf("quote is empty")
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
	if t.Time.IsZero() {
		return fmt.Errorf("time is zero")
	}
	if t.Volume24h != nil && t.Volume24h.Sign() < 0 {
		return fmt.Errorf("volume is negative")
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (t Tick) String() string {
	return fmt.Sprintf(
		"%s(price: %s, volume: %s, time: %s, warning: %v, error: %v)",
		t.Pair,
		t.Price,
		t.Volume24h,
		t.Time,
		t.Warning,
		t.Error,
	)
}

// Meta is an additional metadata for a price or a model.
type Meta interface {
	// Meta returns a map of metadata.
	Meta() map[string]any
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
func (p Pair) Empty() bool {
	return p.Base == "" && p.Quote == ""
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
