package e2e

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

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := goBuild(ctx, "..", "./cmd/gofer/...", "gofer"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "..", "./cmd/ghost/...", "ghost"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "..", "./cmd/spire/...", "spire"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "..", "./cmd/lair/...", "lair"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "..", "./cmd/leeloo/...", "leeloo"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func goBuild(ctx context.Context, wd, path, out string) error {
	cmd := command(ctx, wd, nil, "go", "build", "-gcflags", "all=-N -l", "-o", out, path)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func command(ctx context.Context, wd string, envs []string, bin string, params ...string) *exec.Cmd {
	env := os.Environ()
	for _, e := range envs {
		env = append(env, e)
	}
	cmd := exec.CommandContext(ctx, bin, params...)
	cmd.Dir = wd
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func execCommand(ctx context.Context, wd string, envs []string, bin string, params ...string) ([]byte, error) {
	var buf bytes.Buffer
	env := os.Environ()
	for _, e := range envs {
		env = append(env, e)
	}
	cmd := exec.CommandContext(ctx, bin, params...)
	cmd.Dir = wd
	cmd.Env = env
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return buf.Bytes(), err
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
