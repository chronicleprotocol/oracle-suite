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
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chronicleprotocol/oracle-suite/pkg/price/store"

	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum"
	"github.com/chronicleprotocol/oracle-suite/pkg/ethereum/mocks"
	"github.com/chronicleprotocol/oracle-suite/pkg/log/null"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/oracle"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/local"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

var (
	testAddress     = ethereum.HexToAddress("0x2d800d93b065ce011af83f316cef9f0d005b0aa4")
	testPriceAAABBB = &messages.Price{
		Price: &oracle.Price{
			Wat: "AAABBB",
			Val: big.NewInt(10),
			Age: time.Unix(100, 0),
			V:   1,
		},
		Trace:   nil,
		Version: "0.4.10",
	}
	agent      *Agent
	spire      *Client
	priceStore *store.PriceStore
	ctxCancel  context.CancelFunc
)

func newTestInstances() (*Agent, *Client) {
	var err error
	var ctx context.Context
	ctx, ctxCancel = context.WithCancel(context.Background())

	log := null.New()
	sig := &mocks.Signer{}
	tra := local.New([]byte("test"), 0, map[string]transport.Message{
		messages.PriceV0MessageName: (*messages.Price)(nil),
		messages.PriceV1MessageName: (*messages.Price)(nil),
	})
	_ = tra.Start(ctx)
	priceStore, err = store.New(store.Config{
		Storage:   store.NewMemoryStorage(),
		Signer:    sig,
		Transport: tra,
		Pairs:     []string{"AAABBB", "XXXYYY"},
		Logger:    null.New(),
	})
	if err != nil {
		panic(err)
	}

	sig.On("Recover", mock.Anything, mock.Anything).Return(&testAddress, nil)

	agt, err := NewAgent(AgentConfig{
		PriceStore: priceStore,
		Transport:  tra,
		Signer:     sig,
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
		Signer:  sig,
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
	err := spire.PublishPrice(testPriceAAABBB)
	assert.NoError(t, err)
}

func TestClient_PullPrice(t *testing.T) {
	var err error
	var price *messages.Price

	err = spire.PublishPrice(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		price, err = spire.PullPrice("AAABBB", testAddress.String())
		return price != nil
	}, time.Second)

	assert.NoError(t, err)
	assertEqualPrices(t, testPriceAAABBB, price)
}

func TestClient_PullPrices_ByAssetPrice(t *testing.T) {
	var err error
	var prices []*messages.Price

	err = spire.PublishPrice(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		prices, err = spire.PullPrices("AAABBB", "")
		return len(prices) != 0
	}, time.Second)

	assert.NoError(t, err)
	assert.Len(t, prices, 1)
	assertEqualPrices(t, testPriceAAABBB, prices[0])
}

func TestClient_PullPrices_ByFeeder(t *testing.T) {
	var err error
	var prices []*messages.Price

	err = spire.PublishPrice(testPriceAAABBB)
	assert.NoError(t, err)

	wait(func() bool {
		prices, err = spire.PullPrices("", testAddress.String())
		return len(prices) != 0
	}, time.Second)

	assert.NoError(t, err)
	assert.Len(t, prices, 1)
	assertEqualPrices(t, testPriceAAABBB, prices[0])
}

func assertEqualPrices(t *testing.T, expected, given *messages.Price) {
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
