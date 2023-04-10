# Spire-Bootstrap CLI Readme

Spire-Bootstrap starts the LibP2P bootstrap node for the Spire network.

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/spire-bootstrap@latest`

Alternatively, you can build Spire-Bootstrap using `Makefile` directly from the repository. This approach is recommended
if you
wish to work on Spire-Bootstrap source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Spire-Bootstrap, you have to create configuration file first. By default, the default config file
location is `config.hcl` in the current working directory. You can change the config file location using the `--config`
flag. Spire-Bootstrap supports HCL configuration format.

### Configuration reference

_This configuration is only a reference and not ready for use. The recommended configuration can be found in
the `config.hcl` file located in the root directory._

```hcl
# List of files to include. The files are included in the order they are specified.
# It supports glob patterns.
include = [
  "config/*.hcl"
]

# Custom variables. Accessible in the configuration under the `var` object, e.g. `var.feeds`.
variables {
  myvar = "foo"
}

# Configuration for the transport layer.
transport {
  # Configuration for the LibP2P transport. LibP2P transport uses peer-to-peer communication.
  # Optional.
  libp2p {
    # Seed used to generate the private key for the LibP2P node. 
    # Optional. If not specified, the private key is generated randomly although for bootstrap nodes it must be
    # specified to ensure that the public key is always the same.
    priv_key_seed = "8c8eba62d853d3abdd7f3298341a622a8a9df37c3aba788028c646bdd915227c"

    # Listen addresses for the LibP2P node. The addresses are encoded using multiaddr format.
    listen_addrs = ["/ip4/0.0.0.0/tcp/8000"]

    # Addresses of bootstrap nodes. The addresses are encoded using multiaddr format.
    bootstrap_addrs = [
      "/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi",
      "/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR"
    ]

    # Addresses of direct peers to connect to. The addresses are encoded using multiaddr format.
    # This option must be configured symmetrically on both ends.
    # This option is not useful for bootstrap nodes.
    direct_peers_addrs = []

    # Addresses of peers to block. The addresses are encoded using multiaddr format.
    blocked_addrs = []

    # Should be set to false for bootstrap nodes.
    disable_discovery = false
  }
}
```

### Environment variables

It is possible to use environment variables anywhere in the configuration file. Environment variables are accessible
in the `env` object. For example, to use the `HOME` environment variable in the configuration file, use `env.HOME`.

## Commands

```
Usage:
  spire-bootstrap [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         Starts bootstrap node

Flags:
  -c, --config string                                  ghost config file (default "./config.hcl")
  -h, --help                                           help for spire-bootstrap
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for spire-bootstrap

Use "spire-bootstrap [command] --help" for more information about a command.

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
