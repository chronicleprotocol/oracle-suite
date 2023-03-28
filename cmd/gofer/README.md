# Gofer CLI Readme

> As in a [tool](https://en.wikipedia.org/wiki/Gofer) that specializes in the delivery of special items.

Gofer is a tool that provides reliable asset prices taken from various sources.

If you need reliable price information, getting them from a single source is not the best idea. The data source may fail
or provide incorrect data. Gofer solves this problem. With Gofer, you can define precise price models that specify
exactly, from how many sources you want to pull prices and what conditions they must meet to be considered reliable.

## Table of contents

* [Installation](#installation)
* [Configuration](#configuration)
* [Commands](#commands)
    * [gofer price](#gofer-price)
    * [gofer pairs](#gofer-pairs)
    * [gofer agent](#gofer-agent)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go install github.com/chronicleprotocol/oracle-suite/cmd/gofer@latest`

Alternatively, you can build Gofer using `Makefile` directly from the repository. This approach is recommended if you
wish to work on Gofer source.

```bash
git clone https://github.com/ma/oracle-suite.git
cd oracle-suite
make
```

## Configuration

### Price models configuration

To start working with Gofer, you have to define price models first. Price models are defined in a JSON or YAML file. By
default, the default config file location is `gofer.json` in the current directory. You can change the config file
location using the `--config` flag.

Simple price model for the `BTC/USD` asset pair may look like this:

```hcl
gofer {
  price_model "BTC/USD" "median" {
    source "BTC/USD" "origin" { origin = "bitfinex" }
    source "BTC/USD" "origin" { origin = "coinbasepro" }
    source "BTC/USD" "origin" { origin = "kraken" }
    min_sources = 1
  }
}
```

All price models must be defined under as `price_model` blocks where label is an asset pair name written as `XXX/YYY`,
where `XXX` is the base asset name and `YYY` is the quote asset name. These symbols are case-insensitive.

Each `price_model` and `source` block have two labels. First label is an asset pair name written as `XXX/YYY`,
where `XXX` is the base asset name and `YYY` is the quote asset name. These symbols are case-insensitive. Second label
defines how price should be calculated. Currently, following price calculation methods are supported:

- `origin` - returns price from the first source that provides it. It required one argument `origin` which is a name
  of a provided from which price will be obtained.
- `median` - calculates median price from sources. It requires at least one `source` block and one optional argument
  `min_sources` which is a minimum number of sources that must provide price for it to be considered reliable.
- `indirect` - calculates indirect price using two or more asset pairs. It requires at least one `source` block.
  Price is calculated be calculating cross rate between asset pairs. For example, to calculate price of `BTC/USD` a
  following list of sources can be used: `BTC/ETH`, `ETH/USD`.

Supported origins:

- `balancer` - [Balancer](https://balancer.finance/)
- `balancerV2` - [Balancer](https://balancer.finance/)
- `binance` - [Binance](https://binance.com/)
- `bitfinex` - [Bitfinex](https://bitfinex.com/)
- `bitstamp` - [Bitstamp](https://bitstamp.net/)
- `bithumb` - [Bithumb](https://bithumb.com/)
- `bittrex` - [Bittrex](https://bittrex.com/)
- `coinbasepro` - [CoinbasePro](https://pro.coinbase.com/)
- `cryptocompare` - [CryptoCompare](https://cryptocompare.com/)
- `coinmarketcap` - [CoinMarketCap](https://coinmarketcap.com/)
- `ddex` - [DDEX](https://ddex.net/)
- `folgory` - [Folgory](https://folgory.com/)
- `fx` - [exchangeratesapi.io](https://exchangeratesapi.io/) (fiat currencies)
- `gateio` - [Gateio](https://gate.io/)
- `gemini` - [Gemini](https://gemini.com/)
- `hitbtc` - [HitBTC](https://hitbtc.com/)
- `huobi` - [Huobi](https://huobi.com/)
- `kraken` - [Kraken](https://kraken.com/)
- `kucoin` - [KuCoin](https://kucoin.com/)
- `loopring` - [Loopring](https://loopring.org/)
- `okex` - [OKEx](https://okex.com/)
- `openexchangerates` - [OpenExchangeRates](https://openexchangerates.org)
- `poloniex` - [Poloniex](https://poloniex.com/)
- `sushiswap` - [Sushiswap](https://sushi.com/)
- `uniswap` - [Uniswap V2](https://uniswap.org/)
- `uniswapV2` - [Uniswap V2](https://uniswap.org/)
- `uniswapV3` - [Uniswap V3](https://uniswap.org/blog/uniswap-v3/)
- `upbit` - [Upbit](https://upbit.com/)

### Origins configuration

Some origins might require additional configuration parameters like an API key. You can define these parameters in the
`origins` section of the config file.

```hcl
gofer {
  origin "openexchangerates" {
    type   = "openexchangerates"
    params = {
      api_key = "YOUR_API_KEY"
    }
  }
}
```

The block label is the name of the origin. This name is used in the `source` block to reference the origin. The `type`
parameter defines the origin type. If label and type are the same, the default origin is replaced.

All origins accept the `symbol_aliases` parameter, which is a map of asset symbols to their aliases. This is useful
when the origin uses different symbols than the ones used in the price model. For example, to treat `USDC` as `USD` in
a price model, you can use the following configuration under the `origin` block:

```hcl
symbol_aliases = {
  "USD" = "USDC"
}
```

Depending on the origin type, different additional parameters can be defined:

- `balancer`, `balancerV2`, `sushiswap`, `curve`, `curvefinance`, `wsteth`, `rocketpool`, `uniswap`, `uniswapV2`
  `uniswapV3`:
    - `contracts` - a map of pairs of asset symbols to their Balancer V2 pool contract addresses. For example:
      ```hcl
      contracts = {
        "ETH/USD" = "0x58f6b77148BE49BF7898472268ae8f26377d0AA6"
      }
      ```

- `coinmarketcap`, `fx`, `openexchangerates`:
    - `api_key` - API key used to access the origin.

- `curve`, `curvefinance`, `balancerV2`, `wsteth`, `rocketpool`:
    - `ethereum_client` - Ethereum client used to access the blockchain data.

Additionally, most of the origins accept the `url` parameter, which is a URL of the origin API. If not specified,
the default URL will be used.

### Hooks

In some cases a check should be done after the median price has been obtained. E.g. in the case of `rETH`, a circuit
breaker value is checked against the obtained median, and if the deviation is high enough, a price error will be set.

To define a hook, you can use the `hook` block:

```hcl
gofer {
  # ...
  hook "RETH/ETH" {
    post_price = {
      ethereum_client  = "default"
      circuit_contract = "0xa3105dee5ec73a7003482b1a8968dc88666f3589"
    }
  }
  # ...
}
```

### Example configuration

```hcl
gofer {
  # Origin configuration. If label and type are the same, the default origin is replaced with the one defined here.
  origin "uniswapV3" {
    # Base origin type.
    type = "uniswapV3"

    # List of origin parameters
    params = {
      ethereum_client = "default"
      symbol_aliases  = {
        "BTC" = "WBTC",
        "ETH" = "WETH",
        "USD" = "USDC"
      }
      contracts = {
        "GNO/WETH"  = "0xf56d08221b5942c428acc5de8f78489a97fc5599",
        "LINK/WETH" = "0xa6cc3c2531fdaa6ae1a3ca84c2855806728693e8",
        "MKR/USDC"  = "0xc486ad2764d55c7dc033487d634195d6e4a6917e",
        "MKR/WETH"  = "0xe8c6c9227491c0a8156a0106a0204d881bb7e531",
        "USDC/WETH" = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
        "YFI/WETH"  = "0x04916039b1f59d9745bf6e0a21f191d1e0a84287"
      }
    }
  }

  # Price model configuration. First label is the pair name and must be in the format of "BASE/QUOTE". Second label
  # specifies how the price is calculated. Details about the price model can be found in the "PPrice models configuration"
  # section of this document.
  price_model "ETH/USD" "median" {
    source "ETH/USD" "indirect" {
      source "ETH/BTC" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "ETH/USD" "origin" { origin = "bitstamp" }
    source "ETH/USD" "origin" { origin = "coinbasepro" }
    source "ETH/USD" "origin" { origin = "gemini" }
    source "ETH/USD" "origin" { origin = "kraken" }
    source "ETH/USD" "origin" { origin = "uniswapV3" }
    min_sources = 3
  }

  # Hook configuration.
  hook "ETH/USD" {
    post_price = {
      ethereum_client  = "default"
      circuit_contract = "0x1234567890123456789012345678901234567890"
    }
  }
}

# Gofer origins and hooks may require access to Ethereum blockchain. This section defines Ethereum clients
# used by Gofer. 
# Optional.
ethereum {
  # Configuration for Ethereum clients. The client name is used to reference the client in other sections.
  # It is possible to have multiple clients in the configuration.
  client "default" {
    # RPC URLs is a list of Ethereum RPC URLs to use for the client. Ethereum client uses RPC-Splitter which compares
    # responses from multiple RPC URLs to verify that none of them are compromised. At least three URLs are recommended.
    rpc_urls = ["https://eth.public-rpc.com"]

    # Chain ID of the Ethereum network.
    chain_id = 1

    # Ethereum key to use for signing transactions.
    # Optional. If not specified, the default key is used, the signing is done by the Ethereum node.
    ethereum_key = "default"
  }
}
```

### Environment variables

It is possible to use environment variables anywhere in the configuration file. Environment variables are accessible
in the `env` object. For example, to use the `HOME` environment variable in the configuration file, use `env.HOME`.

## Commands

Gofer is designed from the beginning to work with other programs,
like [oracle-v2](https: //github.com/makerdao/oracles-v2). For this reason, by default, a response is returned as
the [NDJSON](https : //en.wikipedia.org/wiki/JSON_streaming) format. You can change the output format
to `plain`, `json`, `ndjson`, or `trace` using the `--format` flag :

- `plain` - simple, human-readable format with only basic information.
- `json` - json array with list of results.
- `ndjson` - same as `json` but instead of array, elements are returned in new lines.
- `trace` - used to debug price models, prints a detailed graph with all possible information.

### `gofer price`

The `price` command returns a price for one or more asset pairs.If no pairs are provided then prices for all asset
pairs defined in the config file will be returned.When at least one price fails to be retrieved correctly, then the
command returns a non-zero status code.

```

Return prices for given PAIRs.

Usage:
gofer prices [PAIR...] [flags]

Aliases:
prices, price

Flags:
-h, --help help for prices

Global Flags:
-c, --config string config file (default "./gofer.json")
-f, --format plain|trace|json|ndjson output format (default ndjson)
--log.format text|json log format
-v, --log.verbosity string verbosity level (default "info")
--norpc disable the use of RPC agent

```

JSON output for a single asset pair consists of the following fields:

- `type` - may be `aggregator` or `origin`. The `aggregator` value means that a given price has been calculated based on
  other prices, the `origin` value is used when a price is returned directly from an origin.
- `base` - the base asset name.
- `quote` - the quote asset name.
- `price` - the current asset price.
- `bid` - the bid price, 0 if it is impossible to retrive or calculate bid price.
- `ask` - the ask price, 0 if it is impossible to retrive or calculate ask price.
- `vol24` - the volume from last 24 hours, 0 if it is impossible to retrieve or calculate volume.
- `ts` - the date from which the price was retrieved.
- `params` - the list of additional parameters, it always contains the `method` field for aggregators and the `origin`
  field for origins.
- `error` - the optional error message, if this field is present, then price is not relaiable.
- `price` - the list of prices used in calculation. For origins it's always empty.

Example JSON output for BTC/USD pair:

```json
{
  "type": "aggregator",
  "base": "BTC",
  "quote": "USD",
  "price": 45242.13,
  "bid": 45236.308,
  "ask": 45239.98,
  "vol24h": 0,
  "ts": "2021-05-18T10:30:00Z",
  "params": {
    "method": "median",
    "minimumSuccessfulSources": "3"
  },
  "prices": [
    {
      "type": "origin",
      "base": "BTC",
      "quote": "USD",
      "price": 45227.05,
      "bid": 45221.79,
      "ask": 45227.05,
      "vol24h": 8339.77051164,
      "ts": "2021-05-18T10:31:16Z",
      "params": {
        "origin": "bitstamp"
      }
    },
    {
      "type": "origin",
      "base": "BTC",
      "quote": "USD",
      "price": 45242.13,
      "bid": 45236.308,
      "ask": 45240.468,
      "vol24h": 0,
      "ts": "2021-05-18T10:31:18.687607Z",
      "params": {
        "origin": "bittrex"
      }
    }
  ]
}
```

Examples:

```
$ gofer price --format plain
BTC/USD 45291.110000
ETH/USD 3501.636879

$ gofer price BTC/USD --format trace
Price for BTC/USD:
───aggregator(method:median, minimumSuccessfulSources:3, pair:BTC/USD, price:45287.18, timestamp:2021-05-18T10:35:00Z)
    ├──origin(origin:bitstamp, pair:BTC/USD, price:45298.02, timestamp:2021-05-18T10:35:39Z)
    ├──origin(origin:bittrex, pair:BTC/USD, price:45287.18, timestamp:2021-05-18T10:35:43.335185Z)
    ├──origin(origin:coinbasepro, pair:BTC/USD, price:45282.53, timestamp:2021-05-18T10:35:43.285832Z)
    ├──origin(origin:gemini, pair:BTC/USD, price:45266.13, timestamp:2021-05-18T10:35:00Z)
    └──origin(origin:kraken, pair:BTC/USD, price:45291.2, timestamp:2021-05-18T10:35:43.470442Z)
```

### `gofer pairs`

The `pairs` command can be used to check if there are defined price models for given pairs and also to debug existing
price models. When the price model is missing, then the command returns a non-zero status code. If no pairs are provided
then all asset pairs defined in the config file will be returned. In combination with the `--format=trace` flag, the
command will return price models for given pairs.

```
List all supported asset pairs.

Usage:
gofer pairs [PAIR...] [flags]

Aliases:
pairs, pair

Flags:
-h, --help help for pairs

Global Flags:
-c, --config string config file (default "./gofer.json")
-f, --format plain|trace|json|ndjson output format (default ndjson)
--log.format text|json log format
-v, --log.verbosity string verbosity level (default "info")
--norpc disable the use of RPC agent
```

Examples:

```
$ gofer pairs
"BTC/USD"
"ETH/USD"

$ gofer pairs --format plain
BTC/USD
ETH/USD

$ gofer pair BTC/USD --format trace
Graph for BTC/USD:
───median(pair:BTC/USD)
    ├──origin(origin:bitstamp, pair:BTC/USD)
    ├──origin(origin:bittrex, pair:BTC/USD)
    ├──origin(origin:coinbasepro, pair:BTC/USD)
    ├──origin(origin:gemini, pair:BTC/USD)
    └──origin(origin:kraken, pair:BTC/USD)
```

### `gofer agent`

The `agent` command runs Gofer in the agent mode.

Excessive use of the `gofer price` command may invoke many API calls to external services which can lead to
rate-limiting. To avoid this, the prices that were previously retrieved can be reused and updated only as often as is
defined in the `ttl` parameters. To do this, Gofer needs to be run in agent mode.

At first, the agent mode has to be enabled in the configuration file by adding the following field:

```json
{
  "gofer": {
    "rpc": {
      "address": "127.0.0.1:8080"
    }
  }
}
```

The above address is used as the listen address for the internal RPC server and as a server address for a client. Next,
you have to launch the agent using the `gofer agent` command.

From now, the `gofer price` command will retrieve asset prices from the agent instead of retrieving them directly from
the origins. If you want to temporarily disable this behavior you have to use the `--norpc` flag.

## License

[The GNU Affero General Public License](https://www.notion.so/LICENSE)
