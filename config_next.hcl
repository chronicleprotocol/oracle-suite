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
    url  = "https://api-pub.bitfinex.com/v2/tickers?symbols=ALL"
    jq   = ".[] | select(.[0] == \"t\" + ($ucbase + $ucquote)) | {price: .[7], time: now|round, volume: .[8]}"
  }

  origin "bitstamp" {
    type = "tick_generic_jq"
    url  = "https://www.bitstamp.net/api/v2/ticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: .timestamp, volume: .volume}"
  }

  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

#  origin "curve" {} # todo, Need web3

  origin "gemini" {
    type = "tick_generic_jq"
    url  = "https://api.gemini.com/v1/pubticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: (.volume.timestamp/1000), volume: null}"
  }

  origin "huobi" {
    type = "tick_generic_jq"
    url  = "https://api.huobi.pro/market/tickers"
    jq   = ".data[] | select(.symbol == ($lcbase+$lcquote)) | {price: .close, volume: .vol, time: now|round}"
  }

#  origin "ishares" {} # todo, Need parse html

  origin "kraken" {
    type = "tick_generic_jq"
    url  = "https://api.kraken.com/0/public/Ticker?pair=$${ucbase}/$${ucquote}"
    jq   = "($ucbase + \"/\" + $ucquote) as $pair | {price: .result[$pair].c[0]|tonumber, time: now|round, volume: .result[$pair].v[0]|tonumber}"
  }

  origin "okx" {
    type = "tick_generic_jq"
    url  = "https://www.okx.com/api/v5/market/ticker?instId=$${ucbase}-$${ucquote}-SWAP"
    jq   = "{price: .data[0].last|tonumber, time: (.data[0].ts|tonumber/1000), volume: .data[0].vol24h|tonumber}"
  }

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
#      origin "binance" { query = "BTC/USD" } # Not work
      origin "bitstamp" { query = "BTC/USD" }
      origin "coinbase" { query = "BTC/USD" }
      origin "gemini" { query = "BTC/USD" }
      origin "kraken" { query = "BTC/USD" }
    }
  }

  data_model "ETH/BTC" {
    median {
      min_values = 3
      origin "binance" { query = "ETH/BTC" }
      origin "bitstamp" { query = "ETH/BTC" }
      origin "coinbase" { query = "ETH/BTC" }
      origin "gemini" { query = "ETH/BTC" }
      origin "kraken" { query = "ETH/BTC" }
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
      origin "coinbase" { query = "ETH/USD" }
      origin "gemini" { query = "ETH/USD" }
      origin "kraken" { query = "ETH/USD" }
#      origin "uniswapV3" { query = "ETH/USD" }
    }
  }

#  data_model "GNO/USD" {
#    median {
#      min_values = 3
#      indirect {
#        origin "balancerV2" { query = "ETH/GNO" }
#        reference { data_model = "ETH/USD" }
#      }
#      indirect {
#        origin "uniswapV3" { query = "GNO/ETH" }
#        reference { data_model = "ETH/USD" }
#      }
#      indirect {
#        origin "kraken" { query = "GNO/BTC" }
#        reference { data_model = "BTC/USD" }
#      }
#      indirect {
#        origin "binance" { query = "GNO/USDT" }
#        reference { data_model = "USDT/USD" }
#      }
#    }
#  }

#  data_model "IBTA/USD" {
#    origin "ishares" { query = "IBTA/USD" }
#  }

  data_model "LINK/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "LINK/BTC" }
        reference { data_model = "BTC/USD" }
      }
      origin "bitstamp" { query = "LINK/USD" }
      origin "coinbase" { query = "LINK/USD" }
      origin "gemini" { query = "LINK/USD" }
      origin "kraken" { query = "LINK/USD" }
#      indirect {
#        origin "uniswapV3" { query = "LINK/ETH" }
#        reference { data_model = "ETH/USD" }
#      }
    }
  }

  data_model "MATIC/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "MATIC/USD" }
      origin "gemini" { query = "MATIC/USD" }
      indirect {
        origin "huobi" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query = "MATIC/USD" }
    }
  }

  data_model "MKR/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "MKR/BTC" }
        reference { data_model = "BTC/USD" }
      }
      origin "bitstamp" { query = "MKR/USD" }
      origin "coinbase" { query = "MKR/USD" }
      origin "gemini" { query = "MKR/USD" }
      origin "kraken" { query = "MKR/USD" }
#      indirect {
#        origin "uniswapV3" { query = "MKR/ETH" }
#        reference { data_model = "ETH/USD" }
#      }
#      indirect {
#        origin "uniswapV3" { query = "MKR/USDC" }
#        reference { data_model = "USDC/USD" }
#      }
    }
  }

  data_model "MKR/ETH" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "MKR/BTC" }
        reference { data_model = "ETH/BTC" }
      }
      indirect {
        origin "bitstamp" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "coinbase" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "gemini" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "kraken" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
    }
  }

#  data_model "RETH/ETH" {
#    median {
#      min_values = 3
#      origin "balancerV2" { query = "RETH/ETH" }
#      indirect {
#        origin "curve" { query = "RETH/WSTETH" }
#        reference { data_model = "WSTETH/ETH" }
#      }
#      origin "rocketpool" { query = "RETH/ETH" }
#    }
#  }

#  data_model "RETH/USD" {
#    indirect {
#      reference { data_model = "RETH/ETH" }
#      reference { data_model = "ETH/USD" }
#    }
#  }

#  data_model "STETH/ETH" {
#    median {
#      min_values = 2
#      origin "balancerV2" { query = "STETH/ETH" }
#      origin "curve" { query = "STETH/ETH" }
#    }
#  }

  data_model "USDC/USD" {
    median {
      min_values = 2
      origin "coinbase" { query ="USDC/USD" } # Not work
      origin "gemini" { query ="USDC/USD" }
      origin "kraken" { query ="USDC/USD" }
    }
  }

  data_model "USDT/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "BTC/USDT" }
        reference { data_model = "BTC/USD" }
      }
      alias "USDT/USD" {
        origin "bitfinex" { query = "UST/USD" }
      }
      origin "coinbase" { query = "USDT/USD" }
      origin "kraken" { query = "USDT/USD" }
      indirect {
        origin "okx" { query = "BTC/USDT" }
        reference { data_model = "BTC/USD" }
      }
    }
  }

#  data_model "WSTETH/ETH" {
#    indirect {
#      origin "wsteth" { query = "WSTETH/STETH" }
#      reference { data_model = "STETH/ETH" }
#    }
#  }

#  data_model "WSTETH/USD" {
#    indirect {
#      reference { data_model = "WSTETH/ETH" }
#      reference { data_model = "ETH/USD" }
#    }
#  }

  data_model "YFI/USD" {
    median {
      min_values = 2
#      indirect {
#        origin "balancerV2" { query = "ETH/YFI" }
#        reference { data_model = "ETH/USD" }
#      }
      indirect {
        origin "binance" { query = "YFI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "YFI/USD" }
      origin "kraken" { query = "YFI/USD" }
      indirect {
        origin "okx" { query = "YFI/USDT" }
        reference { data_model = "USDT/USD" }
      }
#      indirect {
#        origin "sushiswap" { query = "YFI/ETH" }
#        reference { data_model = "ETH/USD" }
#      }
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
