package teleport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
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

func debugRun(ctx context.Context, wd, path string, params ...string) {
	call(ctx, wd, "dlv", append([]string{"--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "debug", path}, params...)...)
}

func run(ctx context.Context, wd, path string, params ...string) {
	call(ctx, wd, "go", append([]string{"run", path}, params...)...)
}

func call(ctx context.Context, wd, bin string, params ...string) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.CommandContext(ctx, bin, params...)
	cmd.Dir = wd
	cmd.Env = os.Environ()

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
}

func getenv(env string, def string) string {
	v := os.Getenv(env)
	if len(v) == 0 {
		return def
	}
	return v
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func mustReadFile(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}
