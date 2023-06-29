# Test config for the gofernext and ghostnext apps. Not ready for production use.

ethereum {
  # Labels for generating random ethereum keys anew on every app boot.
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

  client "ethereum" {
    rpc_urls = try(env.CFG_ETH_RPC_URLS == "" ? [] : split(",", env.CFG_ETH_RPC_URLS), [
      "https://eth.public-rpc.com"
    ])
    chain_id     = tonumber(try(env.CFG_ETH_CHAIN_ID, "1"))
    ethereum_key = "default"
  }

  client "arbitrum" {
    rpc_urls = try(env.CFG_ETH_ARB_RPC_URLS == "" ? [] : split(",", env.CFG_ETH_ARB_RPC_URLS), [
      "https://arbitrum.public-rpc.com"
    ])
    chain_id     = tonumber(try(env.CFG_ETH_ARB_CHAIN_ID, "42161"))
    ethereum_key = "default"
  }

  client "optimism" {
    rpc_urls = try(env.CFG_ETH_OPT_RPC_URLS == "" ? [] : split(",", env.CFG_ETH_OPT_RPC_URLS), [
      "https://mainnet.optimism.io"
    ])
    chain_id     = tonumber(try(env.CFG_ETH_OPT_CHAIN_ID, "10"))
    ethereum_key = "default"
  }
}

gofernext {
  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

  origin "binance" {
    type = "tick_generic_jq"
    url  = "https://api.binance.com/api/v3/ticker/24hr"
    jq   = ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
  }

  origin "wsteth" {
    type = "wsteth"
    contracts "ethereum" {
      addresses = {
        "WSTETH/STETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"
      }
    }
  }

  data_model "BTC/USD" {
    origin "coinbase" { query = "BTC/USD" }
  }

  data_model "WSTETH/STETH" { # debug
    origin "wsteth" { query = "WSTETH/STETH" }
  }
}

ghostnext {
  ethereum_key = "default"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}
