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

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/defiweb/go-eth/abi"
	"github.com/defiweb/go-eth/hexutil"
	"github.com/defiweb/go-eth/types"
	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/infestor"
	"github.com/chronicleprotocol/infestor/origin"
	"github.com/chronicleprotocol/infestor/smocker"
)

func isProcessRunning(processName string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		if process.Executable() == processName {
			return os.FindProcess(process.Pid())
		}
	}
	return nil, nil
}

func waitForAppRun(ctx context.Context, appName string) error {
	for ctx.Err() == nil {
		process, err := isProcessRunning(appName)
		if err != nil {
			return err
		}
		if process != nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return nil
}

const rpcJSONResult = `{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "%s"
}`
const smockerAPI = "http://127.0.0.1:8081"
const smockerURL = "http://127.0.0.1:8080"
const configRegistry = "0x2712B667D07c7376F2C31642b2D578FB6D5F5364"

func mockConfigRegistry(configRegistry string, ipfs string, hclFile string) ([]*smocker.Mock, error) {
	data :=
		"0x" +
			"0000000000000000000000000000000000000000000000000000000000000020" +
			"00000000000000000000000000000000000000000000000000000000000000" + fmt.Sprintf("%x", len(ipfs)) +
			hexutil.BytesToHex(types.Bytes(ipfs).PadRight(32))[2:]

	latest := abi.MustParseMethod("latest()(string memory)")
	args, _ := latest.EncodeArgs()

	m := smocker.ShouldContainSubstring(hexutil.BytesToHex(args))
	mockRPC := &smocker.Mock{
		Request: smocker.MockRequest{
			Method: smocker.ShouldEqual("POST"),
			Path:   smocker.ShouldEqual("/"),
			Body: &smocker.BodyMatcher{
				BodyString: &m,
			},
		},
		Response: &smocker.MockResponse{
			Status: http.StatusOK,
			Headers: map[string]smocker.StringSlice{
				"Content-Type": []string{
					"application/json",
				},
			},
			Body: fmt.Sprintf(rpcJSONResult, data),
		},
	}

	hcl, _ := os.ReadFile(hclFile)
	mockHttp := &smocker.Mock{
		Request: smocker.MockRequest{
			Method: smocker.ShouldEqual("GET"),
			Path:   smocker.ShouldEqual(strings.ReplaceAll(ipfs, smockerURL, "")),
			//Body: &smocker.BodyMatcher{
			//	BodyString: &m,
			//},
		},
		Response: &smocker.MockResponse{
			Status: http.StatusOK,
			Headers: map[string]smocker.StringSlice{
				"Content-Type": []string{
					"application/text",
				},
			},
			Body: string(hcl),
		},
	}

	mocks := []*smocker.Mock{
		mockRPC,
		mockHttp,
	}
	return mocks, nil
}

func initializeMock(t *testing.T, ctx context.Context, s *smocker.API, ipfs string, hclFile string, btcUsdPrice, ethBtcPrice, ethUsdPrice float64) {
	require.NoError(t, s.Reset(ctx))

	err := infestor.NewMocksBuilder().
		Add(origin.NewExchange("kraken").WithSymbol("BTC/USD").WithPrice(btcUsdPrice)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/BTC").WithPrice(ethBtcPrice)).
		Add(origin.NewExchange("kraken").WithSymbol("ETH/USD").WithPrice(ethUsdPrice)).
		Deploy(*s)
	require.NoError(t, err)

	mocks, err := mockConfigRegistry(configRegistry, ipfs, hclFile)
	require.NoError(t, err)
	require.NoError(t, s.AddMocks(ctx, mocks))
}

func Test_Morph_Run_Ghost(t *testing.T) {
	// Scenario:
	// 1. run morph, then morph will execute ghost with `run` command
	// 2. run spire without morph
	// 3. wait for 5 seconds
	// 4. spire should be able to pull price from ghost
	// 5. cancel the context and wait for morph, spire to quit
	// 6. check if morph and spire still exist

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(smockerAPI)
	initializeMock(t, ctx, s, smockerURL+"/ipfs", "../e2e/testdata/config/ghost.hcl", 1, 1, 1)

	morphCmd := command(ctx, "..", nil, "./morph", "run", "--bin", "./ghost", "--args", "run", "--config-rpc", smockerURL, "--config-registry", configRegistry, "-v", "debug")
	spireCmd := command(ctx, "..", nil, "./spire", "agent", "-c", "./e2e/testdata/config/spire.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = morphCmd.Wait()
		_ = spireCmd.Wait()

		p, err := isProcessRunning("ghost")
		assert.NoError(t, err)
		assert.Nil(t, p)
	}()

	// 2. run spire without morph
	require.NoError(t, spireCmd.Start())
	waitForPort(ctx, "localhost", 30100)

	// 1. run morph, then morph will execute ghost with `run` command
	require.NoError(t, morphCmd.Start())
	time.Sleep(5 * time.Second)

	// wait for morph to run
	require.NoError(t, waitForAppRun(ctx, "morph"))
	waitForPort(ctx, "localhost", 30101)

	// 3. wait for 5 seconds
	time.Sleep(5 * time.Second)

	// 4. spire should be able to pull price from ghost
	btcusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTC/USD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	btcusdPrice, err := parseSpireDataPointMessage(btcusdMessage)
	require.NoError(t, err)

	assert.Equal(t, "1", btcusdPrice.Point.Value.Price)
	assert.InDelta(t, time.Now().Unix(), btcusdPrice.Point.Time.Unix(), 10)
	assert.Equal(t, "BTC/USD", btcusdPrice.Point.Value.Pair)

	// force quit morph, so ghost as child process will be exited smoothly
	err = morphCmd.Process.Signal(syscall.SIGINT)
	assert.NoError(t, err)
	err = morphCmd.Wait()
	assert.NoError(t, err)
}

func Test_Morph_SelfExit(t *testing.T) {
	// Scenario:
	// 1. run morph with wrong command for ghost
	// 2. ghost will exit itself due to wrong command and morph should exit itself after ghost
	// 3. wait for 5 seconds
	// 4. check if morph and ghost still exist, should not exist
	// 5. cancel the context

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(smockerAPI)
	initializeMock(t, ctx, s, smockerURL+"/ipfs", "../e2e/testdata/config/ghost.hcl", 1, 1, 1)

	wrongArg := "wrong"
	morphCmd := command(ctx, "..", []string{
		"CFG_MORPH_INTERVAL=5", // NOTE: set morph interval with 5 seconds
	}, "./morph", "run", "--bin", "./ghost", "--args", wrongArg, "--config-rpc", smockerURL, "--config-registry", configRegistry, "-v", "debug")
	defer func() {
		ctxCancel()
		_ = morphCmd.Wait()
	}()

	// 1. run morph with wrong command for ghost
	require.NoError(t, morphCmd.Start())

	// 2. ghost will exit itself due to wrong command and morph should exit itself after ghost

	// 3. wait for 5 seconds
	time.Sleep(5 * time.Second)

	// 4. check if morph and ghost still exist, should not exist
	p, err := isProcessRunning("ghost")
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func Test_Morph_Restart_Ghost(t *testing.T) {
	// Scenario:
	// 1. set few seconds of interval and run morph, then ghost will be executed.
	// 2. pull price via spire
	// 3. mock rpc with different config (different data model)
	// 4. wait for few seconds
	// 5. morph should detect that config has been changed and ghost should be restarted
	// 6. pull price of another data model via spire, should get value
	// 7. cancel the context, quit morph and spire

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(smockerAPI)
	initializeMock(t, ctx, s, smockerURL+"/ipfs", "../e2e/testdata/config/ghost.hcl", 1, 1, 1)

	morphCmd := command(ctx, "..", []string{
		"CFG_MORPH_INTERVAL=5",
	}, "./morph", "run", "--bin", "./ghost", "--args", "run", "--config-rpc", smockerURL, "--config-registry", configRegistry, "-v", "debug")
	spireCmd := command(ctx, "..", nil, "./spire", "agent", "-c", "./e2e/testdata/config/spire.hcl", "-v", "debug")
	defer func() {
		ctxCancel()
		_ = morphCmd.Wait()
		_ = spireCmd.Wait()

		p, err := isProcessRunning("ghost")
		assert.NoError(t, err)
		assert.Nil(t, p)
	}()

	// 2. pull price via spire
	require.NoError(t, spireCmd.Start())
	waitForPort(ctx, "localhost", 30100)

	// 1. set few seconds of interval and run morph, then ghost will be executed.
	require.NoError(t, morphCmd.Start())
	time.Sleep(5 * time.Second)
	require.NoError(t, waitForAppRun(ctx, "morph"))
	waitForPort(ctx, "localhost", 30101)

	time.Sleep(5 * time.Second)
	btcusdMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTC/USD", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	btcusdPrice, err := parseSpireDataPointMessage(btcusdMessage)
	require.NoError(t, err)

	assert.Equal(t, "1", btcusdPrice.Point.Value.Price)
	assert.InDelta(t, time.Now().Unix(), btcusdPrice.Point.Time.Unix(), 10)
	assert.Equal(t, "BTC/USD", btcusdPrice.Point.Value.Pair)

	// 3. mock rpc with different config (different data model)
	initializeMock(t, ctx, s, smockerURL+"/ipfs2", "../e2e/testdata/config/ghost2.hcl", 1, 0.5, 1)

	// 4. wait for few seconds
	time.Sleep(10 * time.Second)

	// 5. morph should detect that config has been changed and ghost should be restarted

	// 6. pull price of another data model via spire, should get value
	btcethMessage, err := execCommand(ctx, "..", nil, "./spire", "-c", "./e2e/testdata/config/spire.hcl", "pull", "price", "BTC/ETH", "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4")
	require.NoError(t, err)

	btcethPrice, err := parseSpireDataPointMessage(btcethMessage)
	require.NoError(t, err)

	assert.Equal(t, "2", btcethPrice.Point.Value.Price)
	//assert.InDelta(t, time.Now().Unix(), btcethPrice.Point.Time.Unix(), 10)
	assert.Equal(t, "BTC/ETH", btcethPrice.Point.Value.Pair)

	// force quit morph, so ghost as child process will be exited smoothly
	err = morphCmd.Process.Signal(syscall.SIGINT)
	assert.NoError(t, err)
	err = morphCmd.Wait()
	assert.NoError(t, err)
}

func Test_Morph_ForceExit(t *testing.T) {
	// Scenario:
	// 1. run morph, then morph will execute ghost with `run` command
	// 2. send interrupt signal to morph
	// 3. wait for morph to quit and then check if ghost still exists, should not exist

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(smockerAPI)
	initializeMock(t, ctx, s, smockerURL+"/ipfs", "../e2e/testdata/config/ghost.hcl", 1, 1, 1)

	morphCmd := command(ctx, "..", nil, "./morph", "run", "--bin", "./ghost", "--args", "run", "--config-rpc", smockerURL, "--config-registry", configRegistry, "-v", "debug")
	defer func() {
		ctxCancel()
		p, err := isProcessRunning("ghost")
		assert.NoError(t, err)
		assert.Nil(t, p)
	}()

	// 1. run morph, then morph will execute ghost with `run` command
	require.NoError(t, morphCmd.Start())
	time.Sleep(5 * time.Second)
	require.NoError(t, waitForAppRun(ctx, "morph"))

	// 2. send interrupt signal to morph
	p, err := isProcessRunning("morph")
	assert.NoError(t, err)
	assert.NotNil(t, p)

	err = p.Signal(syscall.SIGINT)
	assert.NoError(t, err)

	err = morphCmd.Wait()
	assert.NoError(t, err)
}

func Test_Morph_Running_Ghost(t *testing.T) {
	// Scenario:
	// 1. run ghost with `run` command
	// 2. run morph after that, morph should detect that ghost is running
	// 3. wait 5 seconds
	// 4. send interrupt signal to morph
	// 5. wait for morph to quit and then check if ghost still exists, should not exist

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer ctxCancel()

	s := smocker.NewAPI(smockerAPI)
	initializeMock(t, ctx, s, smockerURL+"/ipfs", "../e2e/testdata/config/ghost.hcl", 1, 1, 1)

	ghostCmd := command(ctx, "..", nil, "./ghost", "run", "-c", "./e2e/testdata/config/ghost.hcl", "-v", "debug")
	morphCmd := command(ctx, "..", []string{
		"CFG_MORPH_INTERVAL=5",
	}, "./morph", "run", "--bin", "./ghost", "--args", "run", "--config-rpc", smockerURL, "--config-registry", configRegistry, "-v", "debug")
	defer func() {
		ctxCancel()
		p, err := isProcessRunning("ghost")
		assert.NoError(t, err)
		assert.Nil(t, p)
	}()

	// 1. run ghost with `run` command
	require.NoError(t, ghostCmd.Start())
	waitForPort(ctx, "localhost", 30101)

	// 2. run morph after that, morph should detect that ghost is running
	require.NoError(t, morphCmd.Start())
	time.Sleep(5 * time.Second)
	require.NoError(t, waitForAppRun(ctx, "morph"))

	// 3. ghost should be interrupted
	err := ghostCmd.Wait()
	assert.NoError(t, err)

	// 4. send interrupt signal to morph
	p, err := isProcessRunning("morph")
	assert.NoError(t, err)
	assert.NotNil(t, p)

	err = p.Signal(syscall.SIGINT)
	assert.NoError(t, err)

	_ = morphCmd.Wait()
}
