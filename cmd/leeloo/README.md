# Leeloo CLI Readme

Leeloo is an application run by Oracles. This application is responsible for collecting specific events from other
blockchains (such as Arbitrium or Optimism), attesting them, and sending them to the Spire P2P network.

Leeloo is one of the components of Maker Teleport: https://forum.makerdao.com/t/introducing-maker-teleport/11550

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Supported events](#supported-events)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go get -u github.com/chronicleprotocol/oracle-suite/cmd/leeloo`.

Alternatively, you can build Gofer using `Makefile` directly from the repository. This approach is recommended if you
wish to work on Gofer source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## Configuration

To start working with Leeloo, you have to create configuration file first. By default, the default config file location
is `config.json` in the current working directory. You can change the config file location using the `--config` flag.
Leeloo supports JSON and YAML configuration files.

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
  "leeloo": {
    "listeners": {
      "teleportEVM": [
        {
          "ethereum": {
            "rpc": [
              "https://ethereum.provider-1.example/rpc",
              "https://ethereum.provider-2.example/rpc",
              "https://ethereum.provider-3.example/rpc"
            ],
            "timeout": 10,
            "gracefulTimeout": 1,
            "gracefulTimeout": 35
          },
          "interval": 60,
          "blocksDelta": [
            30,
            5760,
            11520,
            17280,
            23040,
            28800,
            34560
          ],
          "blocksLimit": 1000,
          "addresses": [
            "0x20265780907778b4d0e9431c8ba5c7f152707f1d"
          ]
        }
      ]
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
- `ethereum` - Configuration of the Ethereum wallet used to sign event messages.
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
            - `name` (`string`) - Name of metric. It can contain references to log fields in the format `$${path}`, where
              path is the dot-separated path to the field.
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
                - `max` - Use higher one.
                - `min` - Use lower one.
                - `replace` (default) - Replace the value with a newer one.
- `leeloo` - Leeloo configuration.
    - `listeners` - Event listeners configuration.
        - `[]teleportEVM` - Configuration of teleport bridge events on EVM compatible blockchains.
            - `ethereum` - Ethereum client configuration.
                - `rpc` (`string|[]string`) - List of RPC server addresses. It is recommended to use at least three
                  addresses from different providers.
                - `timeout` (`int`) - total timeout in seconds (default: 10).
                - `gracefulTimeout` (`int`) - timeout to graceful finish requests to slower RPC nodes, it is used only
                  when it is possible to return a correct response using responses from the remaining RPC nodes (
                  default: 1).
                - `gracefulTimeout` (`int`) - if multiple RPC nodes are used, determines how far one node can be behind
                  the last known block (default: 0).
            - `interval` (`integer`) - Specifies how often (in seconds) the event listener should check for new events.
            - `blocksDelta` (`[]integer`) - List of numbers that specify from which blocks, relative to the newest,
              events should be retrieved.
            - `blocksLimit` (`integer`) - The number of blocks from which events can be retrieved simultaneously. This
              number must be large enough to ensure that no more blocks are added to the blockchain during the time
              interval defined above.
            - `addresses` (`[]string`) - List of addresses of Teleport contracts that emits `TeleportGUID` events.
        - `[]teleportStarknet` - Configuration of teleport bridge events on Starknet.
            - `sequencer` (`string`) - Address of the sequencer endpoint.
            - `interval` (`integer`) - Specifies how often (in seconds) the event listener should check for new events.
            - `blocksDelta` (`[]integer`) - List of numbers that specify from which blocks, relative to the newest,
              events should be retrieved.
            - `blocksLimit` (`integer`) - The number of blocks from which events can be retrieved simultaneously. This
              number must be large enough to ensure that no more blocks are added to the blockchain during the time
              interval defined above.
            - `addresses` (`[]string`) - List of addresses of Teleport contracts that emits `TeleportGUID` events.

### Environment variables

It is possible to use environment variables anywhere in the configuration file. The syntax is similar as in the
shell: `${ENV_VAR}`. If the environment  variable is not set, the error will be returned during the application
startup. To escape the dollar sign, use `\$` or `$$`. The latter syntax is not supported inside variables. It is
possible to define default values for environment variables. To do so, use the following syntax: `${ENV_VAR-default}`.

## Supported events

Currently, only the `teleport` event type is supported:

- Type: `teleport`  
  This type of event is used for events emitted on Ethereum compatible blockchains, like Optimism or Arbitrium. It looks
  for `TeleportGUID` events on specified contract addresses.  
  Reference:  
  [https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol](https://github.com/makerdao/dss-teleport/blob/master/src/TeleportGUID.sol)  
  [https://github.com/chronicleprotocol/oracle-suite/blob/4eed6bcfc59b7eefba171dcc0ae3f4b7188ebb4e/pkg/event/publisher/ethereum/teleport.go#L156](https://github.com/chronicleprotocol/oracle-suite/blob/4eed6bcfc59b7eefba171dcc0ae3f4b7188ebb4e/pkg/event/publisher/ethereum/teleport.go#L156)

## Commands

```
Usage:
  leeloo [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         Start the agent

Flags:
  -c, --config string                                  ghost config file (default "./config.json")
  -h, --help                                           help for leeloo
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for leeloo

Use "leeloo [command] --help" for more information about a command.

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
