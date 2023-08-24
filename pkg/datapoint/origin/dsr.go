package origin

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
)

const DSRLoggerTag = "DSR_ORIGIN"

type DSRConfig struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type DSR struct {
	client            rpc.RPC
	contractAddresses ContractAddresses
	blocks            []int64
	logger            log.Logger
}

func NewDSR(config DSRConfig) (*DSR, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("cannot nil ethereum client")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}

	return &DSR{
		client:            config.Client,
		contractAddresses: config.ContractAddresses,
		blocks:            config.Blocks,
		logger:            config.Logger.WithField("dsr", DSRLoggerTag),
	}, nil
}

func (d *DSR) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	points := make(map[any]datapoint.Point)

	block, err := d.client.BlockNumber(ctx)

	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}

	totals := make([]*big.Int, len(pairs))
	var calls []types.Call
	for i, pair := range pairs {
		contract, _, _, err := d.contractAddresses.ByPair(pair)
		if err != nil {
			points[pair] = datapoint.Point{Error: err}
			continue
		}

		// Reference: https://github.com/makerdao/community/blob/master/content/en/faqs/dsr.md
		// The Dai Savings Rate (DSR) is a variable rate of accrual earned by locking Dai in the DSR smart contract.
		// Dai holders can earn savings automatically and natively while retaining control of their Dai.
		//
		// Reference: https://docs.makerdao.com/smart-contract-modules/rates-module/pot-detailed-documentation
		// The Pot contract is the core of theDai Savings Rate.
		// It allows users to deposit dai and activate the Dai Savings Rate and earning savings on their dai.
		// dsr - the dai savings rate. It starts as 1 (ONE = 10^27), but can be updated by governance.
		callData, err := dsr.EncodeArgs()
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
		for _, blockDelta := range d.blocks {
			resp, err := ethereum.MultiCall(ctx, d.client, calls, types.BlockNumberFromUint64(uint64(block.Int64()-blockDelta)))
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

	const decimals = 27
	scaleUp := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil))
	for i, pair := range pairs {
		if points[pair].Error != nil {
			continue
		}
		avgPrice := new(big.Float).Quo(new(big.Float).SetInt(totals[i]), scaleUp)
		avgPrice = avgPrice.Quo(avgPrice, new(big.Float).SetUint64(uint64(len(d.blocks))))

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
