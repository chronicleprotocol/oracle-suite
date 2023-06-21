# Ghost CLI Readme

Ghost is an application run by feeds. It sends a signed price message to the Spire network.

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Supported events](#supported-events)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/ghost@latest`

Alternatively, you can build Ghost using `Makefile` directly from the repository. This approach is recommended if
you wish to work on Ghost source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Ghost, you have to create configuration file first. By default, the default config file location
is `config.hcl` in the current working directory. You can change the config file location using the `--config` flag.
Ghost supports HCL configuration format.

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
  feeds = [
    "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
    "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
  ]
}

ghost {
  # Ethereum key to use for signing price messages.
  ethereum_key = "default"

  # Specifies the interval in seconds between sending price messages.
  interval = 60

  # List of pairs to send price messages for.
  pairs = [
    "BTC/USD",
    "ETH/BTC",
    "ETH/USD",
  ]
}

# Ghost internally uses Gofer to fetch asset prices. The Gofer configuration is described in the Gofer README.
gofer {
  price_model "BTC/USD" "origin" {
    origin = "kraken"
  }

  price_model "ETH/BTC" "origin" {
    origin = "kraken"
  }

  price_model "ETH/USD" "origin" {
    origin = "kraken"
  }
}

ethereum {
  # Optional list of random Ethereum keys to use for signing. The name of the key is used to reference the key in other 
  # sections.
  rand_keys = ["key"]

  # Configuration for Ethereum keys. The key name is used to reference the key in other sections.
  # It is possible to have multiple keys in the configuration.
  key "default" {
    # Address of the Ethereum key. The address must be present in the keystore.
    address = "0x1234567890123456789012345678901234567890"

    # Path to the keystore directory.
    keystore_path = "./keystore"

    # Path to the file containing the passphrase for the keystore.
    # Optional.
    passphrase_file = "./passphrase"
  }

  # Configuration for Ethereum clients. The client name is used to reference the client in other sections.
  # It is possible to have multiple clients in the configuration.
  client "default" {
    # RPC URLs is a list of Ethereum RPC URLs to use for the client. Ethereum client uses RPC-Splitter which compares
    # responses from multiple RPC URLs to verify that none of them are compromised. At least three URLs are recommended
    # in case of using a 3rd party RPC service.
    rpc_urls = ["https://eth.public-rpc.com"]

    # Chain ID of the Ethereum network.
    chain_id = 1

    # Ethereum key to use for signing transactions.
    # Optional. If not specified, the default key is used, the signing is done by the Ethereum node.
    ethereum_key = "default"
  }
}

# Configuration for the transport layer. 
# Currently, libP2P and WebAPI transports are supported. At least one transport must be configured.
transport {
  # Configuration for the LibP2P transport. LibP2P transport uses peer-to-peer communication.
  # Optional.
  libp2p {
    # List of feed addresses. Only messages signed by these addresses are accepted.
    feeds = var.feeds

    # Seed used to generate the private key for the LibP2P node. 
    # Optional. If not specified, the private key is generated randomly.
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
    direct_peers_addrs = []

    # Addresses of peers to block. The addresses are encoded using multiaddr format.
    blocked_addrs = []

    # Disables node discovery. If disabled, the IP address of a node will not be broadcast to other peers. This option
    # should be used together with direct_peers_addrs.
    disable_discovery = false
  }

  # Configuration for the WebAPI transport. WebAPI transport allows to send messages using HTTP API. It is designed to 
  # use over secure network, e.g. Tor, I2P or VPN. WebAPI sends messages to other nodes using HTTP requests. The list of 
  # nodes is retrieved from the address book.
  # Optional.
  webapi {
    # List of feed addresses. Only messages signed by these addresses are accepted.
    feeds = var.feeds

    # Listen address for the WebAPI transport. The address must be in the format `host:port`.
    # If used with Tor, it is recommended to listen on 0.0.0.0 address.
    listen_addr = "0.0.0.0.8080"

    # Address of SOCKS5 proxy to use for the WebAPI transport. The address must be in the format `host:port`.
    # Optional.
    socks5_proxy_addr = "127.0.0.1:9050"

    # Ethereum key to sign messages that are sent to other nodes. The key must be present in the `ethereum` section.
    # Other nodes only accept messages that are signed by the key that is on the feeds list.
    ethereum_key = "default"

    # Ethereum address book that uses an Ethereum contract to fetch the list of node's addresses.
    # Optional.
    ethereum_address_book {
      # Ethereum contract address where the list of node's addresses is stored.
      contract_addr = "0x1234567890123456789012345678901234567890"

      # Ethereum client to use for fetching the list of node's addresses.
      ethereum_client = "default"
    }

    # Static address book that uses a static list of node's addresses.
    # Optional.
    static_address_book {
      addresses = ["0x1234567890123456789012345678901234567890", "0x1234567890123456789012345678901234567891"]
    }
  }
}
```

### Environment variables

It is possible to use environment variables anywhere in the configuration file. Environment variables are accessible
in the `env` object. For example, to use the `HOME` environment variable in the configuration file, use `env.HOME`.

## Commands

```
Usage:
  ghost [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         

Flags:
  -c, --config string                                  ghost config file (default "./config.hcl")
      --gofer.norpc                                    disable the use of Graph RPC agent
  -h, --help                                           help for ghost
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for ghost

Use "ghost [command] --help" for more information about a command.
```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
