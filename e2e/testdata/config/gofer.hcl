ethereum {
  client "default" {
    rpc_urls = ["http://127.0.0.1:8080"]
    chain_id = 1
  }
}

gofer {
  origin "bitstamp" {
    type   = "bitstamp"
    params = {
      url = "http://127.0.0.1:8080"
    }
  }

  origin "kraken" {
    type   = "kraken"
    params = {
      url = "http://127.0.0.1:8080"
    }
  }
  origin "coinbasepro" {
    type   = "coinbasepro"
    params = {
      url = "http://127.0.0.1:8080"
    }
  }

  origin "balancerV2" {
    type   = "balancerV2"
    params = {
      ethereum_client = "default"
      symbol_aliases  = {
        "ETH" = "WETH"
      }
      contracts = {
        "WETH/GNO"      = "0xF4C0DD9B82DA36C07605df83c8a416F11724d88b",
        "Ref:RETH/WETH" = "0xae78736Cd615f374D3085123A210448E74Fc6393",
        "RETH/WETH"     = "0x1E19CF2D73a72Ef1332C882F20534B6519Be0276",
        "STETH/WETH"    = "0x32296969ef14eb0c6d29669c550d4a0449130230",
        "WETH/YFI"      = "0x186084ff790c65088ba694df11758fae4943ee9e"
      }
    }
  }

  origin "gemini" {
    type   = "gemini"
    params = {
      url = "http://127.0.0.1:8080"
    }
  }

  origin "binance_us" {
    type   = "binance"
    params = {
      url = "http://127.0.0.1:8080"
    }
  }

  origin "bittrex" {
    type   = "bittrex"
    params = {
      url            = "http://127.0.0.1:8080"
      symbol_aliases = {
        "REP" = "REPV2"
      }
    }
  }

  origin "curve" {
    type   = "curve"
    params = {
      ethereum_client = "default"
      contracts       = {
        "RETH/WSTETH" = "0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08",
        "ETH/STETH"   = "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"
      }
    }
  }

  origin "ishares" {
    type = "ishares"
  }

  origin "poloniex" {
    type   = "poloniex"
    params = {
      url            = "http://127.0.0.1:8080"
      symbol_aliases = {
        "REP" = "REPV2"
      }
    }
  }

  origin "rocketpool" {
    type   = "rocketpool"
    params = {
      ethereum_client = "default"
      contracts       = {
        "RETH/ETH" = "0xae78736Cd615f374D3085123A210448E74Fc6393"
      }
    }
  }

  origin "sushiswap" {
    type   = "sushiswap"
    params = {
      symbol_aliases = {
        "ETH" = "WETH",
        "BTC" = "WBTC",
        "USD" = "USDC"
      }
      contracts = {
        "YFI/WETH" = "0x088ee5007c98a9677165d78dd2109ae4a3d04d0c"
      }
    }
  }

  origin "uniswap" {
    type   = "uniswap"
    params = {
      symbol_aliases = {
        "ETH" = "WETH",
        "BTC" = "WBTC",
        "USD" = "USDC"
      }
      contracts = {
        "WETH/USDC" = "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc",
        "LEND/WETH" = "0xab3f9bf1d81ddb224a2014e98b238638824bcf20",
        "LRC/WETH"  = "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70",
        "PAXG/WETH" = "0x9c4fe5ffd9a9fc5678cfbd93aa2d4fd684b67c4c",
        "BAL/WETH"  = "0xa70d458a4d9bc0e6571565faee18a48da5c0d593",
        "YFI/WETH"  = "0x2fdbadf3c4d5a8666bc06645b8358ab803996e28"
      }
    }
  }

  origin "uniswapV3" {
    type   = "uniswapV3"
    params = {
      symbol_aliases = {
        "BTC" = "WBTC",
        "ETH" = "WETH",
        "USD" = "USDC"
      }
      contracts = {
        "GNO/WETH"  = "0xf56d08221b5942c428acc5de8f78489a97fc5599",
        "LINK/WETH" = "0xa6cc3c2531fdaa6ae1a3ca84c2855806728693e8",
        "MKR/USDC"  = "0xc486ad2764d55c7dc033487d634195d6e4a6917e",
        "MKR/WETH"  = "0xe8c6c9227491c0a8156a0106a0204d881bb7e531",
        "USDC/WETH" = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
        "YFI/WETH"  = "0x04916039b1f59d9745bf6e0a21f191d1e0a84287"
      }
    }
  }

  origin "wsteth" {
    type   = "wsteth"
    params = {
      ethereum_client = "default"
      contracts       = {
        "WSTETH/STETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"
      }
    }
  }


  price_model "BTC/USD" "median" {
    source "BTC/USD" "origin" { origin = "binance_us" }
    source "BTC/USD" "origin" { origin = "bitstamp" }
    source "BTC/USD" "origin" { origin = "coinbasepro" }
    source "BTC/USD" "origin" { origin = "gemini" }
    source "BTC/USD" "origin" { origin = "kraken" }
    min_sources = 3
  }

  price_model "ETH/BTC" "median" {
    source "ETH/BTC" "origin" { origin = "binance_us" }
    source "ETH/BTC" "origin" { origin = "bitstamp" }
    source "ETH/BTC" "origin" { origin = "coinbasepro" }
    source "ETH/BTC" "origin" { origin = "gemini" }
    source "ETH/BTC" "origin" { origin = "kraken" }
    min_sources = 3
  }

  price_model "RETH/ETH" "median" {
    source "RETH/ETH" "origin" { origin = "balancerV2" }
    source "RETH/ETH" "indirect" {
      source "RETH/WSTETH" "origin" { origin = "curve" }
      source "WSTETH/ETH" "origin" { origin = "." }
    }
    source "RETH/ETH" "origin" { origin = "rocketpool" }
    min_sources = 3
  }
  hook "RETH/ETH" {
    post_price = {
      ethereum_client  = "default"
      circuit_contract = "0xa3105dee5ec73a7003482b1a8968dc88666f3589"
    }
  }

  price_model "WSTETH/ETH" "indirect" {
    source "WSTETH/STETH" "origin" { origin = "wsteth" }
    source "STETH/ETH" "origin" { origin = "." }
    min_sources = 1
  }

  price_model "STETH/ETH" "median" {
    source "STETH/ETH" "origin" { origin = "balancerV2" }
    source "STETH/ETH" "origin" { origin = "curve" }
    min_sources = 2
  }
}
