package ghost

import (
	"bytes"
	"context"
	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	ethereumConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	spireConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
	transportConfig "github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
	"github.com/chronicleprotocol/oracle-suite/pkg/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

type spireClientConfig struct {
	Ethereum  ethereumConfig.Ethereum   `json:"ethereum"`
	Transport transportConfig.Transport `json:"transport"`
	Spire     spireConfig.Spire         `json:"spire"`
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := goBuild(ctx, "../..", "./cmd/ghost/...", "ghost"); err != nil {
		panic(err)
	}
	if err := goBuild(ctx, "../..", "./cmd/spire/...", "spire"); err != nil {
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

func buildTransport(ctx context.Context, configPath string) (*spire.Client, error) {
	cfg := &spireClientConfig{}
	err := config.ParseFile(cfg, configPath)
	if err != nil {
		return nil, err
	}
	signer, err := cfg.Ethereum.ConfigureSigner()
	if err != nil {
		return nil, err
	}

	cli, err := cfg.Spire.ConfigureClient(spireConfig.ClientDependencies{
		Signer: signer,
	})
	if err != nil {
		return nil, err
	}

	return cli, cli.Start(ctx)
}

func waitForSpire(ctx context.Context, client *spire.Client, assetPair, feeder string) (*messages.Price, error) {
	for ctx.Err() == nil {
		time.Sleep(time.Second)

		price, err := client.PullPrice(assetPair, feeder)
		if err == nil {
			return price, nil
		}
		if err.Error() == "reading body unexpected EOF" {
			continue
		}
		return nil, err
	}
	return nil, nil
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
