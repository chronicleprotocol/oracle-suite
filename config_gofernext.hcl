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
}
