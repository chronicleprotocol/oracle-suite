ethereum {
  rand_keys = ["key1"]

  client "client1" {
    rpc_urls     = ["https://rpc1.example"]
    chain_id     = 1
    ethereum_key = "key1"
  }
}

morph {
  cache_path = "config-cache.hcl"
  ethereum_client = "default"
  config_registry = "0x1234567890123456789012345678901234567890"
  interval = 3600
}
