variables {
  chain_rpc_urls = explode(",", env("CFG_CHAIN_RPC_URLS", env("CFG_RPC_URLS", "")))

  # RPC URLs for specific blockchain clients. SOME apps are chain type aware.
  eth_rpc_urls = explode(",", env("CFG_ETH_RPC_URLS", env("ETH_RPC_URL", "")))
  arb_rpc_urls = explode(",", env("CFG_ARB_RPC_URLS", ""))
  opt_rpc_urls = explode(",", env("CFG_OPT_RPC_URLS", ""))
}

ethereum {
  # Labels for generating random ethereum keys on every app boot.
  # The labels are used to reference ethereum keys in other sections.
  # (optional)
  #
  # If you want to use a specific key, you can set the CFG_ETH_FROM
  # environment variable along with CFG_ETH_KEYS and CFG_ETH_PASS.
  rand_keys = env("CFG_ETH_FROM", "") == "" ? ["default"] : []

  dynamic "key" {
    for_each = env("CFG_ETH_FROM", "") == "" ? [] : [1]
    labels   = ["default"]
    content {
      address         = env("CFG_ETH_FROM", "")
      keystore_path   = env("CFG_ETH_KEYS", "")
      passphrase_file = env("CFG_ETH_PASS", "")
    }
  }

  dynamic "client" {
    for_each = length(var.chain_rpc_urls) == 0 ? [] : [1]
    labels   = ["default"]
    content {
      rpc_urls                    = var.chain_rpc_urls
      chain_id                    = tonumber(env("CFG_CHAIN_ID", "0"))
      ethereum_key                = "default"
      tx_type                     = env("CFG_CHAIN_TX_TYPE", "eip1559")
      gas_priority_fee_multiplier = tonumber(env("CFG_CHAIN_GAS_FEE_MULTIPLIER", "1"))
      gas_fee_multiplier          = tonumber(env("CFG_CHAIN_GAS_PRIORITY_FEE_MULTIPLIER", "1"))
      max_gas_fee                 = tonumber(env("CFG_CHAIN_MAX_GAS_FEE", "0"))
      max_gas_priority_fee        = tonumber(env("CFG_CHAIN_MAX_GAS_PRIORITY_FEE", "0"))
      max_gas_limit               = tonumber(env("CFG_CHAIN_MAX_GAS_LIMIT", "0"))
    }
  }
  dynamic "client" {
    for_each = length(var.eth_rpc_urls) == 0 ? [] : [1]
    labels   = ["ethereum"]
    content {
      rpc_urls     = var.eth_rpc_urls
      chain_id     = tonumber(env("CFG_ETH_CHAIN_ID", "1"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.arb_rpc_urls) == 0 ? [] : [1]
    labels   = ["arbitrum"]
    content {
      rpc_urls     = var.arb_rpc_urls
      chain_id     = tonumber(env("CFG_ARB_CHAIN_ID", "42161"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.opt_rpc_urls) == 0 ? [] : [1]
    labels   = ["optimism"]
    content {
      rpc_urls     = var.opt_rpc_urls
      chain_id     = tonumber(env("CFG_OPT_CHAIN_ID", "10"))
      ethereum_key = "default"
    }
  }
}
