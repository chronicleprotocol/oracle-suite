# Spire CLI Readme

Spire allows for broadcasting signed price messages through a network of peer-to-peer nodes over the gossip-sub protocol
of [libp2p](https://libp2p.io/).

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Usage](#usage)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/spire@latest`

Alternatively, you can build Spire using `Makefile` directly from the repository. This approach is recommended if you
wish to work on Spire source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Spire, you have to create configuration file first. By default, the default config file location
is `config.hcl` in the current working directory. You can change the config file location using the `--config` flag.
Spire supports JSON and YAML configuration files.

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
  "ethereum": {
    "from": "0x2d800d93b065ce011af83f316cef9f0d005b0aa4",
    "keystore": "./keys",
    "password": "password"
  },
  "spire": {
    "rpcListenAddr": "127.0.0.1:9100",
    "pairs": [
      "AAVEUSD",
      "AVAXUSD",
      "BALUSD",
      "BATUSD",
      "BTCUSD",
      "COMPUSD",
      "CRVUSD",
      "DOTUSD",
      "ETHBTC",
      "ETHUSD",
      "FILUSD",
      "LINKUSD",
      "LRCUSD",
      "MANAUSD",
      "PAXGUSD",
      "SNXUSD",
      "SOLUSD",
      "UNIUSD",
      "USDTUSD",
      "WNXMUSD",
      "XRPUSD",
      "XTZUSD",
      "YFIUSD",
      "ZECUSD",
      "ZRXUSD",
      "STETHUSD",
      "WSTETHUSD",
      "MATICUSD"
    ]
  }
}
```

### Configuration reference

- `transport` - Configuration parameters for transports mechanisms used to relay messages.
    - `transport` (string|[]string) - Transport to use. Supported mechanism are: `libp2p` and `webapi`. If empty,
      the `libp2p` is used.
    - `libp2p` - Configuration parameters for the libp2p transport.
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
    - `webapi` - Configuration parameters for the webapi transport. WebAPI transport uses the HTTP protocol to send
      and receive messages. It should be used over a secure network like TOR, I2P or VPN.
        - `listenAddr` - Address on which the WebAPI server will listen for incoming connections. The address must be
          in the format `host:port`. When used with a TOR hidden service, the server should listen on localhost.
        - `socks5ProxyAddr` - Address of the SOCKS5 proxy server. The address must be in the format `host:port`.
        - `addressBookType` (`string|[]string`) - Type of address book to use. Supported types are: `ethereum`
          and `static`.
          `ethereum` type uses a contract deployed on the Ethereum-compatible blockchain to store the list of addresses,
          `static` type uses a static list of addresses defined in the configuration file. It is possible to use
          multiple
          address book types at the same time.
            - `ethereumAddressBook` - Configuration parameters for the Ethereum address book.
                - `addressBookAddr` - Ethereum address of the address book contract.
                - `ethereum` - Ethereum client configuration that is used to interact with the address book contract.
                    - `rpc` (`string|[]string`) - List of RPC server addresses. It is recommended to use at least three
                      addresses from different providers.
                    - `timeout` (`int`) - total timeout in seconds (default: 10).
                    - `gracefulTimeout` (`int`) - timeout to graceful finish requests to slower RPC nodes, it is used
                      only
                      when it is possible to return a correct response using responses from the remaining RPC nodes (
                      default: 1).
                    - `maxBlocksBehind` (`int`) - if multiple RPC nodes are used, determines how far one node can be
                      behind
                      the last known block (default: 0).
            - `staticAddressBook` - Configuration parameters for the static address book.
                - `remoteAddrs` (`[]string`) - List of remote addresses to which messages will be sent.
- `feeds` (`[]string`) - List of hex-encoded addresses of other Oracles. Event messages from Oracles outside that list
  will be ignored.
- `ethereum` - Configuration of the Ethereum wallet used to sign messages.
    - `from` (`string`) - The Ethereum wallet address.
    - `keystore` (`string`) - The keystore path.
    - `password` (`string`) - The path to the password file. If empty, the password is not used.
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
            - `name` (`string`) - Name of metric. It can contain references to log fields in the format `%{path}`,
              where path is the dot-separated path to the field.
            - `tags` (`[string][]string`) - List of metric tags. They can contain references to log fields in the
              format `%{path}`, where path is the dot-separated path to the field.
            - `value` (`string`) - Dot-separated path of the field with the metric value. If empty, the value 1 will be
              used as the metric value.
            - `scaleFactor` (`float`) - Scales the value by the specified number. If it is zero, scaling is not
              applied (default: 0).
            - `onDuplicate` (`string`) - Specifies how duplicated values in the same interval should be handled. Allowed
              options are:
                - `sum` - Add values.
                - `sub` - Subtract values.
                - `max` - Use higher one.
                - `min` - Use lower one.
                - `replace` (default) - Replace the value with a newer one.
- `spire` - Spire configuration.
    - `rpcListenAddr` (`string`) - Listen address for the RPC endpoint provided as the combination of IP address and
      port number.
    - `rpcAgentAddr` (`string`) - Address of the RPC agent.
    - `pairs` (`[]string`) - List of price pairs to be monitored. Only pairs in this list will be available via pull
      command.

### Environment variables

It is possible to use environment variables anywhere in the configuration file. The syntax is similar as in the
shell: `${ENV_VAR}`. If the environment variable is not set, the error will be returned during the application
startup. To escape the dollar sign, use `\$` It is possible to define default values for environment variables.
To do so, use the following syntax: `${ENV_VAR-default}`.

## Usage

### Starting the agent.

```bash
spire agent
```

### Pushing price messages into the network

```bash
cat <<"EOF" | spire push price
{
    "wat": "BTCUSD",
		// price is 32 bytes (no 0x prefix) `seth --to-wei "$_price" eth`
		// i.e. 1.32 * 10e18 => "13200000000000000000"
    "val": "13200000000000000000",
		// unix epoch (seconds only)
		"age": 123456789,
		"r": <string>, // 64 chars long, hex encoded 32 byte value
		"s": <string>, // 64 chars long, hex encoded 32 byte value
		"v": <string>,  // 2 chars long, hex encoded 1 byte value
    "trace": <string> // (optional) human readable price calculation description
}
EOF
```

### Pulling all the prices captured by Spire

```bash
spire pull prices
```

### Pulling a price for a specific asset and a specific feed

```bash
spire pull price BTCUSD 0xFeedEthereumAddress
```

### Streaming price messages from the network

```bash
spire stream prices
```

## Commands

```
Usage:
  spire [command]

Available Commands:
  agent       Starts the Spire agent
  help        Help about any command
  pull        Pulls data from the Spire datastore (require agent)
  push        Push a message to the network (require agent)
  stream      Streams data from the network

Flags:
  -c, --config string                                  spire config file (default "./config.hcl")
  -h, --help                                           help for spire
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for spire

Use "spire [command] --help" for more information about a command.
```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
