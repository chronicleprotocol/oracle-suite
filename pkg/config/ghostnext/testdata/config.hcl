ghostnext {
  ethereum_key = "key1"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}

gofernext {
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

gofer {
  rpc_listen_addr = "localhost:8080"
  rpc_agent_addr  = "localhost:8081"

  origin "uniswapV3" {
    type   = "uniswapV3"
    params = {
      contract_address = "0x1234567890123456789012345678901234567890"
    }
  }

  price_model "AAA/BBB" "median" {
    source "AAA/BBB" "origin" { origin = "uniswapV3" }
    source "AAA/BBB" "indirect" {
      source "AAA/XXX" "origin" { origin = "uniswapV3" }
      source "XXX/BBB" "origin" { origin = "uniswapV3" }
    }
    min_sources = 3
  }
}
