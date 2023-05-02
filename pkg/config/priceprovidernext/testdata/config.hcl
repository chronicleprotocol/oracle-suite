origin "coinbase" {
  origin = "generic_jq"
  url    = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
  jq     = "{price: .price, time: .time, volume: .volume}"
}

origin "binance" {
  origin = "generic_jq"
  url    = "https://api.binance.com/api/v3/ticker/24hr"
  jq     = ".[] | select(.symbol == ($$ucbase + $$ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
}

price_model "primary" "BTC/USD" {
  median "BTC/USD" {
    origin "coinbase" "BTC/USD" { }
    origin "coinbase" "BTC/USD" { }
    indirect "BTC/USD" {
      origin "coinbase" "BTC/USDC" { }
      origin "coinbase" "USDC/USD" { }
    }
    origin "coinbase" "BTC/USD" { }
    min_sources = 2
  }
}

