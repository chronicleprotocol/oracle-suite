package origin

import (
	"context"
	"fmt"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/defiweb/go-eth/rpc"
)

const TickGenericWebLoggerTag = "TICK_GENERIC_WEB3_ORIGIN"

var averageFromBlocks = []int64{0, 10, 20}

const ether = 1e18

type ExchangeHandler interface {
	FetchDataPoints(ctx context.Context, pairs []value.Pair) (map[any]datapoint.Point, error)
}

type TickGenericWeb3Options struct {
	Protocol          string
	Clients           []rpc.RPC
	ContractAddresses []ContractAddresses
	Logger            log.Logger
}

type TickGenericWeb3 struct {
	handler ExchangeHandler
	logger  log.Logger
}

func NewTickGenericWeb3(opts TickGenericWeb3Options) (*TickGenericWeb3, error) {
	if len(opts.Clients) < 1 {
		return nil, fmt.Errorf("rpc clients can not be empty")
	}

	if len(opts.Clients) != len(opts.ContractAddresses) {
		return nil, fmt.Errorf("contract addresses are mismatched with clients")
	}

	inst := &TickGenericWeb3{}
	var err error
	logger := opts.Logger.WithField(opts.Protocol, TickGenericWebLoggerTag)
	if opts.Protocol == "balancerV2" {
		inst.handler, err = NewBalancerV2(opts.Clients[0], opts.ContractAddresses[0], averageFromBlocks, logger)
	}

	if err != nil {
		return nil, err
	}
	if inst.handler == nil {
		return nil, fmt.Errorf("not supported protocol")
	}

	return inst, nil
}

func (g *TickGenericWeb3) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	if g.handler != nil {
		pairs, ok := queryToPairs(query)
		if !ok {
			return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
		}
		return g.handler.FetchDataPoints(ctx, pairs)
	}
	return nil, fmt.Errorf("not found handler")
}
