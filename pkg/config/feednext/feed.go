package feed

import (
	"fmt"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/hashicorp/hcl/v2"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/signer"

	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/feed"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/timeutil"
)

type Config struct {
	// EthereumKey is the name of the Ethereum key to use for signing prices.
	EthereumKey string `hcl:"ethereum_key"`

	// Interval is the interval at which to publish prices in seconds.
	Interval uint32 `hcl:"interval"`

	DataModels []string `hcl:"data_models"`

	// HCL fields:
	Range   hcl.Range       `hcl:",range"`
	Content hcl.BodyContent `hcl:",content"`

	// Configured service:
	feed *feed.Feed
}

type Dependencies struct {
	KeysRegistry ethereumConfig.KeyRegistry
	DataProvider datapoint.Provider
	Transport    transport.Transport
	Logger       log.Logger
}

func (c *Config) ConfigureFeed(d Dependencies) (*feed.Feed, error) {
	if c.feed != nil {
		return c.feed, nil
	}
	if c.Interval == 0 {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Summary:  "Validation error",
			Detail:   "Interval cannot be zero",
			Severity: hcl.DiagError,
			Subject:  c.Content.Attributes["interval"].Range.Ptr(),
		}}
	}
	ethereumKey, ok := d.KeysRegistry[c.EthereumKey]
	if !ok {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error",
			Detail:   fmt.Sprintf("Ethereum key %q is not configured", c.EthereumKey),
			Subject:  c.Content.Attributes["ethereum_key"].Range.Ptr(),
		}
	}
	cfg := feed.Config{
		DataModels:   c.DataModels,
		DataProvider: d.DataProvider,
		Signers:      []datapoint.Signer{signer.NewTick(ethereumKey, crypto.ECRecoverer)},
		Transport:    d.Transport,
		Logger:       d.Logger,
		Interval:     timeutil.NewTicker(time.Second * time.Duration(c.Interval)),
	}
	feedService, err := feed.New(cfg)
	if err != nil {
		return nil, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Runtime error",
			Detail:   fmt.Sprintf("Failed to create the ConfigureFeed service: %v", err),
			Subject:  c.Range.Ptr(),
		}
	}
	c.feed = feedService
	return feedService, nil
}
