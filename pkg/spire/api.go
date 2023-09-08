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

package spire

import (
	"context"
	"fmt"
	"time"

	"github.com/defiweb/go-eth/crypto"
	"github.com/defiweb/go-eth/types"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const defaultRPCTimeout = time.Minute

type Nothing = struct{}

type API struct {
	transport  transport.Service
	priceStore *store.Store
	recover    crypto.Recoverer
	log        log.Logger
}

type PublishArg struct {
	DataPoint *messages.DataPoint
}

type PullPricesArg struct {
	FilterAssetPair string
	FilterFeed      string
}

type PullDataPointsResp struct {
	DataPoints []*messages.DataPoint
}

type PullPriceArg struct {
	AssetPair string
	Feed      string
}

type PullDataPointResp struct {
	DataPoint *messages.DataPoint
}

func (n *API) Publish(arg *PublishArg, _ *Nothing) error {
	n.log.
		WithField("model", arg.DataPoint.Model).
		Info("Publish data point")

	return n.transport.Broadcast(messages.DataPointV1MessageName, arg.DataPoint)
}

func (n *API) PullPoints(arg *PullPricesArg, resp *PullDataPointsResp) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer ctxCancel()

	n.log.
		WithField("model", arg.FilterAssetPair).
		WithField("feed", arg.FilterFeed).
		Info("Pull data points")

	var dataPoints []*messages.DataPoint

	switch {
	case arg.FilterAssetPair != "" && arg.FilterFeed != "":
		price, _, err := n.priceStore.LatestFrom(ctx, types.MustAddressFromHex(arg.FilterFeed), arg.FilterAssetPair)
		if err != nil {
			return err
		}
		point := &messages.DataPoint{
			Model:          price.Model,
			Point:          price.DataPoint,
			ECDSASignature: price.Signature,
		}
		dataPoints = []*messages.DataPoint{point}
	case arg.FilterAssetPair != "":
		points, err := n.priceStore.Latest(ctx, arg.FilterAssetPair)
		if err != nil {
			return err
		}
		for _, p := range points {
			point := &messages.DataPoint{
				Model:          p.Model,
				Point:          p.DataPoint,
				ECDSASignature: p.Signature,
			}
			dataPoints = append(dataPoints, point)
		}
	default:
		return fmt.Errorf("please provide model")
	}

	*resp = PullDataPointsResp{DataPoints: dataPoints}

	return nil
}

func (n *API) PullPoint(arg *PullPriceArg, resp *PullDataPointResp) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer ctxCancel()

	n.log.
		WithField("model", arg.AssetPair).
		WithField("feed", arg.Feed).
		Info("Pull price")

	price, ok, err := n.priceStore.LatestFrom(ctx, types.MustAddressFromHex(arg.Feed), arg.AssetPair)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	point := &messages.DataPoint{
		Model:          price.Model,
		Point:          price.DataPoint,
		ECDSASignature: price.Signature,
	}
	*resp = PullDataPointResp{DataPoint: point}

	return nil
}
