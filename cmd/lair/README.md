# Lair CLI Readme

Lair is an application responsible for collecting signed events from the Spire P2P network, storing them, and providing
an HTTP API to retrieve them along with Oracle signatures.

Lair is one of the components of Maker Teleport: https://forum.makerdao.com/t/introducing-maker-teleport/11550

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [API](#api)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/lair@latest`

Alternatively, you can build Lair using `Makefile` directly from the repository. This approach is recommended if you
wish to work on Lair source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Lair, you have to create configuration file first. By default, the default config file location
is `config.hcl` in the current working directory. You can change the config file location using the `--config` flag.
Lair supports HCL configuration format.

### Configuration reference

```hcl
lair {
  # Listen address for the Lair server. The address must be in the format of "host:port".
  listen_addr = "0.0.0.0:8082"

  # In-memory storage configuration. 
  # Cannot be used together with storage_redis.
  storage_memory {
    # Specifies how long messages should be stored in seconds.
    # Optional. If not specified, the default value is 604800 (7 days).
    ttl = 604800
  }

  # Redis storage configuration.
  # Cannot be used together with storage_memory.
  storage_redis {
    # Specifies how long messages should be stored in seconds.
    # Optional. If not specified, the default value is 604800 (7 days).
    ttl = 604800

    # Redis server address. The address must be in the format of "host:port".
    addr = "127.0.0.1:6379"

    # Redis server username.
    # Optional.
    user = "user"

    # Redis server password.
    # Optional.
    pass = "pass"

    # Redis server database.
    # Optional. If not specified, the default value is 0.
    db = 0

    # Memory limit per feed in bytes.
    # Optional. If not specified, the default value is 0 (no limit).
    memory_limit = 0

    # Enable TLS.
    tls = true

    # TLS server name.
    tls_server_name = "redis-server"

    # TLS certificate file path.
    tls_cert_file = "/path/to/cert.pem"

    # TLS key file path.
    tls_key_file = "/path/to/key.pem"

    # TLS root CA file path.
    tls_root_ca_file = "/path/to/ca.pem"

    # Skip TLS certificate verification.
    tls_insecure_skip_verify = false

    # Enable Redis cluster mode.
    cluster = true

    # Redis cluster addrs. The addresses must be in the format of "host:port".
    cluster_addrs = ["198.51.100.0:6379", "203.0.113.0:6379"]
  }
}

# List of feed addresses. Only messages signed by these addresses are accepted.
feeds = [
  "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
  "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
]

# Configuration for the transport layer. 
# Currently, libP2P and WebAPI transports are supported. At least one transport must be configured.
transport {
  # Configuration for the LibP2P transport. LibP2P transport uses peer-to-peer communication.
  # Optional.
  libp2p {
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

## API

### Sample API response

```
Request:
GET http://127.0.0.1:8080/?type=teleport_evm&index=0x17b4079be1518b2df6e04f9206ac2e2a8822247760627f822aff87dfcad63150
```

```
Response:
Content-Type: application/json
```

```json
[
  {
    "timestamp": 1645275636,
    "data": {
      "event": "fe5b7488e442f5e8bdf7c9af40cc60dcaeda3f2704ebeddcf44f64e3e92a9c9b87de3d18c69fc10999eaf51d6536e28a069b008c8671417aa2695f6d725a8d72d97485d3c569202192525cbc677b366ef8fc2100000000000000000000000007ee98c5ec985fa2675fd4694c621ce2731ad42f500000000000000000000000000000000000000000000000000000000fbb8ecc280c973c158e88e38aa29849a0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000006210e9f4",
      "hash": "ce33e762dcfb265e7bf7c2d77f3a8d87520299557014613a2718e49efc18107f"
    },
    "signatures": {
      "ethereum": {
        "signer": "774d5aa0eee4897a9a6e65cbed845c13ffbc6d16",
        "signature": "7d9dce86f196c5d270653f54c41c6e1092e76c7088ddf0c60754d789a90308603cb3d59e37745a700b5caa1d570b4ecd7339d97734f7e365a01d96e3d65b49551b"
      }
    }
  },
  {
    "timestamp": 1645275636,
    "data": {
      "event": "2fe5b7488e442f5e8bdf7c9af40cc60dcaeda3f2704ebeddcf44f64e3e92a9c9b87de3d18c69fc10999eaf51d6536e28a069b008c8671417aa2695f6d725a8d72d97485d3c569202192525cbc677b366ef8fc2100000000000000000000000007ee98c5ec985fa2675fd4694c621ce2731ad42f500000000000000000000000000000000000000000000000000000000fbb8ecc280c973c158e88e38aa29849a0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000006210e9f4",
      "hash": "ce33e762dcfb265e7bf7c2d77f3a8d87520299557014613a2718e49efc18107f"
    },
    "signatures": {
      "ethereum": {
        "signer": "b41e8d40b7ac4eb34064e079c8eca9d7570eba1d",
        "signature": "ba4da22453ac98647fa5ff3dbce27ac2a9c85d5e88a92ca46ab590c8c54514ba3554a9e687a03ccbae0dac0fad15a8370c71857e294a934d109f16e00cf8ed291c"
      }
    }
  },
  {
    "timestamp": 1645275636,
    "data": {
      "event": "2fe5b7488e442f5e8bdf7c9af40cc60dcaeda3f2704ebeddcf44f64e3e92a9c9b87de3d18c69fc10999eaf51d6536e28a069b008c8671417aa2695f6d725a8d72d97485d3c569202192525cbc677b366ef8fc2100000000000000000000000007ee98c5ec985fa2675fd4694c621ce2731ad42f500000000000000000000000000000000000000000000000000000000fbb8ecc280c973c158e88e38aa29849a0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000006210e9f4",
      "hash": "ce33e762dcfb265e7bf7c2d77f3a8d87520299557014613a2718e49efc18107f"
    },
    "signatures": {
      "ethereum": {
        "signer": "23ce419dce1de6b3647ca2484a25f595132dfbd2",
        "signature": "1cf9005dbb8cbdb5afe5da5e13c6656e935ceb1c72c71a7f462321de08c8e8b41856939172b8ea1c3d9f0803a1b9b4d05fb70645a9f210dbad9e57749d42a6e71c"
      }
    }
  },
  {
    "timestamp": 1645275636,
    "data": {
      "event": "2fe5b7488e442f5e8bdf7c9af40cc60dcaeda3f2704ebeddcf44f64e3e92a9c9b87de3d18c69fc10999eaf51d6536e28a069b008c8671417aa2695f6d725a8d72d97485d3c569202192525cbc677b366ef8fc2100000000000000000000000007ee98c5ec985fa2675fd4694c621ce2731ad42f500000000000000000000000000000000000000000000000000000000fbb8ecc280c973c158e88e38aa29849a0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000006210e9f4",
      "hash": "ce33e762dcfb265e7bf7c2d77f3a8d87520299557014613a2718e49efc18107f"
    },
    "signatures": {
      "ethereum": {
        "signer": "c4756a9dae297a046556261fa3cd922dfc32db78",
        "signature": "fd61dd58b01118532dc011a8cc8014f5cd3c03b0e76c1aca043f3714d3593ad22b4e6f796823d1b6c57345e492dd476dde3c9b34daae51e432c0de38cc3950b01c"
      }
    }
  }
]
```

The requested URL has two parameters, the event type, and index. In the response above we got a list of all messages
(four in the example) for the given event type and index.

The fields in the response are:

- `[]` Array of events emitted during a given transaction.
    - `timestamp` - Date of the event.
    - `[string]data` - List of data associated with the event.
    - `[string]Signatures` - List of the Oracle signatures, where the key is the signature type.
        - `Signer` - Address of the Oracle.
        - `Signature` - Oracle signature.

## Commands

```
Usage:
  lair [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         Start the agent

Flags:
  -c, --config string                                  ghost config file (default "./config.hcl")
  -h, --help                                           help for lair
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for lair

Use "lair [command] --help" for more information about a command.

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
