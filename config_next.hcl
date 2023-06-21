# Test config for the gofernext and ghostnext apps. Not ready for production use.

gofernext {
#  origin "balancer" {} # todo, Need web3

  origin "binance" {
    type = "tick_generic_jq"
    url  = "https://api.binance.com/api/v3/ticker/24hr"
    jq   = ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
  }

  origin "bitfinex" {
    type = "tick_generic_jq"
    url  = "https://api-pub.bitfinex.com/v2/tickers?symbols=t$${ucbase}$${ucquote}"
    jq   = "{price: .[0][7], time: now|round, volume: .[0][8]}"
  }

  origin "bitstamp" {
    type = "tick_generic_jq"
    url  = "https://www.bitstamp.net/api/v2/ticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: .timestamp, volume: .volume}"
  }

#  origin "bithumb" {} # Not used, ETH/USDT price is not correct

#  origin "bittrex" {} # Not used, API does not work, bankruptcy

  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

#  origin "coinmarketcap" {} # Not used, need API key

#  origin "cryptocompare" {
#    type = "tick_generic_jq"
#    url  = "https://min-api.cryptocompare.com/data/pricemultifull?fsyms=$${ucbase}&tsyms=$${ucquote}&tryConversion=false&extraParams=gofer&relaxedValidation=true"
#    jq   = "{price: .RAW.$${ucbase}.$${ucquote}.PRICE, time: .RAW.$${ucbase}.$${ucquote}.LASTUPDATE, volume: .RAW.$${ucbase}.$${ucquote}.VOLUME24HOUR}"
#  } # Not used, need to replace variables in jq

#  origin "curve" {} # todo, Need web3

#  origin "ddex" {
#    type = "tick_generic_jq"
#    url  = "https://api.ddex.io/v4/markets/tickers"
#    jq   = ".data.tickers[] | select(.marketId == ($ucbase)-($ucquote)) | {price: .price, volume: .volume, time: (.updateAt / 1000)}"
#  } # Not used, ETH/DAI price is not correct, how to insert `-`

#  origin "folgory" {} # Not used

#  origin "fx" {} # Not used

#  origin "gateio" {} # Not used

  origin "gemini" {
    type = "tick_generic_jq"
    url  = "https://api.gemini.com/v1/pubticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: (.volume.timestamp/1000), volume: null}"
  }

#  origin "hitbtc" {
#    type = "tick_generic_jq"
#    url  = "https://api.hitbtc.com/api/2/public/ticker?symbols=$${ucbase}$${ucquote}"
#    jq   = "{price: .[0].last|tonumber, time: .[0].timestamp|strptime("%Y-%m-%dT%H:%M:%S.%jZ")|mktime, volume: .[0].volume|tonumber}"
#  } # Not used

  origin "huobi" {
    type = "tick_generic_jq"
    url  = "https://api.huobi.pro/market/tickers"
    jq   = ".data[] | select(.symbol == ($lcbase+$lcquote)) | {price: .close, volume: .vol, time: now|round}"
  }

#  origin "ishares" {} # todo, Need parse html

#  origin "kraken" {
#    type = "tick_generic_jq"
#    url  = "https://api.kraken.com/0/public/Ticker?pair=$${ucbase}/$${ucquote}"
#    jq   = "{price: .result.($ucbase+\/+$ucquote).c[0]|tonumber, time: now|round , volume: .result.($ucbase+\/+$ucquote).v[0]|tonumber}"
##    jq   = "{price: .result.\"ETH\/USD\".c[0]|tonumber, time: now|round , volume: .result.\"ETH/USD\".v[0]|tonumber}"
#  } # todo, how to fill pair name

#  origin "kucoin" {} # Not used

#  origin "kyber" {} # Not used

#  origin "loopring" {} # Not used

#  origin "okex" {} # Not used

#  origin "okx" {
#    type = "tick_generic_jq"
#    url  = "https://www.okx.com/api/v5/market/ticker?instId=$${ucbase}-$${ucquote}-SWAP"
#    jq   = "{price: .data[0].last|tonumber, time: .data[0].ts|tonumber, volume: .data[0].vol24h|tonumber}"
#  } # todo, time should be divided by 1000

#  origin "openexchangerates" {} # todo, Need api key

#  origin "poloniex" {} # Not used

#  origin "rocketpool" {} # todo, Need web3

#  origin "sushiswap" {} # todo, Need web3

#  origin "uniswap" {} # todo, Need web3

  origin "upbit" {
    type = "tick_generic_jq"
    url  = "https://api.upbit.com/v1/ticker?markets=$${ucquote}-$${ucbase}"
    jq   = "{price: .[0].trade_price, time: (.[0].timestamp/1000), volume: .[0].acc_trade_volume_24h}"
  }

#  origin "wsteth" {} # todo, Need web3

  data_model "BTC/USD" {
    median {
      min_values = 3
      origin "coinbase" { query = "BTC/USD" }
      origin "bitstamp" { query = "BTC/USD" }
      origin "gemini" { query = "BTC/USD" }
#      origin "kraken" { query = "BTC/USD" }
    }
  }

  data_model "ETH/BTC" {
    median {
      min_values = 3
      origin "binance" { query = "ETH/BTC" }
      origin "bitstamp" { query = "ETH/BTC" }
      origin "coinbase" { query = "ETH/BTC" }
      origin "gemini" { query = "ETH/BTC" }
#      origin "kraken" { query = "ETH/BTC" }
    }
  }

  data_model "ETH/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "ETH/BTC" }
        reference { data_model = "BTC/USD" }
      }
      origin "bitstamp" { query = "ETH/USD" }
      origin "gemini" { query = "ETH/USD" }
    }
  }

  data_model "LINK/USD" {
    median {
      min_values = 2
      origin "bitstamp" { query = "LINK/USD" }
      origin "gemini" { query = "LINK/USD" }
    }
  }

  data_model "MANA/USD" {
    median {
      min_values = 2
      indirect {
        origin "binance" { query = "MANA/BTC" }
        reference { data_model = "BTC/USD" }
      }
#      origin "binance" { query = "MANA/USD" } # Not work
      origin "coinbase" { query = "MANA/USD" }
#      origin "kraken" { query = "MANA/USD" }
#      indirect {
#        origin "okx" { query = "MANA/USDT" }
#        reference { data_model = "USDT/USD" }
#      }
    }
  }

  data_model "MATIC/USD" {
    median {
      min_values = 1
      indirect {
        origin "binance" { query = "MATIC/USDT" }
        alias "USDT/USD" { # debug
          origin "bitfinex" { query = "UST/USD" }
        }
      }
      origin "coinbase" { query = "MATIC/USD" }
      origin "gemini" { query = "MATIC/USD" }
      indirect {
        origin "huobi" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
#      indirect {
#        origin "upbit" { query = "MANA/KRW" }
#        origin "openexchangerates" { query = "KRW/USD" }
#      }
    }
  }

  data_model "MKR/USD" {
    median {
      min_values = 2
      origin "bitstamp" { query = "MKR/USD" }
      origin "gemini" { query = "MKR/USD" }
    }
  }

  data_model "MKR/ETH" {
    median {
      min_values = 1
      indirect {
        origin "bitstamp" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }

      indirect {
        origin "gemini" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
    }
  }

  data_model "USDC/USD" {
    median {
      min_values = 1
      origin "gemini" { query ="USDC/USD" }
    }
  }

  data_model "USDT/USD" {
    median {
      min_values = 1
      alias "USDT/USD" {
        origin "bitfinex" { query = "UST/USD" }
      }
    }
  }
}

ghostnext {
  ethereum_key = "default"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}
