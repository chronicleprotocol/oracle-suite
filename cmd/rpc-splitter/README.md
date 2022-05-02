# RPC-Splitter CLI Readme

RPC-Splitter is an Ethereum RPC proxy that splits a request across multiple endpoints and verifies that all of them
return the same response.

## Table of contents

* [Installation](#installation)
* [How it works](#how-it-works)
* [Supported methods](#supported-methods)
* [CORS](#cors)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go get -u github.com/chronicleprotocol/oracle-suite/cmd/rpc-splitter`.

Alternatively, you can build RPC-Splitter using `Makefile` directly from the repository. This approach is recommended if
you wish to work on RPC-Splitter source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## How it works

Using third-party RPC providers is convenient, but for tasks that require a high level of security, it makes the
security of the entire system dependent on that one RPC provider. The RPC Splitter helps to solve this problem.

The RPC Splitter works similarly to a proxy, but instead of forwarding the request to one server, it forwards it to
multiple servers and compares the results against each other, thereby guaranteeing data integrity. If more than two
endpoints are specified then if one of the servers returns a different result or error, it will be ignored.

## Supported methods

- `eth_blockNumber` - returns the lowest block number that is equal to or greater than the median of all block numbers
  minus 3
- `eth_getBlockByHash`
- `eth_getBlockByNumber`
- `eth_getTransactionByHash`
- `eth_getTransactionCount`
- `eth_getTransactionReceipt`
- `eth_sendRawTransaction`
- `eth_getBalance`
- `eth_getCode`
- `eth_getStorageAt`
- `eth_call`
- `eth_getLogs`
- `eth_gasPrice` - median value is returned
- `eth_estimateGas` - median value is returned
- `eth_feeHistory`
- `eth_maxPriorityFeePerGas` - median value is returned
- `eth_chainId`
- `net_version`

If the method requires a block number, for `newest` and `pending` tags, the lowest block number that is equal to or
greater than the median of all block numbers minus 3 is returned. The `earliest` tag is not supported.

## CORS

It is possible to enable simple CORS support to allow using RPC-Splitter with tools such as Metamask. When CORS is
enabled, then the `Access-Control-Allow-Origin` header will always equal the `Origin` header from the request.

## Commands

```
Usage:
  rpc-splitter [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  run         Start server

Flags:
  -c, --enable-cors                                    enables CORS requests for all origins
      --eth-rpc strings                                list of ethereum nodes
  -h, --help                                           help for rpc-splitter
  -l, --listen string                                  listen address (default "127.0.0.1:8545")
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
  -t, --timeout int                                    Set request timeout (in seconds) for all RPC calls (default 10)
      --version                                        version for rpc-splitter

Use "rpc-splitter [command] --help" for more information about a command.
```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
