package teleport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEthereum(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(getenv("SMOCKER_URL", "http://127.0.0.1:8081"))
	err := s.Reset(ctx)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	mocks := []*smocker.Mock{
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_blockNumber")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_blockNumber.json")).
			Mock(),
		smocker.NewMockBuilder().
			AddResponseHeader("Content-Type", "application/json").
			SetRequestBodyString(smocker.ShouldContainSubstring("eth_getLogs")).
			SetResponseBody(mustReadFile("./testdata/mock/eth_getLogs.json")).
			Mock(),
	}

	err = s.AddMocks(ctx, mocks)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	run(ctx, "../..", "./cmd/lair/...", "run", "-c", "./e2e/teleport/testdata/config/lair.json", "-v", "debug")
	waitForPort(ctx, "localhost", 30100)
	run(ctx, "../..", "./cmd/leeloo/...", "run", "-c", "./e2e/teleport/testdata/config/leeloo_ethereum.json", "-v", "debug")
	waitForPort(ctx, "localhost", 30101)
	run(ctx, "../..", "./cmd/leeloo/...", "run", "-c", "./e2e/teleport/testdata/config/leeloo2_ethereum.json", "-v", "debug")
	waitForPort(ctx, "localhost", 30102)

	time.Sleep(time.Second * 15)

	res, err := http.Get("http://localhost:30000/?type=teleport_evm&index=0x5f4a7c89123ed655b7fce471f2f14a4b699a9edfabeef6a8d5571976907f1884")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	lairResponse := LairResponse{}
	err = json.Unmarshal(body, &lairResponse)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Len(t, lairResponse, 2)
	sort.Slice(lairResponse, func(i, j int) bool {
		return lairResponse[i].Signatures["ethereum"].Signer < lairResponse[j].Signatures["ethereum"].Signer
	})

	assert.Equal(t,
		"52494e4b4542592d534c4156452d415242495452554d2d31000000000000000052494e4b4542592d4d41535445522d3100000000000000000000000000000000000000000000000000000000d747d98b8a2b28dfd6cd9f0e6015ad2a671611180000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000008180000000000000000000000000000000000000000000000000000000062b1e05f",
		lairResponse[0].Data["event"],
	)
	assert.Equal(t,
		"36223c6974790ab39f3b094fccbeb05b60592983206bddbb9c5fc9d9ede4706f",
		lairResponse[0].Data["hash"],
	)
	assert.Equal(t,
		"2d800d93b065ce011af83f316cef9f0d005b0aa4",
		lairResponse[0].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"e68c360c2c3eb0452369b8829611e2587896e1d990e3924cb6d18c178afda5735eeb99b424ba1f8230b3005f937705743885f8413c249514d8727514c3b324671c",
		lairResponse[0].Signatures["ethereum"].Signature,
	)

	assert.Equal(t,
		"52494e4b4542592d534c4156452d415242495452554d2d31000000000000000052494e4b4542592d4d41535445522d3100000000000000000000000000000000000000000000000000000000d747d98b8a2b28dfd6cd9f0e6015ad2a671611180000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000008180000000000000000000000000000000000000000000000000000000062b1e05f",
		lairResponse[1].Data["event"],
	)
	assert.Equal(t,
		"36223c6974790ab39f3b094fccbeb05b60592983206bddbb9c5fc9d9ede4706f",
		lairResponse[1].Data["hash"],
	)
	assert.Equal(t,
		"e3ced0f62f7eb2856d37bed128d2b195712d2644",
		lairResponse[1].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"36913257c92c309bcbf415a2a041ba1eeb02117c64e59aa73c54ddaee97126ec7b091cf83d65e912bd6d2dbb306a42e466a7080111cc797dd78b621df918b8aa1b",
		lairResponse[1].Signatures["ethereum"].Signature,
	)
}
