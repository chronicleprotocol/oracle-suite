# Test config for the gofernext and ghostnext apps. Not ready for production use.

gofernext {
  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

  origin "binance" {
    type = "tick_generic_jq"
    url  = "https://api.binance.com/api/v3/ticker/24hr"
    jq   = ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
  }

  origin "ishares" {
    type = "ishares"
    url = "https://ishares.com/uk/individual/en/products/287340/ishares-treasury-bond-1-3yr-ucits-etf?switchLocale=y&siteEntryPassthrough=true"
  }

  data_model "BTC/USD" {
    origin "coinbase" { query = "BTC/USD" }
  }

#  data_model "IBTA/USD" { # debug
#    origin "ishares" { query = "IBTA/USD" }
#  }
}

ghostnext {
  ethereum_key = "default"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}
