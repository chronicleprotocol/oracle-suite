package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/treerender"
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
	Model(ctx context.Context, model string) (Model, error)

	// Models describes price models which are used to calculate prices.
	// If no pairs are specified, models for all pairs are returned.
	Models(ctx context.Context, models ...string) (map[string]Model, error)
}

// Model is a simplified representation of a model which is used to calculate
// asset pair prices. The main purpose of this structure is to help the end
// user to understand how prices are derived and calculated.
//
// This structure is purely informational. The way it is used depends on
// a specific implementation.
type Model struct {
	// Meta contains metadata for the model. It should contain information
	// about the model and its parameters.
	Meta Meta

	// Pair is an asset pair for which this model returns a price.
	Pair Pair

	// Models is a list of sub models used to calculate price.
	Models []Model
}

// MarshalText implements the encoding.TextMarshaler interface.
func (m Model) MarshalText() ([]byte, error) {
	return []byte(m.Pair.String()), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Pair.String())
}

// MarshalTrace returns a human-readable representation of the model.
func (m Model) MarshalTrace() ([]byte, error) {
	return treerender.RenderTree(func(node any) treerender.NodeData {
		model := node.(Model)
		meta := model.Meta.Meta()
		typ := "node"
		if n, ok := meta["type"].(string); ok {
			typ = n
			delete(meta, "type")
		}
		meta["pair"] = model.Pair
		var models []any
		for _, m := range model.Models {
			models = append(models, m)
		}
		return treerender.NodeData{
			Name:      typ,
			Params:    meta,
			Ancestors: models,
			Error:     nil,
		}
	}, []any{m}, 0), nil
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

	// SubTicks is a list of sub ticks that are used to obtain this tick.
	SubTicks []Tick

	// Meta contains metadata for the tick. It may contain additional
	// information about the tick and a price model.
	Meta Meta

	// Error is an optional error which occurred during obtaining the price.
	// If error is not nil, then the price is invalid and should not be used.
	//
	// Tick may be invalid for other reasons, hence you should always check
	// the tick for validity by calling Tick.Validate() method.
	Error error
}

// Validate returns an error if the tick is invalid.
func (t Tick) Validate() error {
	if t.Error != nil {
		return t.Error
	}
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
	if t.Time.IsZero() {
		return fmt.Errorf("time is not set")
	}
	if t.Volume24h != nil && t.Volume24h.Sign() < 0 {
		return fmt.Errorf("volume is negative")
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (t Tick) String() string {
	return fmt.Sprintf(
		"%s(price: %s, volume: %s, time: %s, error: %v)",
		t.Pair,
		t.Price,
		t.Volume24h,
		t.Time,
		t.Error,
	)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t Tick) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s: %s", t.Pair, t.Price)), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t Tick) MarshalJSON() ([]byte, error) {
	return json.Marshal(tickToJSON(t))
}

// MarshalTrace returns a human-readable representation of the tick.
func (t Tick) MarshalTrace() ([]byte, error) {
	return treerender.RenderTree(func(node any) treerender.NodeData {
		tick := node.(Tick)
		err := tick.Validate()
		meta := tick.Meta.Meta()
		typ := "tick"
		if n, ok := meta["type"].(string); ok {
			typ = n
			delete(meta, "type")
		}
		meta["pair"] = tick.Pair
		meta["price"] = tick.Price
		meta["time"] = tick.Time.In(time.UTC).Format(time.RFC3339Nano)
		var ticks []any
		for _, t := range tick.SubTicks {
			ticks = append(ticks, t)
		}
		return treerender.NodeData{
			Name:      typ,
			Params:    meta,
			Ancestors: ticks,
			Error:     err,
		}
	}, []any{t}, 0), nil
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

type jsonTick struct {
	Base       string         `json:"base"`
	Quote      string         `json:"quote"`
	Price      float64        `json:"price"`
	Volume24h  float64        `json:"vol24h"`
	Timestamp  time.Time      `json:"ts"`
	Parameters map[string]any `json:"params,omitempty"`
	Prices     []jsonTick     `json:"prices,omitempty"`
	Error      string         `json:"error,omitempty"`
}

func tickToJSON(t Tick) jsonTick {
	var prices []jsonTick
	var errStr string
	for _, tick := range t.SubTicks {
		prices = append(prices, tickToJSON(tick))
	}
	if err := t.Validate(); err != nil {
		errStr = err.Error()
	}
	price := t.Price
	volume := t.Volume24h
	if price == nil {
		price = bn.Float(0)
	}
	if volume == nil {
		volume = bn.Float(0)
	}
	return jsonTick{
		Base:       t.Pair.Base,
		Quote:      t.Pair.Quote,
		Price:      price.Float64(),
		Volume24h:  volume.Float64(),
		Timestamp:  t.Time.In(time.UTC),
		Parameters: t.Meta.Meta(),
		Prices:     prices,
		Error:      errStr,
	}
}
