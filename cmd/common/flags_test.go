package common

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
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

func TestConfigHcl_Contracts(t *testing.T) {
	var ce configEthereum
	var ct configTransport
	var cf configGofer
	var ch configGhost
	var csc configSpectre
	var csr configSpire
	tests := []struct {
		name    string
		config  config2.HasDefaults
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "ethereum",
			config:  &ce,
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name:    "transport",
			config:  &ct,
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name:    "gofer",
			config:  &cf,
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name:    "ghost",
			config:  &ch,
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name:    "spectre",
			config:  &csc,
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name:    "spire",
			config:  &csr,
			envVars: map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		os.Clearenv()
		for k, v := range tt.envVars {
			require.NoError(t, os.Setenv(k, v))
		}

		t.Run(tt.name, func(t *testing.T) {
			stdout := os.Stdout
			defer func() { os.Stdout = stdout }()

			r, w, _ := os.Pipe()
			os.Stdout = w

			var cf = ConfigFlagsForConfig(tt.config)
			require.NoError(t, cf.FlagSet().Parse([]string{"--config.hcl"}))
			argued, err := cf.Load(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.True(t, argued)

			_ = w.Close()
			out, _ := io.ReadAll(r)
			resp := strings.Trim(string(out), "\n")

			expected, err := os.ReadFile("./testdata/" + tt.name + ".txt")
			assert.Equal(t, string(expected), resp)
		})
	}
}
