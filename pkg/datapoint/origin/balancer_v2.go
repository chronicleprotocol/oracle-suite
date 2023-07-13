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

//go:embed balancerv2_abi.json
var balancerV2PoolABI []byte

const BalancerV2LoggerTag = "BALANCERV2_ORIGIN"

type BalancerV2Config struct {
	Client             rpc.RPC
	ContractAddresses  map[string]string
	ReferenceAddresses map[string]string
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

func NewBalancerV2(config BalancerV2Config) (*BalancerV2, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	a, err := abi.ParseJSON(balancerV2PoolABI)
	if err != nil {
		return nil, err
	}
	addresses, err := convertAddressMap(config.ContractAddresses)
	if err != nil {
		return nil, err
	}
	refAddresses, err := convertAddressMap(config.ReferenceAddresses)
	if err != nil {
		return nil, err
	}

	return &BalancerV2{
		client:             config.Client,
		contractAddresses:  addresses,
		referenceAddresses: refAddresses,
		abi:                a,
		variable:           0, // PAIR_PRICE
		blocks:             config.Blocks,
		logger:             config.Logger.WithField("balancerV2", BalancerV2LoggerTag),
	}, nil
}

//nolint:funlen,gocyclo
func (b *BalancerV2) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
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
	refs := make([]*big.Int, len(pairs))
	var calls []types.Call
	for i, pair := range pairs {
		contract, inverted, err := b.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}
		if inverted {
			points[pair] = datapoint.Point{Error: fmt.Errorf("cannot use inverted pair to retrieve price: %s",
				pair.String())}
			continue
		}

		// Calls for `getLatest`
		callData, err := b.abi.Methods["getLatest"].EncodeArgs(b.variable)
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

		ref, inverted, err := b.referenceAddresses.ByPair(pair)
		if err == nil {
			if inverted {
				points[pair] = datapoint.Point{Error: fmt.Errorf(
					"cannot use inverted pair to retrieve price: %s",
					pair.String(),
				)}
				continue
			}
			callData, err := b.abi.Methods["getPriceRateCache"].EncodeArgs(types.MustAddressFromHex(ref.String()))
			if err != nil {
				points[pair] = datapoint.Point{Error: fmt.Errorf(
					"failed to pack contract args for getPriceRateCache (pair %s): %w",
					pair.String(),
					err,
				)}
				continue
			}
			calls = append(calls, types.Call{
				To:    &contract,
				Input: callData,
			})
		}

		totals[i] = new(big.Int).SetInt64(0)
		refs[i] = new(big.Int).SetInt64(0)
	}

	if len(calls) > 0 {
		for _, blockDelta := range b.blocks {
			resp, err := ethereum.MultiCall(ctx, b.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
			if err != nil {
				return nil, err
			}

			n := 0
			for i := 0; i < len(pairs); i++ {
				if points[pairs[i]].Error != nil {
					continue
				}
				price := new(big.Int).SetBytes(resp[n][0:32])
				_, _, err := b.referenceAddresses.ByPair(pairs[i])
				if err == nil {
					refPrice := new(big.Int).SetBytes(resp[n+1][0:32])
					refs[i] = new(big.Int).Add(refs[i], refPrice)
					n++ // next response was already used, ignore
				}
				totals[i] = new(big.Int).Add(totals[i], price)
				n++
			}
		}
	}

	for i, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}
		avgPrice := new(big.Float).Quo(new(big.Float).SetInt(totals[i]), new(big.Float).SetUint64(ether))
		avgPrice = new(big.Float).Quo(avgPrice, new(big.Float).SetUint64(uint64(len(b.blocks))))
		if refs[i].Cmp(big.NewInt(0)) > 0 { // Non Zero, then multiply with ref price
			refPrice := new(big.Float).Quo(new(big.Float).SetInt(refs[i]), new(big.Float).SetUint64(ether))
			avgPrice = new(big.Float).Quo(
				new(big.Float).Mul(avgPrice, refPrice),
				new(big.Float).SetUint64(uint64(len(b.blocks))),
			)
		}

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
