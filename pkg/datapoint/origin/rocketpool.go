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

//go:embed rocketpool_abi.json
var rocketPoolABI []byte

const RocketPoolLoggerTag = "ROCKETPOOL_ORIGIN"

type RocketPoolOptions struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type RocketPool struct {
	client                    rpc.RPC
	contractAddresses         ContractAddresses
	abi                       *abi.Contract
	baseIndex, quoteIndex, dx *big.Int
	blocks                    []int64
	logger                    log.Logger
}

func NewRocketPool(opts RocketPoolOptions) (*RocketPool, error) {
	if opts.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if opts.Logger == nil {
		opts.Logger = null.New()
	}

	a, err := abi.ParseJSON(rocketPoolABI)
	if err != nil {
		return nil, err
	}

	return &RocketPool{
		client:            opts.Client,
		contractAddresses: opts.ContractAddresses,
		abi:               a,
		baseIndex:         big.NewInt(0),
		quoteIndex:        big.NewInt(1),
		dx:                new(big.Int).Mul(big.NewInt(1), new(big.Int).SetUint64(ether)),
		blocks:            opts.Blocks,
		logger:            opts.Logger.WithField("curve", RocketPoolLoggerTag),
	}, nil
}

//nolint:funlen,gocyclo
func (r *RocketPool) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := r.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make([]*big.Int, len(pairs))
	var calls []types.Call
	for i, pair := range pairs {
		contract, inverted, err := r.contractAddresses.AddressByPair(pair)
		if err != nil {
			return nil, err
		}

		var callData []byte
		if !inverted {
			callData, err = r.abi.Methods["getExchangeRate"].EncodeArgs()
		} else {
			callData, err = r.abi.Methods["getRethValue"].EncodeArgs(big.NewInt(0).SetUint64(ether))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get contract args for pair: %s: %w", pair.String(), err)
		}
		calls = append(calls, types.Call{
			To:    &contract,
			Input: callData,
		})
		totals[i] = new(big.Int).SetInt64(0)
	}

	for _, blockDelta := range r.blocks {
		resp, err := ethereum.MultiCall(ctx, r.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
		if err != nil {
			return nil, fmt.Errorf("failed multicall: %w", err)
		}
		if len(calls) != len(resp) {
			return nil, fmt.Errorf("unexpected number of multicall results, expected %d, got %d",
				len(calls), len(resp))
		}
		if len(resp) != len(pairs) {
			return nil, fmt.Errorf("unexpected number of multicall results with pairs, expected %d, got %d",
				len(resp), len(pairs))
		}

		for i, _ := range pairs {
			price := new(big.Int).SetBytes(resp[0][0:32])
			totals[i] = totals[i].Add(totals[i], price)
		}
	}

	for i, pair := range pairs {
		avgPrice := new(big.Float).Quo(new(big.Float).SetInt(totals[i]), new(big.Float).SetUint64(ether))
		avgPrice = avgPrice.Quo(avgPrice, new(big.Float).SetUint64(uint64(len(r.blocks))))

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
