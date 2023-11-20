morph {
  cache_path      = "config-cache.hcl"
  ethereum_client = "default"
  config_registry = "0x0000000000000000000000000000000000000000"
  interval        = 3600

  app {
    bin             = ""
    args            = ""
    waiting_quiting = 60
  }
}
ethereum {
  rand_keys = ["default"]

  client "ethereum" {
    rpc_urls     = ["https://eth.public-rpc.com"]
    ethereum_key = "default"
    chain_id     = 1
  }
}
