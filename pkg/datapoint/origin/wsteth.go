package origin

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

//go:embed wsteth_abi.json
var wstethABI []byte

const WrappedStakedETHLoggerTag = "WSTETH_ORIGIN"

type WrappedStakedETHConfig struct {
	Client            rpc.RPC
	ContractAddresses map[string]string
	Logger            log.Logger
	Blocks            []int64
}

type WrappedStakedETH struct {
	client            rpc.RPC
	contractAddresses ContractAddresses
	abi               *abi.Contract
	blocks            []int64
	logger            log.Logger
}

func NewWrappedStakedETH(config WrappedStakedETHConfig) (*WrappedStakedETH, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	a, err := abi.ParseJSON(wstethABI)
	if err != nil {
		return nil, err
	}
	addresses, err := convertAddressMap(config.ContractAddresses)
	if err != nil {
		return nil, err
	}

	return &WrappedStakedETH{
		client:            config.Client,
		contractAddresses: addresses,
		abi:               a,
		blocks:            config.Blocks,
		logger:            config.Logger.WithField("wsteth", WrappedStakedETHLoggerTag),
	}, nil
}

//nolint:funlen
func (w *WrappedStakedETH) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := w.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make([]*big.Int, len(pairs))
	var calls []types.Call
	for i, pair := range pairs {
		contract, inverted, err := w.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}
		if inverted {
			points[pair] = datapoint.Point{Error: fmt.Errorf(
				"cannot use inverted pair to retrieve price: %s", pair.String())}
			continue
		}

		callData, err := w.abi.Methods["stEthPerToken"].EncodeArgs()
		if err != nil {
			points[pair] = datapoint.Point{Error: fmt.Errorf(
				"failed to get contract args for pair: %s: %w",
				pair.String(),
				err,
			)}
			continue
		}
		calls = append(calls, types.Call{
			To:    &contract,
			Input: callData,
		})
		totals[i] = new(big.Int).SetInt64(0)
	}

	if len(calls) > 0 {
		for _, blockDelta := range w.blocks {
			resp, err := ethereum.MultiCall(ctx, w.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
			if err != nil {
				return nil, err
			}

			n := 0
			for i := 0; i < len(pairs); i++ {
				if points[pairs[i]].Error != nil {
					continue
				}
				price := new(big.Int).SetBytes(resp[n][0:32])
				totals[i] = totals[i].Add(totals[i], price)
				n++
			}
		}
	}

	for i, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}
		avgPrice := new(big.Float).Quo(new(big.Float).SetInt(totals[i]), new(big.Float).SetUint64(ether))
		avgPrice = avgPrice.Quo(avgPrice, new(big.Float).SetUint64(uint64(len(w.blocks))))

		tick := value.Tick{
			Pair:      pair,
			Price:     bn.Float(avgPrice),
			Volume24h: nil,
		}
		points[pair] = datapoint.Point{
			Value: tick,
			Time:  time.Now(),
		}
	}

	return points, nil
}
