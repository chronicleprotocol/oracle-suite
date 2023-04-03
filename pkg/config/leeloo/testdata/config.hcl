leeloo {
  ethereum_key = "key1"

  teleport_evm {
    ethereum_client     = "client1"
    interval            = 60
    prefetch_period     = 120
    block_confirmations = 3
    block_limit         = 100
    replay_after        = [600, 1200]
    contract_addrs      = ["0x1234567890123456789012345678901234567890", "0x2345678901234567890123456789012345678901"]
  }

  teleport_starknet {
    sequencer       = "http://localhost:8080"
    interval        = 60
    prefetch_period = 120
    replay_after    = [600, 1200]
    contract_addrs  = ["0x3456789012345678901234567890123456789012", "0x4567890123456789012345678901234567890123"]
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
