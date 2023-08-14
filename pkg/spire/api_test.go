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
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint"
	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/value"
	"github.com/chronicleprotocol/oracle-suite/pkg/util/bn"
	"github.com/defiweb/go-eth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/datapoint/store"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var (
	testAddress     = types.MustAddressFromHex("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	testPriceAAABBB = &messages.DataPoint{
		Model: "AAA/BBB",
		Value: datapoint.Point{
			Value: value.StaticValue{Value: bn.Float(10.0)},
			Time:  time.Unix(1234567890, 0),
			Meta: map[string]any{
				"addr": testAddress.String(),
			},
		},
		Signature: types.MustSignatureFromBytes(bytes.Repeat([]byte{0x01}, 65)),
	}
	agent      *Agent
	spire      *Client
	priceStore *store.Store
	ctxCancel  context.CancelFunc
)

func newTestInstances() (*Agent, *Client) {
	var err error
	var ctx context.Context
	ctx, ctxCancel = context.WithCancel(context.Background())

	log := null.New()
	rec := &mocks.Recoverer{}
	tra := local.New([]byte("test"), 0, map[string]transport.Message{
		messages.DataPointV1MessageName: (*messages.DataPoint)(nil),
	})
	_ = tra.Start(ctx)
	priceStore, err = store.New(store.Config{
		Storage:    store.NewMemoryStorage(),
		Transport:  tra,
		Models:     []string{"AAA/BBB", "XXX/YYY"},
		Logger:     null.New(),
		Recoverers: []datapoint.Recoverer{rec},
	})
	if err != nil {
		panic(err)
	}

	rec.On("RecoverMessage", mock.Anything, mock.Anything).Return(&testAddress, nil)

	agt, err := NewAgent(AgentConfig{
		PriceStore: priceStore,
		Transport:  tra,
		Address:    "127.0.0.1:0",
		Logger:     log,
	})
	if err != nil {
		panic(err)
	}
	err = priceStore.Start(ctx)
	if err != nil {
		panic(err)
	}
	err = agt.Start(ctx)
	if err != nil {
		panic(err)
	}

	cli, err := NewClient(ClientConfig{
		Address: agt.srv.Addr().String(),
	})
	if err != nil {
		panic(err)
	}
	err = cli.Start(ctx)
	if err != nil {
		panic(err)
	}

	return agt, cli
}

func TestMain(m *testing.M) {
	agent, spire = newTestInstances()
	retCode := m.Run()

	ctxCancel()
	<-agent.Wait()
	<-spire.Wait()
	<-priceStore.Wait()

	os.Exit(retCode)
}

func TestClient_PublishPrice(t *testing.T) {
	err := spire.Publish(testPriceAAABBB)
	assert.NoError(t, err)
}

func TestClient_PullPrice(t *testing.T) {
	var err error
	var price *messages.DataPoint

	err = spire.Publish(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		price, err = spire.PullPrice("AAA/BBB", testAddress.String())
		return price != nil
	}, time.Second)

	assert.NoError(t, err)
	assertEqualValue(t, testPriceAAABBB, price)
}

func TestClient_PullPrices_ByAssetPrice(t *testing.T) {
	var err error
	var prices []*messages.DataPoint

	err = spire.Publish(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		prices, err = spire.PullPrices("AAA/BBB", "")
		return len(prices) != 0
	}, time.Second)

	assert.NoError(t, err)
	assert.Len(t, prices, 1)
	assertEqualValue(t, testPriceAAABBB, prices[0])
}

func TestClient_PullPrices_ByFeed(t *testing.T) {
	var err error
	var prices []*messages.DataPoint

	err = spire.Publish(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		prices, err = spire.PullPrices("AAA/BBB", testAddress.String())
		return len(prices) != 0
	}, time.Second)

	assert.NoError(t, err)
	assert.Len(t, prices, 1)
	assertEqualValue(t, testPriceAAABBB, prices[0])
}

func TestClient_PullPricesErrorOnEmptyFilters(t *testing.T) {
	var err error
	var prices []*messages.DataPoint

	err = spire.Publish(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		prices, err = spire.PullPrices("", "")
		return len(prices) == 0
	}, time.Second)

	assert.Error(t, err)
	assert.Len(t, prices, 0)
}

func assertEqualValue(t *testing.T, expected, given *messages.DataPoint) {
	je, _ := json.Marshal(expected)
	jg, _ := json.Marshal(given)
	assert.JSONEq(t, string(je), string(jg))
}

func wait(cond func() bool, timeout time.Duration) {
	tn := time.Now()
	for {
		if cond() {
			break
		}
		if time.Since(tn) > timeout {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
