package teleport

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/chronicleprotocol/infestor/smocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStarknet(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)

	s := smocker.NewAPI(env("SMOCKER_URL", "http://127.0.0.1:8081"))
	err := s.Reset(ctx)
	if err != nil {
		require.Fail(t, err.Error())
	}

	var mocks []*smocker.Mock
	mocks = append(mocks,
		smocker.NewMockBuilder().
			SetRequestPath(smocker.ShouldEqual("/feeder_gateway/get_block")).
			AddRequestQueryParam("blockNumber", smocker.ShouldEqual("null")).
			AddResponseHeader("Content-Type", "application/json").
			SetResponseBody(mustReadFile("./testdata/mock/starknet.json")).
			Mock(),
	)
	for n := 191500; n < 191504; n++ {
		mocks = append(mocks,
			smocker.NewMockBuilder().
				SetRequestPath(smocker.ShouldEqual("/feeder_gateway/get_block")).
				AddRequestQueryParam("blockNumber", smocker.ShouldEqual(strconv.Itoa(n))).
				AddResponseHeader("Content-Type", "application/json").
				SetResponseBody(mustReadFile("./testdata/mock/starknet.json")).
				Mock(),
		)
	}
	err = s.AddMocks(ctx, mocks)
	if err != nil {
		require.Fail(t, err.Error())
	}

	cmd1 := command(ctx, "../..", "./lair", "run", "-c", "./e2e/teleport/testdata/config/lair.json", "-v", "debug")
	cmd2 := command(ctx, "../..", "./leeloo", "run", "-c", "./e2e/teleport/testdata/config/leeloo_starknet.json", "-v", "debug")
	cmd3 := command(ctx, "../..", "./leeloo", "run", "-c", "./e2e/teleport/testdata/config/leeloo2_starknet.json", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = cmd1.Wait()
		_ = cmd2.Wait()
		_ = cmd3.Wait()
	}()

	if err := cmd1.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30100)
	if err := cmd2.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30101)
	if err := cmd3.Start(); err != nil {
		require.Fail(t, err.Error())
	}
	waitForPort(ctx, "localhost", 30102)

	lairResponse, err := waitForLair(ctx, "http://localhost:30000/?type=teleport_starknet&index=0x57a333bfccf30465cf287460c9c4bb7b21645213bc9cca7fbe99e1b9167d202", 2)
	if err != nil {
		require.Fail(t, err.Error())
	}

	require.Len(t, lairResponse, 2)

	assert.Equal(t,
		"474f45524c492d534c4156452d535441524b4e45542d31000000000000000000474f45524c492d4d41535445522d3100000000000000000000000000000000000000000000000000000000008aa7c51a6d380f4d9e273add4298d913416031ec0000000000000000000000008aa7c51a6d380f4d9e273add4298d913416031ec0000000000000000000000000000000000000000000000008ac7230489e80000000000000000000000000000000000000000000000000000000000000000000d0000000000000000000000000000000000000000000000000000000062822c1c",
		lairResponse[0].Data["event"],
	)
	assert.Equal(t,
		"3507a75b6cda5f180fa8e3ddf7bcb967699061a8f95549b73ecd2673dd14aa97",
		lairResponse[0].Data["hash"],
	)
	assert.Equal(t,
		"2d800d93b065ce011af83f316cef9f0d005b0aa4",
		lairResponse[0].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"ce94ba34ef4551559f44b3cdf53158cddb6c746fa935448af3f3e6027e217c9f66336e4e6dd9c8ee8408f11bbbd6b562d7f8c3213115abdaf67f731377686a6c1c",
		lairResponse[0].Signatures["ethereum"].Signature,
	)

	assert.Equal(t,
		"474f45524c492d534c4156452d535441524b4e45542d31000000000000000000474f45524c492d4d41535445522d3100000000000000000000000000000000000000000000000000000000008aa7c51a6d380f4d9e273add4298d913416031ec0000000000000000000000008aa7c51a6d380f4d9e273add4298d913416031ec0000000000000000000000000000000000000000000000008ac7230489e80000000000000000000000000000000000000000000000000000000000000000000d0000000000000000000000000000000000000000000000000000000062822c1c",
		lairResponse[1].Data["event"],
	)
	assert.Equal(t,
		"3507a75b6cda5f180fa8e3ddf7bcb967699061a8f95549b73ecd2673dd14aa97",
		lairResponse[1].Data["hash"],
	)
	assert.Equal(t,
		"e3ced0f62f7eb2856d37bed128d2b195712d2644",
		lairResponse[1].Signatures["ethereum"].Signer,
	)
	assert.Equal(t,
		"47dc14e7d26ebe431cd8d108de4d5fa9ef8f4db6d66174f8e828f9e2cd4504b13d42e9c351c9378fcc3c7522194fb268252cd17842b459f08682cd2eb77c52701c",
		lairResponse[1].Signatures["ethereum"].Signature,
	)
}
