# Spire CLI Readme

Spire allows for broadcasting signed price messages through a network of peer-to-peer nodes over the gossip-sub protocol of [libp2p](https://libp2p.io/).

- Table of contents

## Installation

To install it, you'll first need Go installed on your machine. Then you can use
standard Go command:

```bash
go get -u github.com/chronicleprotocol/oracle-suite/cmd/spire
```

Alternatively, you can build Spire using `Makefile` directly from the repository:

```bash
git clone https://github.com/makerdao/oracle-suite.git
cd oracle-suite
make
```

This will build the binary in `bin`directory.

## Quick Start

Spire consists of two main parts:

1. P2P node (agent) responsible for connecting to other peers.
2. Spire CLI, that allows for sending and receiving messages.

### Minimal steps

- Create a new ethereum key.
- Edit the `spire.json` configuration file to point to the created file

```bash
{
  "ethereum": {
    "from": "0xYourEthereumAddress",
    "keystore": "/path/to/the/keystore/directory",
    "password": "/path/to/a/text/file/with/polain/text/password"
  }, 
  ...
}
```

- Start the agent.

```bash
spire agent
```

- Now you can push price messages into the network

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

- ...and you can pull all the prices captured by spire

```bash
spire pull prices
```

- or pull a price for a specific asset and a specific feed

```bash
spire pull price BTCUSD 0xFeedEthereumAddress
```
