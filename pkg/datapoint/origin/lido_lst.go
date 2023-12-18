//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package origin

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/defiweb/go-eth/rpc"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/maputil"
)

const LidoLSTLoggerTag = "LIDO_LST_ORIGIN"

type LidoLSTConfig struct {
	Client            rpc.RPC
	ContractAddresses ContractAddresses
	Logger            log.Logger
	Blocks            []int64
}

type LidoLST struct {
	ctx    context.Context
	client rpc.RPC
	stETH  types.Address
	logger log.Logger
}

func NewLidoLST(config LidoLSTConfig) (*LidoLST, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("ethereum client not set")
	}
	if config.Logger == nil {
		config.Logger = null.New()
	}
	if len(maputil.Keys(config.ContractAddresses)) < 1 {
		return nil, fmt.Errorf("contract address not set")
	}
	keys := maputil.Keys(config.ContractAddresses)

	return &LidoLST{
		client: config.Client,
		stETH:  config.ContractAddresses[keys[0]],
		logger: config.Logger.WithField("lido_lst", LidoLSTLoggerTag),
	}, nil
}

type rebaseEvent struct {
	blockNumber     *big.Int
	reportTimestamp *big.Int
	preTotalEther   *big.Int
	preTotalShares  *big.Int
	postTotalEther  *big.Int
	postTotalShares *big.Int
	timeElapsed     *big.Int
}

func (r *LidoLST) getRebaseEvents(fromBlock, toBlock *types.BlockNumber) ([]rebaseEvent, error) {
	filter := types.NewFilterLogsQuery().
		SetAddresses(r.stETH).
		SetFromBlock(fromBlock).
		SetToBlock(toBlock).
		SetTopics([]types.Hash{tokenRebased.Topic0()})
	logs, err := r.client.GetLogs(r.ctx, *filter)
	if err != nil {
		return nil, err
	}

	var events []rebaseEvent
	for _, log := range logs {
		var (
			reportTimestamp, timeElapsed,
			preTotalShares, preTotalEther, postTotalShares, postTotalEther,
			sharesMintedAsFees *big.Int
		)
		if err := tokenRebased.DecodeValues(log.Topics, log.Data,
			&reportTimestamp, &timeElapsed,
			&preTotalShares, &preTotalEther, &postTotalShares, &postTotalEther,
			&sharesMintedAsFees); err != nil {
			return nil, err
		}
		events = append(events, rebaseEvent{
			blockNumber:     log.BlockNumber,
			reportTimestamp: reportTimestamp,
			preTotalEther:   preTotalEther,
			preTotalShares:  preTotalShares,
			postTotalEther:  postTotalEther,
			postTotalShares: postTotalShares,
			timeElapsed:     timeElapsed,
		})
	}
	return events, nil
}

func (r *LidoLST) getLastRebaseEvent(block uint64) (*rebaseEvent, error) {
	const BlocksByDay = uint64(7600)
	const DayLimit = uint64(7)
	const DefaultStepBlock = 50000

	// https://etherscan.io/tx/0x251f8cc3fea4be64c1f9a9afd8ba5c03472f61285ffdb895ab32cc9d339c836e#eventlog
	toBlock := block
	fromBlock := toBlock - BlocksByDay*2 // fetch 2 days events
	events, err := requestWithBlockStep[rebaseEvent](fromBlock, toBlock, DefaultStepBlock, r.getRebaseEvents)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		return &events[len(events)-1], nil
	}
	toBlock = block
	fromBlock = toBlock - BlocksByDay*DayLimit
	events, err = requestWithBlockStep[rebaseEvent](fromBlock, toBlock, DefaultStepBlock, r.getRebaseEvents)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		return &events[len(events)-1], nil
	}
	return nil, fmt.Errorf("not found TokenRebased event")
}

func (r *LidoLST) calculateAprFromRebaseEvent(event rebaseEvent) *bn.DecFloatPointNumber {
	// Reference: https://github.com/lidofinance/lido-ethereum-sdk/blob/main/packages/sdk/src/statistics/apr.ts#L17
	// LidoSDKApr::calculateAprFromRebaseEvent
	const lidoDecimals = 27
	const secondsInYearInt = 31536000
	const pointDecimals = 18
	tenTo27 := new(big.Float).Quo(big.NewFloat(10), big.NewFloat(lidoDecimals))
	preShareRate := bn.DecFloatPoint(event.preTotalEther).
		Mul(bn.DecFloatPoint(tenTo27)).
		Div(bn.DecFloatPoint(event.preTotalShares))
	postShareRate := bn.DecFloatPoint(event.postTotalEther).
		Mul(bn.DecFloatPoint(tenTo27)).
		Div(bn.DecFloatPoint(event.postTotalShares))
	mulForPrecision := bn.DecFloatPoint(1000)
	secondsInYear := bn.DecFloatPoint(secondsInYearInt)
	userAPR := secondsInYear.Mul(postShareRate.Sub(preShareRate).Mul(mulForPrecision)).
		Div(preShareRate).
		Div(bn.DecFloatPoint(event.timeElapsed))

	return userAPR.Mul(bn.DecFloatPoint(100)).Div(mulForPrecision).SetPrec(pointDecimals)
}

func (r *LidoLST) FetchDataPoints(ctx context.Context, query []any) (map[any]datapoint.Point, error) {
	pairs, ok := queryToPairs(query)
	if !ok {
		return nil, fmt.Errorf("invalid query type: %T, expected []Pair", query)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})

	r.ctx = ctx
	pair := pairs[0]

	// Remove `DAYS` or `DAY` suffix and extract number of days
	daysStr := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(pair.Quote), "days", ""), "day", "")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return nil, fmt.Errorf("quote token should be `nDAYS`, n is digit")
	}

	block, err := r.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get block number, %w", err)
	}
	lastRebaseEvent, err := r.getLastRebaseEvent(block.Uint64())
	if err != nil {
		return nil, err
	}

	apr := r.calculateAprFromRebaseEvent(*lastRebaseEvent)
	for i := 0; i < days-1; i++ {
		lastRebaseEvent, _ = r.getLastRebaseEvent(lastRebaseEvent.blockNumber.Uint64() - uint64(1))
		lastAPR := r.calculateAprFromRebaseEvent(*lastRebaseEvent)
		apr = apr.Add(lastAPR)
	}

	points := make(map[any]datapoint.Point)
	points[pair] = datapoint.Point{
		Value: value.NewTick(pair, apr.Div(bn.DecFloatPoint(days)), nil),
		Time:  time.Now(),
	}

	return points, nil
}
