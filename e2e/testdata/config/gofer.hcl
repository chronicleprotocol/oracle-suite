ethereum {
  client "ethereum" {
    rpc_urls = ["http://127.0.0.1:8080"]
    chain_id = 1
  }
}

gofer {
  origin "balancerV2" {
    type = "balancerV2"
    contracts "ethereum" {
      addresses = {
        "WETH/GNO"    = "0xF4C0DD9B82DA36C07605df83c8a416F11724d88b" # WeightedPool2Tokens
        "RETH/WETH"   = "0x1E19CF2D73a72Ef1332C882F20534B6519Be0276" # MetaStablePool
        "WSTETH/WETH" = "0x32296969ef14eb0c6d29669c550d4a0449130230" # MetaStablePool
      }
      references = {
        "RETH/WETH"   = "0xae78736Cd615f374D3085123A210448E74Fc6393" # token0 of RETH/WETH
        "WSTETH/WETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0" # token0 of WSTETH/WETH
      }
    }
  }

  origin "binance" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/api/v3/ticker/24hr"
    jq   = ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
  }

  origin "bitstamp" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/api/v2/ticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: .timestamp, volume: .volume}"
  }

  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }

  origin "curve" {
    type = "curve"
    contracts "ethereum" {
      addresses = {
        # int256, stableswap
        "RETH/WSTETH"   = "0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08",
        "ETH/STETH"     = "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022",
        "DAI/USDC/USDT" = "0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7",
        "FRAX/USDC"     = "0xDcEF968d416a41Cdac0ED8702fAC8128A64241A2",
      }
      addresses2 = {
        # uint256, cryptoswap
        "WETH/LDO"       = "0x9409280DC1e6D33AB7A8C6EC03e5763FB61772B5",
        "USDT/WBTC/WETH" = "0xD51a44d3FaE010294C616388b506AcdA1bfAAE46",
        "WETH/YFI"       = "0xC26b89A667578ec7b3f11b2F98d6Fd15C07C54ba",
        "WETH/RETH"      = "0x0f3159811670c117c372428D4E69AC32325e4D0F"
      }
    }
  }

  origin "kraken" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/0/public/Ticker?pair=$${ucbase}/$${ucquote}"
    jq   = "($ucbase + \"/\" + $ucquote) as $pair | {price: .result[$pair].c[0]|tonumber, time: now|round, volume: .result[$pair].v[0]|tonumber}"
  }

  origin "gemini" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/v1/pubticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: (.volume.timestamp/1000), volume: .volume[$ucquote]|tonumber}"
  }

  origin "rocketpool" {
    type = "rocketpool"
    contracts "ethereum" {
      addresses = {
        "RETH/ETH" = "0xae78736Cd615f374D3085123A210448E74Fc6393"
      }
    }
  }

  origin "sushiswap" {
    type = "sushiswap"
    contracts "ethereum" {
      addresses = {
        "YFI/WETH"  = "0x088ee5007c98a9677165d78dd2109ae4a3d04d0c",
        "WETH/CRV"  = "0x58Dc5a51fE44589BEb22E8CE67720B5BC5378009",
        "DAI/WETH"  = "0xC3D03e4F041Fd4cD388c549Ee2A29a9E5075882f",
        "WBTC/WETH" = "0xCEfF51756c56CeFFCA006cD410B03FFC46dd3a58",
        "LINK/WETH" = "0xC40D16476380e4037e6b1A2594cAF6a6cc8Da967"
      }
    }
  }

  origin "uniswapV2" {
    type = "uniswapV2"
    contracts "ethereum" {
      addresses = {
        "STETH/WETH" = "0x4028DAAC072e492d34a3Afdbef0ba7e35D8b55C4",
        "MKR/DAI"    = "0x517F9dD285e75b599234F7221227339478d0FcC8",
        "YFI/WETH"   = "0x2fDbAdf3C4D5A8666Bc06645B8358ab803996E28"
      }
    }
  }

  origin "uniswapV3" {
    type = "uniswapV3"
    contracts "ethereum" {
      addresses = {
        "GNO/WETH"    = "0xf56D08221B5942C428Acc5De8f78489A97fC5599",
        "LINK/WETH"   = "0xa6Cc3C2531FdaA6Ae1A3CA84c2855806728693e8",
        "MKR/USDC"    = "0xC486Ad2764D55C7dc033487D634195d6e4A6917E",
        "MKR/WETH"    = "0xe8c6c9227491C0a8156A0106A0204d881BB7E531",
        "USDC/WETH"   = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640",
        "YFI/WETH"    = "0x04916039B1f59D9745Bf6E0a21f191D1e0A84287",
        "AAVE/WETH"   = "0x5aB53EE1d50eeF2C1DD3d5402789cd27bB52c1bB",
        "WETH/CRV"    = "0x919Fa96e88d67499339577Fa202345436bcDaf79",
        "DAI/USDC"    = "0x5777d92f208679db4b9778590fa3cab3ac9e2168",
        "FRAX/USDT"   = "0xc2A856c3afF2110c1171B8f942256d40E980C726",
        "GNO/WETH"    = "0xf56D08221B5942C428Acc5De8f78489A97fC5599",
        "LDO/WETH"    = "0xa3f558aebAecAf0e11cA4b2199cC5Ed341edfd74",
        "UNI/WETH"    = "0x1d42064Fc4Beb5F8aAF85F4617AE8b3b5B8Bd801",
        "WBTC/WETH"   = "0x4585FE77225b41b697C938B018E2Ac67Ac5a20c0",
        "USDC/SNX"    = "0x020C349A0541D76C16F501Abc6B2E9c98AdAe892",
        "ARB/WETH"    = "0x755E5A186F0469583bd2e80d1216E02aB88Ec6ca",
        "DAI/FRAX"    = "0x97e7d56A0408570bA1a7852De36350f7713906ec",
        "WSTETH/WETH" = "0x109830a1AAaD605BbF02a9dFA7B0B92EC2FB7dAa",
        "MATIC/WETH"  = "0x290A6a7460B308ee3F19023D2D00dE604bcf5B42"
      }
    }
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
    median {
      min_values = 3
      origin "binance" { query = "BTC/USD" }
      origin "bitstamp" { query = "BTC/USD" }
      origin "coinbase" { query = "BTC/USD" }
      origin "gemini" { query = "BTC/USD" }
      origin "kraken" { query = "BTC/USD" }
    }
  }

  data_model "ETH/BTC" {
    median {
      min_values = 3
      origin "binance" { query = "ETH/BTC" }
      origin "bitstamp" { query = "ETH/BTC" }
      origin "coinbase" { query = "ETH/BTC" }
      origin "gemini" { query = "ETH/BTC" }
      origin "kraken" { query = "ETH/BTC" }
    }
  }

  data_model "RETH/ETH" {
    median {
      min_values = 3
      alias "RETH/ETH" {
        origin "balancerV2" { query = "RETH/WETH" }
      }
      indirect {
        origin "curve" { query = "RETH/WSTETH" }
        reference { data_model = "WSTETH/ETH" }
      }
      origin "rocketpool" { query = "RETH/ETH" }
    }
  }

  data_model "RETH/WETH" {
    origin "balancerV2" { query = "RETH/WETH" }
  }

  data_model "RETH/WSTETH" {
    origin "curve" { query = "RETH/WSTETH" }
  }

  data_model "WSTETH/ETH" {
    median {
      min_values = 2
      alias "WSTETH/ETH" {
        origin "uniswapV3" { query = "WSTETH/WETH" }
      }
      indirect {
        origin "wsteth" { query = "WSTETH/STETH" }
        reference { data_model = "STETH/ETH" }
      }
    }
  }

  data_model "WSTETH/STETH" {
    origin "wsteth" { query = "WSTETH/STETH" }
  }

  data_model "STETH/ETH" {
    median {
      min_values = 2
      alias "STETH/ETH" {
        origin "balancerV2" { query = "WSTETH/WETH" }
      }
      origin "curve" { query = "STETH/ETH" }
    }
  }
}
