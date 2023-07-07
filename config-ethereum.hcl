variables {
  # RPC URLs for specific blockchain clients. SOME apps are chain type aware.
  eth_rpc_urls = try(env.CFG_ETH_RPC_URLS == "" ? [] : split(",", env.CFG_ETH_RPC_URLS), [
    "https://eth.public-rpc.com"
  ])
  arb_rpc_urls = try(env.CFG_ARB_RPC_URLS == "" ? [] : split(",", env.CFG_ARB_RPC_URLS), [
    "https://arbitrum.public-rpc.com"
  ])
  opt_rpc_urls = try(env.CFG_OPT_RPC_URLS == "" ? [] : split(",", env.CFG_OPT_RPC_URLS), [
    "https://mainnet.optimism.io"
  ])
}

ethereum {
  # Labels for generating random ethereum keys on every app boot.
  # The labels are used to reference ethereum keys in other sections.
  # (optional)
  #
  # If you want to use a specific key, you can set the CFG_ETH_FROM
  # environment variable along with CFG_ETH_KEYS and CFG_ETH_PASS.
  rand_keys = try(env.CFG_ETH_FROM, "") == "" ? ["default"] : []

  dynamic "key" {
    for_each = try(env.CFG_ETH_FROM, "") == "" ? [] : [1]
    labels   = ["default"]
    content {
      address         = try(env.CFG_ETH_FROM, "")
      keystore_path   = try(env.CFG_ETH_KEYS, "")
      passphrase_file = try(env.CFG_ETH_PASS, "")
    }
  }

  dynamic "client" {
    for_each = length(var.eth_rpc_urls) > 0 ? [1] : []
    labels   = ["ethereum"]
    content {
      rpc_urls     = var.eth_rpc_urls
      chain_id     = tonumber(try(env.CFG_ETH_CHAIN_ID, "1"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.arb_rpc_urls) > 0 ? [1] : []
    labels   = ["arbitrum"]
    content {
      rpc_urls     = var.arb_rpc_urls
      chain_id     = tonumber(try(env.CFG_ARB_CHAIN_ID, "42161"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.opt_rpc_urls) > 0 ? [1] : []
    labels   = ["optimism"]
    content {
      rpc_urls     = var.opt_rpc_urls
      chain_id     = tonumber(try(env.CFG_OPT_CHAIN_ID, "10"))
      ethereum_key = "default"
    }
  }
}
