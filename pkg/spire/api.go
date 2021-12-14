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
	"strings"

	"github.com/chronicleprotocol/oracle-suite/pkg/datastore"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/log"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

type Nothing = struct{}

type API struct {
	transport transport.Transport
	datastore datastore.Datastore
	signer    ethereum.Signer
	log       log.Logger
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

	return n.transport.Broadcast(messages.PriceMessageName, arg.Price)
}

func (n *API) PullPrices(arg *PullPricesArg, resp *PullPricesResp) error {
	n.log.
		WithField("assetPair", arg.FilterAssetPair).
		WithField("feeder", arg.FilterFeeder).
		Info("Pull prices")

	var prices []*messages.Price
	for fp, p := range n.datastore.Prices().All() {
		if arg.FilterAssetPair != "" && arg.FilterAssetPair != fp.AssetPair {
			continue
		}
		if arg.FilterFeeder != "" && !strings.EqualFold(arg.FilterFeeder, fp.Feeder.String()) {
			continue
		}
		prices = append(prices, p)
	}

	*resp = PullPricesResp{Prices: prices}

	return nil
}

func (n *API) PullPrice(arg *PullPriceArg, resp *PullPriceResp) error {
	n.log.
		WithField("assetPair", arg.AssetPair).
		WithField("feeder", arg.Feeder).
		Info("Pull price")

	*resp = PullPriceResp{
		Price: n.datastore.Prices().Feeder(arg.AssetPair, ethereum.HexToAddress(arg.Feeder)),
	}

	return nil
}
