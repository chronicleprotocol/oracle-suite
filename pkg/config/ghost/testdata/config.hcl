ghost {
  ethereum_key = "key1"
  interval     = 60

  pairs = [
    "ETH/USD",
    "BTC/USD",
  ]
}

gofer {
  rpc_listen_addr = "localhost:8080"
  rpc_agent_addr  = "localhost:8081"

  origin "origin" {
    type   = "origin"
    params = {
      contract_address = "0x1234567890123456789012345678901234567890"
    }
  }

  price_model "AAA/BBB" "median" {
    source "AAA/BBB" "origin" { origin = "origin1" }
    source "AAA/BBB" "indirect" {
      source "AAA/XXX" "origin" { origin = "origin2" }
      source "XXX/BBB" "origin" { origin = "origin3" }
    }
    min_sources = 3
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
