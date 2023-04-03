spire {
  ethereum_key    = "key1"
  rpc_listen_addr = "127.0.0.1:9101"
  rpc_agent_addr  = "127.0.0.1:9101"
  pairs           = [
    "BTCUSD",
    "ETHBTC",
  ]
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
