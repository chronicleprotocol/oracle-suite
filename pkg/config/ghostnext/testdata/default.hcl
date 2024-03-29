ghost {
  ethereum_key = "default"
  interval     = 60
  data_models  = ["BTC/USD", "DAI/USD", "ETH/USD", "GNO/USD", "MKR/USD", "RETH/USD", "SDAI/USD", "USDC/USD", "WSTETH/ETH", "WSTETH/USD", "DSR/RATE", "MATIC/USD", "SDAI/DAI", "SDAI/ETH", "SDAI/MATIC", "BTCUSD", "ETHUSD", "ETHBTC", "GNOUSD", "IBTAUSD", "LINKUSD", "MATICUSD", "MKRUSD", "RETHUSD", "WSTETHUSD", "YFIUSD", "AAVE/USD", "ARB/USD", "AVAX/USD", "BNB/USD", "CRV/USD", "ETH/BTC", "FRAX/USD", "GNO/ETH", "IBTA/USD", "LDO/USD", "LINK/USD", "MKR/ETH", "OP/USD", "RETH/ETH", "SNX/USD", "SOL/USD", "STETH/ETH", "STETH/USD", "UNI/USD", "USDT/USD", "WBTC/USD", "XTZ/USD", "YFI/USD"]
}
gofer {
  origin "balancerV2" {
    type = "balancerV2"

    contracts "ethereum" {
      addresses = {
        "RETH/WETH" = "0x1e19cf2d73a72ef1332c882f20534b6519be0276"
        "WETH/GNO"  = "0xf4c0dd9b82da36c07605df83c8a416f11724d88b"
      }
      references = {
        "RETH/WETH" = "0xae78736cd615f374d3085123a210448e74fc6393"
      }
    }
  }
  origin "composableBalancerV2" {
    type = "composable_balancerV2"

    contracts "ethereum" {
      addresses = {
        "GHO/LUSD"                    = "0x3fa8c89704e5d07565444009e5d9e624b40be813"
        "WSTETH/WSTETH_WETH_BPT/WETH" = "0x93d199263632a4ef4bb438f1feb99e57b4b5f0bd"
      }
    }
  }
  origin "weightedBalancerV2" {
    type = "weighted_balancerV2"

    contracts "ethereum" {
      addresses = {
        "SD/ETHX"      = "0x034e2d995b39a88ab9a532a9bf0deddac2c576ea"
        "WUSDM/WSTETH" = "0x54ca50ee86616379420cc56718e12566aa75abbe"
      }
    }
  }
  origin "binance" {
    type = "tick_generic_jq"
    url  = "https://api.binance.com/api/v3/ticker/24hr"
    jq   = ".[] | select(.symbol == ($ucbase + $ucquote)) | {price: .lastPrice, volume: .volume, time: (.closeTime / 1000)}"
  }
  origin "bitfinex" {
    type = "tick_generic_jq"
    url  = "https://api-pub.bitfinex.com/v2/tickers?symbols=ALL"
    jq   = ".[] | select(.[0] == \"t\" + ($ucbase + $ucquote) or .[0] == \"t\" + ($ucbase + \":\" + $ucquote) ) | {price: .[7], time: now|round, volume: .[8]}"
  }
  origin "bitstamp" {
    type = "tick_generic_jq"
    url  = "https://www.bitstamp.net/api/v2/ticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: .timestamp, volume: .volume}"
  }
  origin "bybit" {
    type = "tick_generic_jq"
    url  = "https://api.bybit.com/v5/market/tickers?category=spot&symbol=$${ucbase}$${ucquote}"
    jq   = "{price: .result.list[0].lastPrice|tonumber, volume: .result.list[0].volume24h|tonumber, time: (.time/1000)|round}"
  }
  origin "coinbase" {
    type = "tick_generic_jq"
    url  = "https://api.pro.coinbase.com/products/$${ucbase}-$${ucquote}/ticker"
    jq   = "{price: .price, time: .time, volume: .volume}"
  }
  origin "curve" {
    type = "curve"

    contracts "ethereum" {
      addresses = {
        "CRVUSD/SDAI"   = "0x1539c2461d7432cc114b0903f1824079bfca2c92"
        "CRVUSD/USDM"   = "0x2dabf79e16ceb92b651651f47b6e835c9db5828a"
        "DAI/USDC/USDT" = "0xbebc44782c7db0a1a60cb6fe97d0b483032ff1c7"
        "ETH/ETHX"      = "0x59ab5a5b5d617e478a2479b0cad80da7e2831492"
        "ETH/STETH"     = "0xdc24316b9ae028f1497c275eb9192a3ea0f67022"
        "FRAX/USDC"     = "0xdcef968d416a41cdac0ed8702fac8128a64241a2"
        "RETH/WSTETH"   = "0x447ddd4960d9fdbf6af9a790560d0af76795cb08"
        "USDC/CRVUSD"   = "0x4dece678ceceb27446b35c672dc7d61f30bad69e"
        "USDT/CRVUSD"   = "0x390f3595bca2df7d23783dfd126427cceb997bf4"
      }
      addresses2 = {
        "CRVUSD/WETH/CRV" = "0x4ebdf703948ddcea3b11f675b4d1fba9d2414a14"
        "USDT/WBTC/WETH"  = "0xd51a44d3fae010294c616388b506acda1bfaae46"
        "WETH/LDO"        = "0x9409280dc1e6d33ab7a8c6ec03e5763fb61772b5"
        "WETH/RETH"       = "0x0f3159811670c117c372428d4e69ac32325e4d0f"
        "WETH/YFI"        = "0xc26b89a667578ec7b3f11b2f98d6fd15c07c54ba"
        "WSTETH/ETHX"    = "0x14756a5ed229265f86990e749285bdd39fe0334f"
      }
    }
  }
  origin "degate" {
    type = "tick_generic_jq"
    url  = "https://v1-mainnet-backend.degate.com/order-book-ws-api/ticker?base_token_id=$${ucbase}&quote_token_id=$${ucquote}"
    jq   = "{price: .data.last_price|tonumber, time: now|round, volume: .data.volume|tonumber}"
  }
  origin "dsr" {
    type = "dsr"

    contracts "ethereum" {
      addresses = {
        "DSR/RATE" = "0x197e90f9fad81970ba7976f33cbd77088e5d7cf7"
      }
    }
  }
  origin "gate" {
    type = "tick_generic_jq"
    url  = "https://api.gateio.ws/api/v4/spot/tickers"
    jq   = ".[] | select(.currency_pair == ($ucbase + \"_\" + $ucquote)) | {price:.last, volume: null, time:now|round}"
  }
  origin "gemini" {
    type = "tick_generic_jq"
    url  = "https://api.gemini.com/v1/pubticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: (.volume.timestamp/1000), volume: .volume[$ucquote]|tonumber}"
  }
  origin "hitbtc" {
    type = "tick_generic_jq"
    url  = "https://api.hitbtc.com/api/2/public/ticker?symbols=$${ucbase}$${ucquote}"
    jq   = "{price: .[0].last|tonumber, time: .[0].timestamp|strptime(\"%Y-%m-%dT%H:%M:%S.%fZ\")|mktime, volume: .[0].volumeQuote|tonumber}"
  }
  origin "huobi" {
    type = "tick_generic_jq"
    url  = "https://api.huobi.pro/market/tickers"
    jq   = ".data[] | select(.symbol == ($lcbase+$lcquote)) | {price: .close, volume: .vol, time: now|round}"
  }
  origin "ishares" {
    type = "ishares"
    url  = "https://ishares.com/uk/individual/en/products/287340/ishares-treasury-bond-1-3yr-ucits-etf?switchLocale=y&siteEntryPassthrough=true"
  }
  origin "kraken" {
    type = "tick_generic_jq"
    url  = "https://api.kraken.com/0/public/Ticker?pair=$${ucbase}/$${ucquote}"
    jq   = "($ucbase + \"/\" + $ucquote) as $pair | {price: .result[$pair].c[0]|tonumber, time: now|round, volume: .result[$pair].v[0]|tonumber}"
  }
  origin "kucoin" {
    type = "tick_generic_jq"
    url  = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=$${ucbase}-$${ucquote}"
    jq   = "{price: .data.price, time: (.data.time/1000)|round, volume: null}"
  }
  origin "lido_lst" {
    type = "lido_lst"

    contracts "ethereum" {
      addresses = {
        "STETH/ERC20" = "0xae7ab96520de3a18e5e111b5eaab095312d7fe84"
      }
    }
  }
  origin "okx" {
    type = "tick_generic_jq"
    url  = "https://www.okx.com/api/v5/market/ticker?instId=$${ucbase}-$${ucquote}&instType=SPOT"
    jq   = "{price: .data[0].last|tonumber, time: (.data[0].ts|tonumber/1000), volume: .data[0].vol24h|tonumber}"
  }
  origin "pancakeswapV3" {
    type = "uniswapV3"

    contracts "ethereum" {
      addresses = {
        "RETH/WETH"   = "0x2201d2400d30bfd8172104b4ad046d019ca4e7bd"
        "WSTETH/WETH" = "0x3a1b97fc25fa45832f588ed3bfb2a0f74ddbd4f8"
      }
    }
  }
  origin "rocketpool" {
    type = "rocketpool"

    contracts "ethereum" {
      addresses = {
        "RETH/ETH" = "0xae78736cd615f374d3085123a210448e74fc6393"
      }
    }
  }
  origin "sdai" {
    type = "sdai"

    contracts "ethereum" {
      addresses = {
        "SDAI/DAI" = "0x83f20f44975d03b1b09e64809b757c47f942beea"
      }
    }
  }
  origin "sushiswap" {
    type = "sushiswap"

    contracts "ethereum" {
      addresses = {
        "DAI/WETH"  = "0xc3d03e4f041fd4cd388c549ee2a29a9e5075882f"
        "LINK/WETH" = "0xc40d16476380e4037e6b1a2594caf6a6cc8da967"
        "WBTC/WETH" = "0xceff51756c56ceffca006cd410b03ffc46dd3a58"
        "WETH/CRV"  = "0x58dc5a51fe44589beb22e8ce67720b5bc5378009"
        "YFI/WETH"  = "0x088ee5007c98a9677165d78dd2109ae4a3d04d0c"
      }
    }
  }
  origin "uniswapV2" {
    type = "uniswapV2"

    contracts "ethereum" {
      addresses = {
        "MKR/DAI"    = "0x517f9dd285e75b599234f7221227339478d0fcc8"
        "STETH/WETH" = "0x4028daac072e492d34a3afdbef0ba7e35d8b55c4"
        "YFI/WETH"   = "0x2fdbadf3c4d5a8666bc06645b8358ab803996e28"
      }
    }
  }
  origin "uniswapV3" {
    type = "uniswapV3"

    contracts "ethereum" {
      addresses = {
        "AAVE/WETH"   = "0x5ab53ee1d50eef2c1dd3d5402789cd27bb52c1bb"
        "ARB/WETH"    = "0x755e5a186f0469583bd2e80d1216e02ab88ec6ca"
        "DAI/FRAX"    = "0x97e7d56a0408570ba1a7852de36350f7713906ec"
        "DAI/USDC"    = "0x5777d92f208679db4b9778590fa3cab3ac9e2168"
        "ETHX/WETH"   = "0x1b9669b12959ad51b01fabcf01eabdfadb82f578"
        "FRAX/USDT"   = "0xc2a856c3aff2110c1171b8f942256d40e980c726"
        "GNO/WETH"    = "0xf56d08221b5942c428acc5de8f78489a97fc5599"
        "LDO/WETH"    = "0xa3f558aebaecaf0e11ca4b2199cc5ed341edfd74"
        "LINK/WETH"   = "0xa6cc3c2531fdaa6ae1a3ca84c2855806728693e8"
        "MATIC/WETH"  = "0x290a6a7460b308ee3f19023d2d00de604bcf5b42"
        "MKR/USDC"    = "0xc486ad2764d55c7dc033487d634195d6e4a6917e"
        "MKR/WETH"    = "0xe8c6c9227491c0a8156a0106a0204d881bb7e531"
        "MNT/WETH"    = "0xf4c5e0f4590b6679b3030d29a84857f226087fef"
        "SD/USDC"     = "0xc72abb13b6bdfa64770cb5b1f57bebd36a91a29e"
        "RETH/WETH"   = "0xa4e0faa58465a2d369aa21b3e42d43374c6f9613"
        "UNI/WETH"    = "0x1d42064fc4beb5f8aaf85f4617ae8b3b5b8bd801"
        "USDC/SNX"    = "0x020c349a0541d76c16f501abc6b2e9c98adae892"
        "USDC/WETH"   = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
        "WBTC/WETH"   = "0x4585fe77225b41b697c938b018e2ac67ac5a20c0"
        "WETH/CRV"    = "0x919fa96e88d67499339577fa202345436bcdaf79"
        "WSTETH/WETH" = "0x109830a1aaad605bbf02a9dfa7b0b92ec2fb7daa"
        "WUSDM/SDAI"  = "0x330b0c153c57cbca6538d143021954368ca0969f"
        "YFI/WETH"    = "0x04916039b1f59d9745bf6e0a21f191d1e0a84287"
      }
    }
  }
  origin "upbit" {
    type = "tick_generic_jq"
    url  = "https://api.upbit.com/v1/ticker?markets=$${ucquote}-$${ucbase}"
    jq   = "{price: .[0].trade_price, time: (.[0].timestamp/1000), volume: .[0].acc_trade_volume_24h}"
  }
  origin "wsteth" {
    type = "wsteth"

    contracts "ethereum" {
      addresses = {
        "WSTETH/STETH" = "0x7f39c581f595b53c5cb19bd0b3f8da6c935e2ca0"
      }
    }
  }
  data_model "AAVE/USD" {
    median {
      min_values = 4

      indirect {
        origin "binance" { query = "AAVE/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "AAVE/USD" }
      indirect {
        origin "okx" { query = "AAVE/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query = "AAVE/USD" }
      origin "bitstamp" { query = "AAVE/USD" }
      indirect {
        alias "AAVE/ETH" {
          origin "uniswapV3" { query = "AAVE/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "ARB/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "ARB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "ARB/USD" }
      origin "kraken" { query = "ARB/USD" }
      indirect {
        alias "ARB/ETH" {
          origin "uniswapV3" { query = "ARB/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "okx" { query = "ARB/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "AVAX/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "AVAX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "AVAX/USD" }
      origin "kraken" { query = "AVAX/USD" }
      origin "bitstamp" { query = "AVAX/USD" }
      indirect {
        origin "kucoin" { query = "AVAX/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "BNB/USD" {
    median {
      min_values = 2

      indirect {
        origin "binance" { query = "BNB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "kucoin" { query = "BNB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "BNB/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "BTC/USD" {
    median {
      min_values = 3

      origin "bitstamp" { query = "BTC/USD" }
      origin "coinbase" { query = "BTC/USD" }
      origin "gemini" { query = "BTC/USD" }
      origin "kraken" { query = "BTC/USD" }
    }
  }
  data_model "CRV/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "CRV/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "CRV/USD" }
      indirect {
        alias "CRV/ETH" {
          origin "uniswapV3" { query = "CRV/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      origin "kraken" { query = "CRV/USD" }
      indirect {
        alias "ETH/CRV" {
          origin "sushiswap" { query = "WETH/CRV" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "okx" { query = "CRV/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "CRVUSD/USD" {
    median {
      min_values = 2

      indirect {
        origin "curve" { query = "CRVUSD/USDC" }
        reference { data_model = "USDC/USD" }
      }
      indirect {
        origin "curve" { query = "CRVUSD/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        alias "CRVUSD/ETH" {
          origin "curve" { query = "CRVUSD/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        alias "CRVUSD/SDAI" {
          origin "curve" { query = "CRVUSD/SDAI" }
        }
        reference { data_model = "SDAI/USD" }
      }
    }
  }
  data_model "DAI/USD" {
    median {
      min_values = 5

      indirect {
        alias "DAI/USDC" {
          origin "uniswapV3" { query = "DAI/USDC" }
        }
        reference { data_model = "USDC/USD" }
      }
      indirect {
        origin "binance" { query = "USDT/DAI" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query = "DAI/USD" }
      origin "coinbase" { query = "DAI/USD" }
      origin "gemini" { query = "DAI/USD" }
      indirect {
        origin "okx" { query = "ETH/DAI" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        alias "DAI/ETH" {
          origin "sushiswap" { query = "DAI/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "curve" { query = "DAI/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "DSR/RATE" {
    origin "dsr" { query = "DSR/RATE" }
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
  data_model "ETH/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "ETH/BTC" }
        reference { data_model = "BTC/USD" }
      }
      origin "bitstamp" { query = "ETH/USD" }
      origin "coinbase" { query = "ETH/USD" }
      origin "gemini" { query = "ETH/USD" }
      origin "kraken" { query = "ETH/USD" }
      indirect {
        alias "ETH/USDC" {
          origin "uniswapV3" { query = "WETH/USDC" }
        }
        reference { data_model = "USDC/USD" }
      }
    }
  }
  data_model "ETHX/USD" {
    median {
      min_values = 3

      indirect {
        origin "curve" { query = "WSTETH/ETHX" }
        reference { data_model = "WSTETH/USD" }
      }
      indirect {
        origin "curve" { query = "ETH/ETHX" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        alias "ETHX/ETH" {
          origin "uniswapV3" { query = "ETHX/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "weightedBalancerV2" { query = "SD/ETHX" }
        reference { data_model = "SD/USD" }
      }
    }
  }
  data_model "FRAX/USD" {
    median {
      min_values = 2

      indirect {
        origin "curve" { query = "FRAX/USDC" }
        reference { data_model = "USDC/USD" }
      }
      indirect {
        origin "uniswapV3" { query = "FRAX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "uniswapV3" { query = "DAI/FRAX" }
        reference { data_model = "DAI/USD" }
      }
    }
  }
  data_model "GNO/ETH" {
    indirect {
      reference { data_model = "GNO/USD" }
      reference { data_model = "ETH/USD" }
    }
  }
  data_model "GNO/USD" {
    median {
      min_values = 2

      indirect {
        alias "GNO/ETH" {
          origin "uniswapV3" { query = "GNO/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "binance" { query = "GNO/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "GNO/USD" }
      indirect {
        alias "GNO/ETH" {
          origin "balancerV2" { query = "GNO/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "IBTA/USD" {
    origin "ishares" {
      query               = "IBTA/USD"
      freshness_threshold = 28800
      expiry_threshold    = 86400
    }
  }
  data_model "LDO/USD" {
    median {
      min_values = 4

      indirect {
        origin "binance" { query = "LDO/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "LDO/USD" }
      indirect {
        alias "LDO/ETH" {
          origin "uniswapV3" { query = "LDO/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      origin "kraken" { query = "LDO/USD" }
      indirect {
        alias "LDO/ETH" {
          origin "curve" { query = "LDO/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "LINK/USD" {
    median {
      min_values = 5

      indirect {
        origin "binance" { query = "LINK/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "LINK/USD" }
      indirect {
        alias "LINK/ETH" {
          origin "uniswapV3" { query = "LINK/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      origin "kraken" { query = "LINK/USD" }
      origin "gemini" { query = "LINK/USD" }
      origin "bitstamp" { query = "LINK/USD" }
      indirect {
        alias "LINK/ETH" {
          origin "sushiswap" { query = "LINK/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "MATIC/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "MATIC/USD" }
      indirect {
        origin "kucoin" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query = "MATIC/USD" }
      indirect {
        alias "MATIC/ETH" {
          origin "uniswapV3" { query = "MATIC/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "MKR/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "MKR/BTC" }
        reference { data_model = "BTC/USD" }
      }
      origin "bitstamp" { query = "MKR/USD" }
      origin "coinbase" { query = "MKR/USD" }
      origin "gemini" { query = "MKR/USD" }
      origin "kraken" { query = "MKR/USD" }
      indirect {
        alias "MKR/ETH" {
          origin "uniswapV3" { query = "MKR/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "uniswapV2" { query = "MKR/DAI" }
        reference { data_model = "DAI/USD" }
      }
    }
  }
  data_model "MKR/ETH" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "MKR/BTC" }
        reference { data_model = "ETH/BTC" }
      }
      indirect {
        origin "bitstamp" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "coinbase" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "gemini" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "kraken" { query = "MKR/USD" }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "MNT/USD" {
    median {
      min_values = 2

      indirect {
        alias "MNT/ETH" {
          origin "uniswapV3" { query = "MNT/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "bybit" { query = "MNT/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "OP/USD" {
    median {
      min_values = 2

      indirect {
        origin "binance" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "OP/USD" }
      indirect {
        origin "okx" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "kucoin" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "RETH/ETH" {
    median {
      min_values = 3

      alias "RETH/ETH" {
        origin "uniswapV3" { query = "RETH/WETH" }
      }
      alias "RETH/ETH" {
        origin "balancerV2" { query = "RETH/WETH" }
      }
      alias "RETH/ETH" {
        origin "pancakeswapV3" { query = "RETH/WETH" }
      }
      alias "RETH/ETH" {
        origin "curve" { query = "RETH/WETH" }
      }
      origin "rocketpool" { query = "RETH/ETH" }
    }
  }
  data_model "RETH/USD" {
    indirect {
      reference { data_model = "RETH/ETH" }
      reference { data_model = "ETH/USD" }
    }
  }
  data_model "SD/USD" {
    median {
      min_values = 3

      indirect {
        origin "gate" { query = "SD/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "SD/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "bybit" { query = "SD/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "uniswapV3" { query = "SD/USDC" }
        reference { data_model = "USDC/USD" }
      }
    }
  }
  data_model "SDAI/DAI" {
    origin "sdai" { query = "SDAI/DAI" }
  }
  data_model "SDAI/ETH" {
    indirect {
      reference { data_model = "SDAI/USD" }
      reference { data_model = "ETH/USD" }
    }
  }
  data_model "SDAI/MATIC" {
    indirect {
      reference { data_model = "SDAI/USD" }
      reference { data_model = "MATIC/USD" }
    }
  }
  data_model "SDAI/USD" {
    indirect {
      reference { data_model = "SDAI/DAI" }
      reference { data_model = "DAI/USD" }
    }
  }
  data_model "SNX/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "SNX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "SNX/USD" }
      indirect {
        origin "uniswapV3" { query = "USDC/SNX" }
        reference { data_model = "USDC/USD" }
      }
      origin "kraken" { query = "SNX/USD" }
      indirect {
        origin "okx" { query = "SNX/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "SOL/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "SOL/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "SOL/USD" }
      origin "kraken" { query = "SOL/USD" }
      origin "gemini" { query = "SOL/USD" }
      indirect {
        origin "okx" { query = "SOL/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "STETH/ETH" {
    median {
      min_values = 2

      alias "STETH/ETH" {
        origin "uniswapV2" { query = "STETH/WETH" }
      }
      origin "curve" { query = "STETH/ETH" }
    }
  }
  data_model "STETH/USD" {
    median {
      min_values = 2

      indirect {
        alias "STETH/ETH" {
          origin "uniswapV2" { query = "STETH/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "curve" { query = "STETH/ETH" }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "okx" { query = "STETH/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }
  data_model "UNI/USD" {
    median {
      min_values = 4

      indirect {
        origin "binance" { query = "UNI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "UNI/USD" }
      origin "kraken" { query = "UNI/USD" }
      origin "bitstamp" { query = "UNI/USD" }
      indirect {
        alias "UNI/ETH" {
          origin "uniswapV3" { query = "UNI/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "USDC/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "BTC/USDC" }
        reference { data_model = "BTC/USD" }
      }
      origin "kraken" { query = "USDC/USD" }
      indirect {
        origin "curve" { query = "USDC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "bitstamp" { query = "USDC/USD" }
      origin "gemini" { query = "USDC/USD" }
    }
  }
  data_model "USDM/USD" {
    median {
      min_values = 2

      indirect {
        alias "USDM/USDC" {
          origin "degate" { query = "58/2" }
        }
        reference { data_model = "USDC/USD" }
      }
      indirect {
        origin "curve" { query = "CRVUSD/USDM" }
        reference { data_model = "CRVUSD/USD" }
      }
      indirect {
        alias "USDM/WSTETH" {
          origin "weightedBalancerV2" { query = "WUSDM/WSTETH" }
        }
        reference { data_model = "WSTETH/USD" }
      }
      indirect {
        alias "USDM/SDAI" {
          origin "uniswapV3" { query = "WUSDM/SDAI" }
        }
        reference { data_model = "SDAI/USD" }
      }
    }
  }
  data_model "USDT/USD" {
    median {
      min_values = 3

      indirect {
        origin "binance" { query = "BTC/USDT" }
        reference { data_model = "BTC/USD" }
      }
      alias "USDT/USD" {
        origin "bitfinex" { query = "UST/USD" }
      }
      origin "coinbase" { query = "USDT/USD" }
      origin "kraken" { query = "USDT/USD" }
      indirect {
        origin "okx" { query = "BTC/USDT" }
        reference { data_model = "BTC/USD" }
      }
    }
  }
  data_model "WBTC/USD" {
    median {
      min_values = 3

      indirect {
        alias "WBTC/ETH" {
          origin "uniswapV3" { query = "WBTC/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "binance" { query = "WBTC/BTC" }
        reference { data_model = "BTC/USD" }
      }
      indirect {
        origin "curve" { query = "WBTC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "WBTC/USD" }
      indirect {
        alias "WBTC/ETH" {
          origin "sushiswap" { query = "WBTC/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "WSTETH/ETH" {
    median {
      min_values = 3

      alias "WSTETH/ETH" {
        origin "uniswapV3" { query = "WSTETH/WETH" }
      }
      alias "WSTETH/ETH" {
        origin "composableBalancerV2" { query = "WSTETH/WETH" }
      }
      indirect {
        origin "wsteth" { query = "WSTETH/STETH" }
        origin "curve" { query = "ETH/STETH" }
      }
      alias "WSTETH/ETH" {
        origin "pancakeswapV3" { query = "WSTETH/WETH" }
      }
    }
  }
  data_model "WSTETH/USD" {
    indirect {
      reference { data_model = "WSTETH/ETH" }
      reference { data_model = "ETH/USD" }
    }
  }
  data_model "XTZ/USD" {
    median {
      min_values = 2

      indirect {
        origin "binance" { query = "XTZ/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "XTZ/USD" }
      origin "kraken" { query = "XTZ/USD" }
      indirect {
        origin "bitfinex" { query = "XTZ/BTC" }
        reference { data_model = "BTC/USD" }
      }
    }
  }
  data_model "YFI/USD" {
    median {
      min_values = 4

      indirect {
        origin "binance" { query = "YFI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "YFI/USD" }
      indirect {
        alias "ETH/YFI" {
          origin "curve" { query = "WETH/YFI" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "okx" { query = "YFI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        alias "YFI/ETH" {
          origin "sushiswap" { query = "YFI/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        alias "YFI/ETH" {
          origin "uniswapV2" { query = "YFI/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
    }
  }
  data_model "BTCUSD" {
    reference { data_model = "BTC/USD" }
  }
  data_model "ETHUSD" {
    reference { data_model = "ETH/USD" }
  }
  data_model "ETHBTC" {
    reference { data_model = "ETH/BTC" }
  }
  data_model "GNOUSD" {
    reference { data_model = "GNO/USD" }
  }
  data_model "IBTAUSD" {
    reference { data_model = "IBTA/USD" }
  }
  data_model "LINKUSD" {
    reference { data_model = "LINK/USD" }
  }
  data_model "MATICUSD" {
    reference { data_model = "MATIC/USD" }
  }
  data_model "MKRUSD" {
    reference { data_model = "MKR/USD" }
  }
  data_model "RETHUSD" {
    reference { data_model = "RETH/USD" }
  }
  data_model "WSTETHUSD" {
    reference { data_model = "WSTETH/USD" }
  }
  data_model "YFIUSD" {
    reference { data_model = "YFI/USD" }
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
transport {
  libp2p {
    feeds              = ["0x130431b4560cd1d74a990ae86c337a33171ff3c6", "0x16655369eb59f3e1cafbcfac6d3dd4001328f747", "0x3cb645a8f10fb7b0721eabae958f77a878441cb9", "0x4b0e327c08e23dd08cb87ec994915a5375619aa2", "0x4f95d9b4d842b2e2b1d1ac3f2cf548b93fd77c67", "0x60da93d9903cb7d3ed450d4f81d402f7c4f71dd9", "0x71ecff5261baa115dcb1d9335c88678324b8a987", "0x75ef8432566a79c86bbf207a47df3963b8cf0753", "0x77eb6cf8d732fe4d92c427fcdd83142db3b742f7", "0x83e23c207a67a9f9cb680ce84869b91473403e7d", "0x8afbd9c3d794ed8df903b3468f4c4ea85be953fb", "0x8de9c5f1ac1d4d02bbfc25fd178f5daa4d5b26dc", "0x8ff6a38a1cd6a42caac45f08eb0c802253f68dfd", "0xa580bbcb1cee2bcec4de2ea870d20a12a964819e", "0xa8eb82456ed9bae55841529888cde9152468635a", "0xac8519b3495d8a3e3e44c041521cf7ac3f8f63b3", "0xc00584b271f378a0169dd9e5b165c0945b4fe498", "0xc9508e9e3ccf319f5333a5b8c825418abec688ba", "0xd09506dac64aaa718b45346a032f934602e29cca", "0xd27fa2361bc2cfb9a591fb289244c538e190684b", "0xd72ba9402e9f3ff01959d6c841ddd13615ffff42", "0xd94bbe83b4a68940839cd151478852d16b3ef891", "0xda1d2961da837891f43235fddf66bad26f41368b", "0xe6367a7da2b20ecb94a25ef06f3b551bab2682e6", "0xfbaf3a7eb4ec2962bd1847687e56aaee855f5d00", "0xfeed00aa3f0845afe52df9ecfe372549b74c69d2"]
    listen_addrs       = ["/ip4/0.0.0.0/tcp/8000"]
    bootstrap_addrs    = ["/dns/spire-bootstrap1.chroniclelabs.io/tcp/8000/p2p/12D3KooWFYkJ1SghY4KfAkZY9Exemqwnh4e4cmJPurrQ8iqy2wJG", "/dns/spire-bootstrap2.chroniclelabs.io/tcp/8000/p2p/12D3KooWD7eojGbXT1LuqUZLoewRuhNzCE2xQVPHXNhAEJpiThYj"]
    direct_peers_addrs = []
    blocked_addrs      = []
    ethereum_key       = "default"
  }
}
