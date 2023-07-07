spectre {
  interval = 60

  median {
    ethereum_client = "client1"
    contract_addr   = "0x1234567890123456789012345678901234567890"
    pair            = "BTCUSD"
    spread          = 1
    expiration      = 300
  }

  median {
    ethereum_client = "client1"
    contract_addr   = "0x2345678901234567890123456789012345678901"
    pair            = "ETHUSD"
    spread          = 3
    expiration      = 400
  }

  feeds = []
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
