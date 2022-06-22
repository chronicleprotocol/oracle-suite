package teleport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
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

func getenv(env string, def string) string {
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
