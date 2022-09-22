# Gofer CLI Readme

> As in a [tool](https://en.wikipedia.org/wiki/Gofer) that specializes in the delivery of special items.

Gofer is a tool that provides reliable asset prices taken from various sources.

If you need reliable price information, getting them from a single source is not the best idea. The data source may fail
or provide incorrect data. Gofer solves this problem. With Gofer, you can define precise price models that specify
exactly, from how many sources you want to pull prices and what conditions they must meet to be considered reliable.

## Table of contents

* [Installation](#installation)
* [Price models](#price-models)
* [Configuration](#configuration)
* [Commands](#commands)
    * [gofer price](#gofer-price)
    * [gofer pairs](#gofer-pairs)
    * [gofer agent](#gofer-agent)
* [License](#license)

## Installation

To install it, you'll first need Go installed on your machine. Then you can use standard Go
command: `go get -u github.com/chronicleprotocol/oracle-suite/cmd/gofer`.

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

```json
{
  "gofer": {
    "priceModels": {
      "BTC/USD": {
        "method": "median",
        "sources": [
          [
            {
              "origin": "bitstamp",
              "pair": "BTC/USD",
              "ttl": 60
            }
          ]
        ],
        "params": {
          "minimumSuccessfulSources": 1
        }
      }
    }
  }
}
```

All price models must be defined under the `priceModels` key as a map where a key is an asset pair name written
as `XXX/YYY`, where `XXX` is the base asset name and `YYY`
is the quote asset name. These symbols are case-insensitive. The `/` as a separator is the only requirement here.

Price model for each asset pair consists of three keys: `method`, `sources` and `params`:

- `sources` - contains a list of sources used to determine asset price. Each source must consist of one or more asset
  pairs. If multiple asset pairs are given, then the cross rate between them will be calculated. Each asset pair
  consists of two mandatory keys: `origin`, `pair`, and one optional: `ttl` (which is set to `60` by default).
    - `origin` - a name of a provider from which price will be obtained. Currently, following providers are supported:
        - `balancer` - [Balancer](https://balancer.finance/)
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
        - `ftx` - [FTX](https://ftx.com/)
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
        - `.` - a special value (single dot) which refers to another price model in the config.
    - `pair` - a name of a pair to be fetched from given origin.
    - `ttl` - a number of seconds after which the price should be updated. Additionally, if the price is older than the
      time defined by TTL by one minute, then the price will be considered outdated.

  As stated earlier, multiple sources may be provided to calculate the cross rate between different assets. For example,
  to get `BTC/JPY` price, you may provide the following list of sources:

    ```json
    [
      {"origin": "bitstamp", "pair": "BTC/USD"}, 
      {"origin": "fx", "pair": "USD/JPY"}
    ]
    ```

  To correctly calculate the cross rate, all adjacent pairs in a list must have a common asset.

- `params` - usage depends on the value of the `method` field.
- `method` - specifies the method used to calculate a single asset price from a given sources list. Currently, only
  the `median` method is supported:
    - `median` - calculates the median price from given sources. This method requires one parameter to be provided in
      the `params` field:
        - `minimumSuccessfulSources` - minimum number of successfully retrieved sources to consider calculated median
          price as reliable.
        - `postPriceHook` - In some cases a check should be done after the median price has been obtained. E.g. in the
          case of `rETH`, a circuit breaker value is checked against the obtained median, and if the deviation is high
          enough, a price error will be set.

### Origins configuration

Some origins might require additional configuration parameters like an `API Key`. In the current implementation, we
have `openexchangerates` and `coinmarketcap`. Both of these origins require an `API Key`. To configure these origins we
have to provide `origins` field in the configuration file.

Example:

```json
{
  "gofer": {
    "origins": {
      "openexchangerates": {
        "type": "openexchangerates",
        "params": {
          "apiKey": "API_KEY"
        }
      }
    }
  }
}
```

- `type` - this key corresponds to the built-in origin set
- `params` - this object will map the params to the specific origin configuration (apiKey is one example)

### Configuration reference

- `ethereum` - Ethereum client configuration. It is used by Origins, which pulls prices directly from the blockchain.
    - `rpc` (`string|[]string`) - List of RPC server addresses. It is recommended to use at least three addresses from
      different providers.
    - `timeout` (`int`) - total timeout in seconds (default: 10).
    - `gracefulTimeout` (`int`) - timeout to graceful finish requests to slower RPC nodes, it is used only
      when it is possible to return a correct response using responses from the remaining RPC nodes (
      default: 1).
    - `gracefulTimeout` (`int`) - if multiple RPC nodes are used, determines how far one node can be behind
      the last known block (default: 0).
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
              where
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
- `gofer` - Gofer configuration.
    - `rpcListenAddr` (`string`) - Listen address for the RPC endpoint provided as the combination of IP address and
      port number. This parameter is optional. If specified, Gofer will attempt to retrieve prices from the specified
      RPC endpoint.
    - `origins` - [Origins configuration](#origins-configuration)
    - `priceModels` - [Price models configuration](#price-models-configuration)

### Environment variables

It is possible to use environment variables anywhere in the configuration file. The syntax is similar as in the
shell: `${ENV_VAR}`. If the environment  variable is not set, the error will be returned during the application
startup. To escape the dollar sign, use `\$` or `$$`. The latter syntax is not supported inside variables. It is
possible to define default values for environment variables. To do so, use the following syntax: `${ENV_VAR-default}`.

## Commands

Gofer is designed from the beginning to work with other programs,
like [oracle-v2](https://github.com/makerdao/oracles-v2). For this reason, by default, a response is returned as
the [NDJSON](https://en.wikipedia.org/wiki/JSON_streaming) format. You can change the output format to `plain`, `json`
, `ndjson`, or `trace` using the `--format` flag:

- `plain` - simple, human-readable format with only basic information.
- `json` - json array with list of results.
- `ndjson` - same as `json` but instead of array, elements are returned in new lines.
- `trace` - used to debug price models, prints a detailed graph with all possible information.

### `gofer price`

The `price` command returns a price for one or more asset pairs. If no pairs are provided then prices for all asset
pairs defined in the config file will be returned. When at least one price fails to be retrieved correctly, then the
command returns a non-zero status code.

```
Return prices for given PAIRs.

Usage:
  gofer prices [PAIR...] [flags]

Aliases:
  prices, price

Flags:
  -h, --help   help for prices

Global Flags:
  -c, --config string                    config file (default "./gofer.json")
  -f, --format plain|trace|json|ndjson   output format (default ndjson)
      --log.format text|json             log format
  -v, --log.verbosity string             verbosity level (default "info")
      --norpc                            disable the use of RPC agent
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

```
{
   "type":"aggregator",
   "base":"BTC",
   "quote":"USD",
   "price":45242.13,
   "bid":45236.308,
   "ask":45239.98,
   "vol24h":0,
   "ts":"2021-05-18T10:30:00Z",
   "params":{
      "method":"median",
      "minimumSuccessfulSources":"3"
   },
   "prices":[
      {
         "type":"origin",
         "base":"BTC",
         "quote":"USD",
         "price":45227.05,
         "bid":45221.79,
         "ask":45227.05,
         "vol24h":8339.77051164,
         "ts":"2021-05-18T10:31:16Z",
         "params":{
            "origin":"bitstamp"
         }
      },
      {
         "type":"origin",
         "base":"BTC",
         "quote":"USD",
         "price":45242.13,
         "bid":45236.308,
         "ask":45240.468,
         "vol24h":0,
         "ts":"2021-05-18T10:31:18.687607Z",
         "params":{
            "origin":"bittrex"
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
  -h, --help   help for pairs

Global Flags:
  -c, --config string                    config file (default "./gofer.json")
  -f, --format plain|trace|json|ndjson   output format (default ndjson)
      --log.format text|json             log format
  -v, --log.verbosity string             verbosity level (default "info")
      --norpc                            disable the use of RPC agent
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
