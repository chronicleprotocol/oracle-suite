# Leeloo CLI Readme

Leeloo is an application run by Oracles. The application is responsible for collecting specified events from a different
blockchain (such as Arbitrium or Optimism) and sending them to the Spire P2P network.

Leeloo is one of the components of Maker Wormhole: https://forum.makerdao.com/t/introducing-maker-wormhole/11550

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
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

### Example configuration

```json
{
  "transport": {
    "libp2p": {
      "privKeySeed": "02082cf471002b5c5dfefdd6cbd30666ff02c4df90169f766877caec26ed4f88",
      "listenAddrs": ["/ip4/0.0.0.0/tcp/8000"],
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
    "password": "password",
  },
  "leeloo": {
    "listeners": {
      "wormhole": [
        {
          "rpc": [
            "https://ethereum.provider-1.example/rpc",
            "https://ethereum.provider-2.example/rpc",
            "https://ethereum.provider-3.example/rpc"
          ],
          "interval": 60,
          "blocksBehind": [30, 5760, 11520, 17280, 23040, 28800, 34560],
          "maxBlocks": 1000,
          "addresses": ["0x20265780907778b4d0e9431c8ba5c7f152707f1d"]
        }
      ]
    }
  }
}
```

### Configuration reference

- `transport` - Configuration parameters for transports mechanisms used to relay messages.
    - `libp2p` - Configuration parameters for the libp2p transport (Spire network).
        - `privKeySeed` - The random hex-encoded 32 bytes. It is used to generate unique identity in libp2p network. It may
          be empty to generate a random secret.
        - `listenAddrs` - The list of listen addresses for the libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `bootstrapAddrs` - The list of addresses of bootstrap nodes for the libp2p node encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `directPeersAddrs` - The list of direct peers addresses to which messages will be send directly encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format. This option has to be configured symmetrically
          at both ends.
        - `blockedAddrs` - The list of blocked peeers or addresses encoded using the
          [multiaddress](https://docs.libp2p.io/concepts/addressing/) format.
        - `disableDiscovery` - Disables node discoverability. If enabled, then IP address of a node will not be broadcast.
          to other peers. This option must be used along with `directPeersAddrs`.
- `feeds` - List of hex encoded addresses of other Oracles. Event messages from Oracles outside that list will be ignored.
- `ethereum` - Configuraiton of Ethereum wallet used to sign event messages.
  - `from` - Ethereum wallet address.
  - `keystore` - Keystore path.
  - `password` - Path to password file. If empty, no password is used.
- `leeloo` - Leeloo configuration.
  - `listeners` - Event listeners configuration.
    - `wormhole` - Configuration of "wormhole" event listener. It listens for `WormhholeGUID` events on Ethereum-compatible
      blockchains.
      - `rpc` - List of RPC server addresses. If more than one is used, then rpc-splitter is used. it is highly recommended
        to use at least three addresses from different providers.
      - `interval` - How often listener should check for a new events.
      - `blocksBehind` - The list of numbers that specify from which blocks, relative to the newest, events should be
        retrieved. 
      - `maxBlocks` - The number of block from which events may be fetched at once. The number must be large enough to 
        ensure that no more blocks are added to the blockchain, within the interval defined above.
      - `addresses` - Addresses of Wormhole contracts that emits `WormholeGUID` events.

## Commands

```
Usage:
  leeloo [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         

Flags:
  -c, --config string                                  ghost config file (default "./config.json")
  -h, --help                                           help for leeloo
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
      --version                                        version for leeloo

Use "leeloo [command] --help" for more information about a command.
➜  oracle-suite git:(sc-448/lair-storage-mechanism) ✗ 

```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
