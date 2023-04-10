# RPC-Splitter CLI Readme

RPC-Splitter is an Ethereum RPC proxy that splits a request across multiple endpoints, compares them against each other
to guarantee data integrity.

## Table of contents

* [Installation](#installation)
* [How it works](#how-it-works)
* [Supported methods](#supported-methods)
* [CORS](#cors)
* [Commands](#commands)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/rpc-splitter@latest`

Alternatively, you can build RPC-Splitter using `Makefile` directly from the repository. This approach is recommended if
you wish to work on RPC-Splitter source.

```bash
git clone https://github.com/chronicleprotocol/oracle-suite.git
cd oracle-suite
make
```

## How it works

Using third-party RPC providers offers convenience, but when tasks demand high levels of security, the entire system's
security becomes dependent on a single RPC provider. The RPC Splitter addresses this issue.

The RPC Splitter functions similarly to a proxy; however, instead of directing the request to one server, it forwards it
to multiple servers and compares the results, ensuring data integrity.

The behavior of the RPC Splitter varies depending on the number of provided endpoints. With one endpoint, the value is
forwarded as is. For two endpoints, both responses must be identical. When there are three or more endpoints, one
response may differ.

## Supported methods

- `eth_blockNumber` - Returns the lowest block number that is equal to or greater than the last known block minus the
  value specified in the `--max-blocks-behind` argument.
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
- `eth_gasPrice` - Median value is returned. For two valid responses, a lower one.
- `eth_estimateGas` - Median value is returned. For two valid responses, a lower one.
- `eth_feeHistory`
- `eth_maxPriorityFeePerGas` - Median value is returned. For two valid responses, a lower one.
- `eth_chainId`
- `net_version`

If the method requires a block number, the newest and pending tags will be replaced with the latest block number using
the same algorithm as the eth_blockNumber endpoint. The earliest tag is not supported.

## CORS

It is possible to enable basic CORS support, which allows the use of RPC-Splitter with tools like Metamask. When CORS is
enabled, the `Access-Control-Allow-Origin` header will always match the `Origin` header from the request.

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
      --eth-rpc strings                                list of ethereum RPC nodes
  -g, --graceful-timeout int                           set timeout to graceful finish requests to slower RPC nodes (default 1)
  -h, --help                                           help for rpc-splitter
  -l, --listen string                                  listen address (default "127.0.0.1:8545")
      --log.format text|json                           log format (default text)
  -v, --log.verbosity panic|error|warning|info|debug   verbosity level (default warning)
  -b, --max-blocks-behind int                          determines how far one node can be behind the last known block (default 10)
  -t, --timeout int                                    set request timeout in seconds (default 10)
      --version                                        version for rpc-splitter
```

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
