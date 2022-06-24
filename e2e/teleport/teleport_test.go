package teleport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"testing"
	"time"
)

type LairResponse []struct {
	Timestamp  int                              `json:"timestamp"`
	Data       map[string]string                `json:"data"`
	Signatures map[string]LairResponseSignature `json:"signatures"`
}

type LairResponseSignature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := goBuild(ctx, "../..", "./cmd/lair/...", "lair"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "../..", "./cmd/leeloo/...", "leeloo"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func goBuild(ctx context.Context, wd, path, out string) error {
	cmd := command(ctx, wd, "go", "build", "-o", out, path)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func command(ctx context.Context, wd, bin string, params ...string) *exec.Cmd {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.CommandContext(ctx, bin, params...)
	cmd.Dir = wd
	cmd.Env = os.Environ()
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	return cmd
}

func env(env string, def string) string {
	v := os.Getenv(env)
	if len(v) == 0 {
		return def
	}
	return v
}

func mustReadFile(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func waitForLair(ctx context.Context, url string, responses int) (LairResponse, error) {
	lairResponse := LairResponse{}
	for ctx.Err() == nil {
		time.Sleep(time.Second)
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			continue
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &lairResponse)
		if err != nil {
			return nil, err
		}
		if len(lairResponse) == responses {
			break
		}
	}
	sort.Slice(lairResponse, func(i, j int) bool {
		return lairResponse[i].Signatures["ethereum"].Signer < lairResponse[j].Signatures["ethereum"].Signer
	})
	return lairResponse, nil
}

func waitForPort(ctx context.Context, host string, port int) {
	for ctx.Err() == nil {
		if isPortOpen(host, port) {
			return
		}
		time.Sleep(time.Second)
	}
}

func isPortOpen(host string, port int) bool {
	c, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	c.Close()
	return true
}
