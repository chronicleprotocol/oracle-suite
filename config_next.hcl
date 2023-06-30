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

  origin "uniswapV3" {
    type = "uniswapV3"
    contracts "ethereum" {
      addresses = {
        "GNO/WETH"  = "0xf56D08221B5942C428Acc5De8f78489A97fC5599",
        "LINK/WETH" = "0xa6Cc3C2531FdaA6Ae1A3CA84c2855806728693e8",
        "MKR/USDC"  = "0xC486Ad2764D55C7dc033487D634195d6e4A6917E",
        "MKR/WETH"  = "0xe8c6c9227491C0a8156A0106A0204d881BB7E531",
        "USDC/WETH" = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640",
        "YFI/WETH"  = "0x04916039B1f59D9745Bf6E0a21f191D1e0A84287"
      }
    }
  }

  origin "sushiswap" {
    type = "sushiswap"
    contracts "ethereum" {
      addresses = {
        "YFI/WETH" = "0x088ee5007c98a9677165d78dd2109ae4a3d04d0c"
      }
    }
  }

  data_model "ETH/USD" { # debug
    alias "ETH/USD" {
      origin "uniswapV3" { query = "WETH/USDC" }
    }
  }

  data_model "GNO/ETH" { # debug
    alias "GNO/ETH" {
      origin "uniswapV3" { query = "GNO/WETH" }
    }
  }

  data_model "LINK/ETH" { # debug
    alias "LINK/ETH" {
      origin "uniswapV3" { query = "LINK/WETH" }
    }
  }

  data_model "MKR/USDC" { # debug
    alias "MKR/USDC" {
      origin "uniswapV3" { query = "MKR/USDC" }
    }
  }

  data_model "MKR/ETH" { # debug
    alias "MKR/ETH" {
      origin "uniswapV3" { query = "MKR/WETH" }
    }
  }

  data_model "YFI/ETH" { # debug
    median {
      min_values = 2
      alias "YFI/ETH" {
        origin "uniswapV3" { query = "YFI/WETH" }
      }
      alias "YFI/ETH" {
        origin "sushiswap" { query = "YFI/WETH" }
      }
    }
  }
}

ghostnext {
  ethereum_key = "default"
  interval     = 60

  data_models = [
    "BTC/USD"
  ]
}
