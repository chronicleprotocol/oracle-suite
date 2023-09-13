package e2e

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/defiweb/go-eth/types"

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

	const blockNumber int = 100
	stEthPerToken := 0.9

	WSTETH2ETH := stEthPerToken
	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("wsteth").
			WithSymbol("WSTETH/STETH").
			WithFunctionData("stEthPerToken", []origin.FunctionData{
				{
					Address: types.MustAddressFromHex("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"), // WSTETH/STETH
					Args:    []any{},
					Return:  []any{big.NewInt(int64(stEthPerToken * 1e18))},
				},
			}).
			WithCustom("blockNumber", blockNumber)).
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

	out, err := execCommand(ctx, "..", nil, "./gofer", "-c", "./e2e/testdata/config/gofer.hcl", "-v", "debug", "price", "WSTETH/STETH", "--format", "json")
	require.NoError(t, err)

	priceMap, err := parseGoferPrice(out)
	require.NoError(t, err)
	p := priceMap["WSTETH/STETH"]

	assert.Equal(t, "reference", p.Meta["type"])
	assert.Equal(t, WSTETH2ETH, p.Value.Price)
	assert.Greater(t, len(p.SubPoints), 0)
}
