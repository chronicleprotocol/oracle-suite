package common

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chronicleprotocol/oracle-suite/config"
	config2 "github.com/chronicleprotocol/oracle-suite/pkg/config"
	gofer "github.com/chronicleprotocol/oracle-suite/pkg/config/dataprovider"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/ethereum"
	feed "github.com/chronicleprotocol/oracle-suite/pkg/config/feednext"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/relay"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/spire"
	"github.com/chronicleprotocol/oracle-suite/pkg/config/transport"
)

type configEthereum struct {
	Ethereum ethereum.Config `hcl:"ethereum,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configEthereum) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Ethereum,
	}
}

type configTransport struct {
	Transport transport.Config `hcl:"transport,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configTransport) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Ethereum,
		config.Transport,
	}
}

type configGofer struct {
	Gofer gofer.Config `hcl:"gofer,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configGofer) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Gofer,
	}
}

type configGhost struct {
	Ghost feed.Config `hcl:"ghost,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configGhost) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Ghost,
	}
}

type configSpectre struct {
	Spectre relay.Config `hcl:"spectre,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configSpectre) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Ethereum,
		config.Spectre,
	}
}

type configSpire struct {
	Spire spire.ConfigSpire `hcl:"spire,block"`

	// HCL fields:
	Remain  hcl.Body        `hcl:",remain"` // To ignore unknown blocks.
	Content hcl.BodyContent `hcl:",content"`
}

func (configSpire) DefaultEmbeds() [][]byte {
	return [][]byte{
		config.Defaults,
		config.Contracts,
		config.Spire,
	}
}

// Mapping to seek expected string for given test case.
// Key of map has the following format: [config type]:[ENV]=[VAL],[ENV2]=[VAL2] ...
// i.e. spectre:CFG_ENVIRONMENT=prod;CFG_CHAIN_NAME=eth
// Value of map is the expected string for given key.
var completeExpectedHCLs = map[string]string{}

func init() {
	files := []string{
		"./testdata/prod.txt",
		"./testdata/stage.txt",
		"./testdata/prod-arb1.txt",
		"./testdata/prod-eth.txt",
		"./testdata/prod-gno.txt",
		"./testdata/prod-oeth.txt",
		"./testdata/prod-zkevm.txt",
		"./testdata/stage-arb-goerli.txt",
		"./testdata/stage-gno.txt",
		"./testdata/stage-gor.txt",
		"./testdata/stage-mango.txt",
		"./testdata/stage-ogor.txt",
		"./testdata/stage-sep.txt",
		"./testdata/stage-zkevm.txt",
	}
	for _, file := range files {
		readFile, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(readFile)
		scanner.Split(bufio.ScanLines)
		testCase := ""
		expected := strings.Builder{}
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "TestCase for ") {
				if testCase != "" {
					completeExpectedHCLs[testCase] = strings.Trim(expected.String(), "\n")
					expected.Reset()
				}
				// i.e.: spectre:CFG_ENVIRONMENT=prod;CFG_CHAIN_NAME=eth
				testCase = line[13:]
				expected.Reset()
				continue
			}
			expected.WriteString(line)
			expected.WriteString("\n")
		}
		completeExpectedHCLs[testCase] = strings.Trim(expected.String(), "\n")
		expected.Reset()
		readFile.Close()
	}
}

func TestConfigHCL_Env_Chain(t *testing.T) {
	types := map[string]reflect.Type{
		"ethereum":  reflect.TypeOf((*configEthereum)(nil)).Elem(),
		"transport": reflect.TypeOf((*configTransport)(nil)).Elem(),
		"gofer":     reflect.TypeOf((*configGofer)(nil)).Elem(),
		"ghost":     reflect.TypeOf((*configGhost)(nil)).Elem(),
		"spectre":   reflect.TypeOf((*configSpectre)(nil)).Elem(),
		"spire":     reflect.TypeOf((*configSpire)(nil)).Elem(),
	}

	for testCase, expected := range completeExpectedHCLs {
		// testCase format: [config type]:ENV=VAL,ENV2=VAL2...
		// spectre-prod-oeth:CFG_ENVIRONMENT=prod;CFG_CHAIN_NAME=eth
		tokens := strings.Split(testCase, ":")
		if len(tokens) < 2 {
			continue
		}
		confType := tokens[0]
		envStrings := strings.Join(tokens[1:], ":")

		os.Clearenv()
		for _, envStr := range strings.Split(envStrings, ";") {
			if !strings.HasPrefix(envStr, "CFG_") {
				continue
			}
			// envStr: CFG_ENVIRONMENT=prod
			envs := strings.Split(envStr, "=")
			if len(envs) != 2 {
				continue
			}
			require.NoError(t, os.Setenv(envs[0], envs[1]))
		}

		refType, ok := types[confType]
		if !ok {
			continue
		}

		t.Run(testCase, func(t *testing.T) {
			stdout := os.Stdout
			defer func() { os.Stdout = stdout }()

			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create new Config instance per each test case
			configRef := reflect.New(refType)
			configInst := configRef.Interface().(config2.HasDefaults)

			var cf = ConfigFlagsForConfig(configInst)
			require.NoError(t, cf.FlagSet().Parse([]string{"--config.hcl"}))
			argued, err := cf.Load(&configInst)
			require.NoError(t, err)
			require.True(t, argued)

			_ = w.Close()
			out, _ := io.ReadAll(r)
			resp := strings.Trim(string(out), "\n")

			assert.Equal(t, expected, resp)

			// Retry to load exported config again and check data integrity
			r, w, _ = os.Pipe()
			os.Stdout = w
			alterRef := reflect.New(refType)
			alterInst := alterRef.Interface().(config2.HasDefaults)
			alterCf := ConfigFlagsWithEmbeds(out)
			require.NoError(t, alterCf.FlagSet().Parse([]string{"--config.hcl"}))
			alterArgued, err := alterCf.Load(&alterInst)
			require.NoError(t, err)
			require.True(t, alterArgued)
			_ = w.Close()
			out, _ = io.ReadAll(r)
			alterResp := strings.Trim(string(out), "\n")

			assert.Equal(t, resp, alterResp)
		})
	}
}
