package origin

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"
)

//go:embed balancerv2_abi.json
var balancerV2PoolABI []byte

const BalancerV2LoggerTag = "BALANCERV2_ORIGIN"

type BalancerV2Options struct {
	Client             rpc.RPC
	ContractAddresses  ContractAddresses
	ReferenceAddresses ContractAddresses
	Logger             log.Logger
	Blocks             []int64
}

type BalancerV2 struct {
	client             rpc.RPC
	contractAddresses  ContractAddresses
	referenceAddresses ContractAddresses
	abi                *abi.Contract
	variable           byte
	blocks             []int64
	logger             log.Logger
}

func NewBalancerV2(opts BalancerV2Options) (*BalancerV2, error) {
	if opts.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if opts.Logger == nil {
		opts.Logger = null.New()
	}

	a, err := abi.ParseJSON(balancerV2PoolABI)
	if err != nil {
		return nil, err
	}
	return &BalancerV2{
		client:             opts.Client,
		contractAddresses:  opts.ContractAddresses,
		referenceAddresses: opts.ReferenceAddresses,
		abi:                a,
		variable:           0, // PAIR_PRICE
		blocks:             opts.Blocks,
		logger:             opts.Logger.WithField("balancerV2", BalancerV2LoggerTag),
	}, nil
}

func (b *BalancerV2) FetchDataPoints(
	ctx context.Context,
	query []any,
) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := b.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make([]*big.Int, len(pairs))
	var calls []types.Call
	for i, pair := range pairs {
		contract, inverted, err := b.contractAddresses.AddressByPair(pair)
		if err != nil {
			return nil, err
		}
		if inverted {
			return nil, fmt.Errorf("cannnot use inverted pair to retrieve price: %s", pair.String())
		}

		// Calls for `getLatest`
		callData, err := b.abi.Methods["getLatest"].EncodeArgs(b.variable)
		if err != nil {
			return nil, fmt.Errorf("failed to pack contract args for getLatest (pair %s): %w", pair.String(), err)
		}
		calls = append(calls, types.Call{
			To:    &contract,
			Input: callData,
		})

		ref, inverted, err := b.referenceAddresses.AddressByPair(pair)
		if err == nil {
			if inverted {
				return nil, fmt.Errorf("cannot use inverted pair to retrieve price: %s", pair.String())
			}
			callData, err := b.abi.Methods["getPriceRateCache"].EncodeArgs(types.MustAddressFromHex(ref.String()))
			if err != nil {
				return nil, fmt.Errorf(
					"failed to pack contract args for getPriceRateCache (pair %s): %w",
					pair.String(),
					err,
				)
			}
			calls = append(calls, types.Call{
				To:    &contract,
				Input: callData,
			})
		}

		totals[i] = new(big.Int).SetInt64(0)
	}

	for _, blockDelta := range b.blocks {
		resp, err := ethereum.MultiCall(ctx, b.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
		if err != nil {
			return nil, fmt.Errorf("failed multicall for getLatest: %w", err)
		}

		n := 0
		for i := 0; i < len(pairs); i++ {
			price := new(big.Int).SetBytes(resp[n][0:32])

			_, _, err := b.referenceAddresses.AddressByPair(pairs[i])
			if err == nil {
				refPrice := new(big.Int).SetBytes(resp[n+1][0:32])
				price = price.Quo(price.Mul(price, refPrice), new(big.Int).SetUint64(ether))
				n++ // next response was already used, ignore
			}

			totals[i] = totals[i].Add(totals[i], price)
			n++
		}
	}

	for i, pair := range pairs {
		avgPrice := new(big.Float).Quo(new(big.Float).SetInt(totals[i]), new(big.Float).SetUint64(ether))
		avgPrice = avgPrice.Quo(avgPrice, new(big.Float).SetUint64(uint64(len(b.blocks))))

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
