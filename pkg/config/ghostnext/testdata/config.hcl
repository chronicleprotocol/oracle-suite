ghost {
  ethereum_key = "key1"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}

gofer {
  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

  data_model "BTC/USD" {
    origin "coinbase" { query = "BTC/USD" }
  }
}

ethereum {
  rand_keys = ["key1"]

  client "client1" {
    rpc_urls     = ["https://rpc1.example"]
    chain_id     = 1
    ethereum_key = "key1"
  }
}

transport {
  libp2p {
    feeds             = ["0x1234567890123456789012345678901234567890"]
    listen_addrs      = ["/ip4/0.0.0.0/tcp/6000"]
    disable_discovery = false
    ethereum_key      = "key1"
  }
}

morph {
  cache_path = "cache.hcl"
  ethereum_client = "client1"
  config_registry = "0x2345678901234567890123456789012345678901"
  interval = 3600
}