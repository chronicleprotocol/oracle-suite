package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Gofer_WSTETH(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI("http://127.0.0.1:8081")
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("wsteth").WithSymbol("WSTETH/ETH").WithPrice(1)).
		Add(origin.NewExchange("balancerV2").WithSymbol("STETH/ETH").
			WithCustom("match", "32296969ef14eb0c6d29669c550d4a044913023").
			WithCustom("response", "0000000000000000000000000000000000000000000000000000000000f1adbd00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000dde952b2b374c17")).
		Add(origin.NewExchange("curve").WithSymbol("STETH/ETH").
			WithCustom("match", "dc24316b9ae028f1497c275eb9192a3ea0f67022").
			WithCustom("response", "0000000000000000000000000000000000000000000000000000000000f1adba00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000ddde23599edb5d5")).
		Add(origin.NewExchange("ethrpc")).
		Deploy(*s)
	require.NoError(t, err)

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_chainId")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_chainId.json")).
			Mock(),
	}
	require.NoError(t, s.AddMocks(ctx, mocks))

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "--norpc", "price", "WSTETH/ETH")
	require.NoError(t, err)

	p, err := parseGoferPrice(out)
	require.NoError(t, err)

	assert.Equal(t, "aggregator", p.Type)
	assert.Equal(t, float64(1.0697381612529844), p.Price)
	assert.Greater(t, len(p.Prices), 0)
}
