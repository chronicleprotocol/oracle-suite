//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
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
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

const defaultRPCTimeout = time.Minute

type Nothing = struct{}

type API struct {
	transport  transport.Transport
	priceStore *store.PriceStore
	signer     ethereum.Signer
	log        log.Logger
}

type PublishPriceArg struct {
	Price *messages.Price
}

type PullPricesArg struct {
	FilterAssetPair string
	FilterFeeder    string
}

type PullPricesResp struct {
	Prices []*messages.Price
}

type PullPriceArg struct {
	AssetPair string
	Feeder    string
}

type PullPriceResp struct {
	Price *messages.Price
}

func (n *API) PublishPrice(arg *PublishPriceArg, _ *Nothing) error {
	n.log.
		WithFields(arg.Price.Price.Fields(n.signer)).
		Info("Publish price")

	if err := n.transport.Broadcast(messages.PriceV0MessageName, arg.Price.AsV0()); err != nil {
		return err
	}
	if err := n.transport.Broadcast(messages.PriceV1MessageName, arg.Price.AsV1()); err != nil {
		return err
	}

	return nil
}

func (n *API) PullPrices(arg *PullPricesArg, resp *PullPricesResp) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer ctxCancel()

	n.log.
		WithField("assetPair", arg.FilterAssetPair).
		WithField("feeder", arg.FilterFeeder).
		Info("Pull prices")

	var err error
	var prices []*messages.Price

	switch {
	case arg.FilterAssetPair != "" && arg.FilterFeeder != "":
		price, err := n.priceStore.GetByFeeder(ctx, arg.FilterAssetPair, ethereum.HexToAddress(arg.FilterFeeder))
		if err != nil {
			return err
		}
		prices = []*messages.Price{price}
	case arg.FilterAssetPair != "":
		prices, err = n.priceStore.GetByAssetPair(ctx, arg.FilterAssetPair)
		if err != nil {
			return err
		}
	case arg.FilterFeeder != "":
		feederPrices, err := n.priceStore.GetAll(ctx)
		if err != nil {
			return err
		}
		for fp, price := range feederPrices {
			if strings.EqualFold(arg.FilterFeeder, fp.Feeder.String()) {
				prices = append(prices, price)
			}
		}
	}

	*resp = PullPricesResp{Prices: prices}

	return nil
}

func (n *API) PullPrice(arg *PullPriceArg, resp *PullPriceResp) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), defaultRPCTimeout)
	defer ctxCancel()

	n.log.
		WithField("assetPair", arg.AssetPair).
		WithField("feeder", arg.Feeder).
		Info("Pull price")

	price, err := n.priceStore.GetByFeeder(ctx, arg.AssetPair, ethereum.HexToAddress(arg.Feeder))
	if err != nil {
		return err
	}

	*resp = PullPriceResp{Price: price}

	return nil
}
