# Test config for the gofernext apps. Not ready for production use.

gofernext {
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

  origin "ishares" {
    type = "ishares"
    url = "https://ishares.com/uk/individual/en/products/287340/ishares-treasury-bond-1-3yr-ucits-etf?switchLocale=y&siteEntryPassthrough=true"
  }

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

  origin "upbit" {
    type = "tick_generic_jq"
    url  = "https://api.upbit.com/v1/ticker?markets=$${ucquote}-$${ucbase}"
    jq   = "{price: .[0].trade_price, time: (.[0].timestamp/1000), volume: .[0].acc_trade_volume_24h}"
  }

  origin "wsteth" {
    type = "wsteth"
    contracts "ethereum" {
      addresses = {
        "WSTETH/STETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"
      }
    }
  }

  data_model "BTC/USD" {
    median {
      min_values = 3
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
    }
  }

  data_model "IBTA/USD" {
    origin "ishares" { query = "IBTA/USD" }
  }

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

  data_model "USDC/USD" {
    median {
      min_values = 2
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

  data_model "YFI/USD" {
    median {
      min_values = 2
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
    }
  }
}
