package gofere2e

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
)

func xTestMain(m *testing.M) {
	ctx := context.Background()
	if err := goBuild(ctx, "../..", "./cmd/gofer/...", "gofer"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GOFER_BINARY_PATH", "../../gofer"); err != nil {
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
