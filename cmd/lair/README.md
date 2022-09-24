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
command: `go get -u github.com/chronicleprotocol/oracle-suite/cmd/lair`.

Alternatively, you can build Lair using `Makefile` directly from the repository. This approach is recommended if you
wish to work on Lair source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Lair, you have to create configuration file first. By default, the default config file location
is `config.json` in the current working directory. You can change the config file location using the `--config` flag.
Lair supports JSON and YAML configuration files.

```bash

### Example configuration

```json
{
  "transport": {
    "transport": "libp2p",
    "libp2p": {
      "privKeySeed": "02082cf471002b5c5dfefdd6cbd30666ff02c4df90169f766877caec26ed4f88",
      "listenAddrs": [
        "/ip4/0.0.0.0/tcp/8000"
      ],
      "bootstrapAddrs": [
        "/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi",
        "/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR"
      ],
      "directPeersAddrs": [],
      "blockedAddrs": [],
      "disableDiscovery": false
    }
  },
  "feeds": [
    "0xDA1d2961Da837891f43235FddF66BAD26f41368b",
    "0x4b0E327C08e23dD08cb87Ec994915a5375619aa2",
    "0x75ef8432566A79C86BBF207A47df3963B8Cf0753",
    "0x83e23C207a67a9f9cB680ce84869B91473403e7d",
    "0xFbaF3a7eB4Ec2962bd1847687E56aAEE855F5D00",
    "0xfeEd00AA3F0845AFE52Df9ECFE372549B74C69D2",
    "0x71eCFF5261bAA115dcB1D9335c88678324b8A987",
    "0x8ff6a38A1CD6a42cAac45F08eB0c802253f68dfD",
    "0x16655369Eb59F3e1cAFBCfAC6D3Dd4001328f747",
    "0xD09506dAC64aaA718b45346a032F934602e29cca",
    "0xc00584B271F378A0169dd9e5b165c0945B4fE498",
    "0x60da93D9903cb7d3eD450D4F81D402f7C4F71dd9",
    "0xa580BBCB1Cee2BCec4De2Ea870D20a12A964819e",
    "0xD27Fa2361bC2CfB9A591fb289244C538E190684B",
    "0x8de9c5F1AC1D4d02bbfC25fD178f5DAA4D5B26dC",
    "0xE6367a7Da2b20ecB94A25Ef06F3b551baB2682e6",
    "0xA8EB82456ed9bAE55841529888cDE9152468635A",
    "0x130431b4560Cd1d74A990AE86C337a33171FF3c6",
    "0x8aFBD9c3D794eD8DF903b3468f4c4Ea85be953FB",
    "0xd94BBe83b4a68940839cD151478852d16B3eF891",
    "0xC9508E9E3Ccf319F5333A5B8c825418ABeC688BA",
    "0x77EB6CF8d732fe4D92c427fCdd83142DB3B742f7",
    "0x3CB645a8f10Fb7B0721eaBaE958F77a878441Cb9",
    "0x4f95d9B4D842B2E2B1d1AC3f2Cf548B93Fd77c67",
    "0xaC8519b3495d8A3E3E44c041521cF7aC3f8F63B3",
    "0xd72BA9402E9f3Ff01959D6c841DDD13615FFff42"
  ],
  "lair": {
    "listenAddr": "127.0.0.1:8082",
    "storage": {
      "type": "redis",
      "redis": {
        "address": "127.0.0.1:7001",
        "password": "password",
        "db": 0,
        "tls": true 
      }
    }
  }
}
```

### Configuration reference

- `transport` - Configuration parameters for transports mechanisms used to relay messages.
    - `transport` (string) - Transport to use. Supported mechanism are: `libp2p` and `ssb`. If empty, the `libp2p` is
      used.
    - `libp2p` - Configuration parameters for the libp2p transport (Spire network).
        - `privKeySeed` (`string`) - The random hex-encoded 32 bytes. It is used to generate a unique identity on the
          libp2p network. The value may be empty to generate a random seed.
        - `listenAddrs` (`[]string`) - List of listening addresses for libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `bootstrapAddrs` (`[]string`) - List of addresses of bootstrap nodes for the libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `directPeersAddrs` (`[]string`) - List of direct peer addresses to which messages will be sent directly.
          Addresses are encoded using the format. [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
          This option must be configured symmetrically on both ends.
        - `blockedAddrs` (`[]string`) - List of blocked peers or IP addresses encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `disableDiscovery` (`bool`) - Disables node discovery. If enabled, the IP address of a node will not be
          broadcast to other peers. This option must be used together with `directPeersAddrs`.
- `feeds` (`[]string`) - List of hex-encoded addresses of other Oracles. Event messages from Oracles outside that list
  will be ignored.
- `logger` - Optional logger configuration.
    - `grafana` - Configuration of Grafana logger. Grafana logger can extract values from log messages and send them to
      Grafana Cloud.
        - `enable` (`string`) - Enable Grafana metrics.
        - `interval` (`int`) - Specifies how often, in seconds, logs should be sent to the Grafana Cloud server. Logs
          with the same name in that interval will be replaced with never ones.
        - `endpoint` (`string`) - Graphite server endpoint.
        - `apiKey` (`string`) - Graphite API key.
        - `[]metrics` - List of metric definitions
            - `matchMessage` (`string`) - Regular expression that must match a log message.
            - `matchFields` (`[string]string`) - Map of fields whose values must match a regular expression.
            - `name` (`string`) - Name of metric. It can contain references to log fields in the format `$${path}`,
              where path is the dot-separated path to the field.
            - `tags` (`[string][]string`) - List of metric tags. They can contain references to log fields in the
              format `${path}`, where path is the dot-separated path to the field.
            - `value` (`string`) - Dot-separated path of the field with the metric value. If empty, the value 1 will be
              used as the metric value.
            - `scaleFactor` (`float`) - Scales the value by the specified number. If it is zero, scaling is not
              applied (default: 0).
            - `onDuplicate` (`string`) - Specifies how duplicated values in the same interval should be handled. Allowed
              options are:
                - `sum` - Add values.
                - `sub` - Subtract values.
                - `max` - Use higher value.
                - `min` - Use lower value.
                - `replace` (default) - Replace the value with a newer one.
- `lair` - Lair configuration.
    - `value` (`string`) - Dot-separated path of the field with the metric value. If empty, the value 1 will be used as
      the metric value.
        - `listenAddr` (`string`) - Listen address for the HTTP server provided as the combination of IP address and
          port number.
        - `storage` - Configure the data storage mechanism used by Lair.
            - `type` (`string`) - Type of the storage mechanism. Supported mechanism are: `redis` and `memory` (
              default: `memory`).
            - `redis` - Configuration for the Redis storage mechanism. Ignored if `type` is not `redis`.
                - `ttl` (`integer`) - Specifies how long messages should be stored in seconds. (default: 604800 seconds
                  about one week)
                - `address` (`string`) - Redis server address provided as the combination of IP address or host and port
                  number, e.g. `0.0.0.0:8080`.
                - `username` (`string`) - Redis server username for ACL.
                - `password` (`string`) - Redis server password.
                - `db` (`int`) - Redis server database number.
                - `memoryLimit` (`int`) - Memory limit per Oracle in bytes. If 0 or not specified, no limit is applied.
                - `tls` (`bool`) - Enables TLS connection to Redis server. (default: `false`)
                - `tlsServerName` (`string`) - Server name used to verify the hostname on the returned certificates from
                  the server. Ignored if empty. (default: `""`)
                - `tlsCertFile` (`string`) - Path to the PEM encoded certificate file. (default: `""`)
                - `tlsKeyFile` (`string`) - Path to the PEM encoded private key file. (default: `""`)
                - `tlsRootCAFile` (`string`) - Path to the PEM encoded root certificate file. (default: `""`)
                - `tlsInsecureSkipVerify` (`bool`) - Disables TLS certificate verification. (default: `false`)
            - `memory` - Configuration the memory storage mechanism. Ignored if `type` is not `memory`.
                - `ttl` (`int`) - Specifies how long messages should be stored in seconds. (default: 604800 seconds -
                  about one week)

### Environment variables

It is possible to use environment variables anywhere in the configuration file. The syntax is similar as in the
shell: `${ENV_VAR}`. If the environment variable is not set, the error will be returned during the application
startup. To escape the dollar sign, use `\$` or `$$`. The latter syntax is not supported inside variables. It is
possible to define default values for environment variables. To do so, use the following syntax: `${ENV_VAR-default}`.

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
  -c, --config string                                  ghost config file (default "./config.json")
  -h, --help                                           help for lair
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for lair

Use "lair [command] --help" for more information about a command.

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
