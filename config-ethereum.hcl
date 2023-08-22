variables {
  # RPC URLs for specific blockchain clients. SOME apps are chain type aware.
  eth_rpc_urls = explode(",", env("CFG_ETH_RPC_URLS", env("CFG_SPECTRE_TARGET_NETWORK", "") == "" ? "https://eth.public-rpc.com,https://cloudflare-eth.com,https://ethereum.publicnode.com" : ""))
  arb_rpc_urls = explode(",", env("CFG_ARB_RPC_URLS", ""))
  opt_rpc_urls = explode(",", env("CFG_OPT_RPC_URLS", ""))

  # Configures the spectre target network.
  spectre_target_network = env("CFG_SPECTRE_TARGET_NETWORK", "")
  spectre_rpc_urls       = explode(",", env("CFG_RPC_URLS", "https://rpc2.sepolia.org,https://rpc-sepolia.rockx.com,https://rpc.sepolia.ethpandaops.io"))
  spectre_chain_id       = tonumber(env("CFG_CHAIN_ID", "11155111"))
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

  # Spectre client
  dynamic "client" {
    for_each = var.spectre_target_network == "" || length(var.spectre_rpc_urls) == 0 ? [] : [1]
    labels   = ["spectre"]
    content {
      rpc_urls     = var.spectre_rpc_urls
      chain_id     = var.spectre_chain_id
      ethereum_key = "default"
    }
  }
}
