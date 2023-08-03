# Test config for the gofernext apps. Not ready for production use.

gofernext {
  origin "balancerV2" {
    type = "balancerV2"
    contracts "ethereum" {
      addresses = {
        "WETH/GNO" = "0xF4C0DD9B82DA36C07605df83c8a416F11724d88b" # WeightedPool2Tokens
        "RETH/WETH" = "0x1E19CF2D73a72Ef1332C882F20534B6519Be0276" # MetaStablePool
        "STETH/WETH" = "0x32296969ef14eb0c6d29669c550d4a0449130230" # MetaStablePool
        "WETH/YFI" = "0x186084ff790c65088ba694df11758fae4943ee9e" # WeightedPool2Tokens
      }
      references = {
        "RETH/WETH" = "0xae78736Cd615f374D3085123A210448E74Fc6393" # token0 of RETH/WETH
        "STETH/WETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0" # token0 of STETH/WETH
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
    jq   = ".[] | select(.[0] == \"t\" + ($ucbase + $ucquote)) | {price: .[7], time: now|round, volume: .[8]}"
  }

  origin "bitstamp" {
    type = "tick_generic_jq"
    url  = "https://www.bitstamp.net/api/v2/ticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: .timestamp, volume: .volume}"
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
        "RETH/WSTETH" = "0x447Ddd4960d9fdBF6af9a790560d0AF76795CB08",
        "ETH/STETH"   = "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022",
        "CRV/WETH" = "0x919Fa96e88d67499339577Fa202345436bcDaf79",
        "DAI/USDC" = "0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7",
        "USDT/USDC" = "0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7",
        "WBTC/WETH" = "0xD51a44d3FaE010294C616388b506AcdA1bfAAE46"
      }
    }
  }

  origin "gemini" {
    type = "tick_generic_jq"
    url  = "https://api.gemini.com/v1/pubticker/$${lcbase}$${lcquote}"
    jq   = "{price: .last, time: (.volume.timestamp/1000), volume: .volume[$ucquote]|tonumber}"
  }

  origin "hitbtc" {
    type = "tick_generic_jq"
    url  = "https://api.hitbtc.com/api/2/public/ticker?symbols=$${ucbase}$${ucquote}"
    jq   = "{price: .[0].last|tonumber, time: .[0].timestamp|strptime(\"%Y-%m-%dT%H:%M:%S.%jZ\")|mktime, volume: .[0].volumeQuote|tonumber}"
  }

  origin "huobi" {
    type = "tick_generic_jq"
    url  = "https://api.huobi.pro/market/tickers"
    jq   = ".data[] | select(.symbol == ($lcbase+$lcquote)) | {price: .close, volume: .vol, time: now|round}"
  }

  origin "ishares" {
    type = "ishares"
    url = "https://ishares.com/uk/individual/en/products/287340/ishares-treasury-bond-1-3yr-ucits-etf?switchLocale=y&siteEntryPassthrough=true"
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

  origin "okx" {
    type = "tick_generic_jq"
    url  = "https://www.okx.com/api/v5/market/ticker?instId=$${ucbase}-$${ucquote}-SWAP"
    jq   = "{price: .data[0].last|tonumber, time: (.data[0].ts|tonumber/1000), volume: .data[0].vol24h|tonumber}"
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
        "YFI/WETH" = "0x088ee5007c98a9677165d78dd2109ae4a3d04d0c"
      }
    }
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
        "YFI/WETH"  = "0x04916039B1f59D9745Bf6E0a21f191D1e0A84287",
        "USDC/WETH"  = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640",
        "DAI/WETH"  = "0x60594a405d53811d3BC4766596EFD80fd545A270",
        "UNI/WETH"  = "0x1d42064Fc4Beb5F8aAF85F4617AE8b3b5B8Bd801",
        "USDC/WETH"  = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640",
        "WBTC/WETH"  = "0x4585FE77225b41b697C938B018E2Ac67Ac5a20c0"
      }
    }
  }

  origin "uniswapV3Arb" {
    type = "uniswapV3"
    contracts "arbitrum" {
      addresses = {
        "WETH/ARB" = "0xC6F780497A95e246EB9449f5e4770916DCd6396A"
      }
    }
  }

  origin "uniswapV3Op" {
    type = "uniswapV3"
    contracts "optimism" {
      addresses = {
        "OP/WETH" = "0xFC1f3296458F9b2a27a0B91dd7681C4020E09D05",
        "SNX/WETH" = "0x0392b358CE4547601BEFa962680BedE836606ae2"
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
        "WSTETH/STETH" = "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"
      }
    }
  }

  data_model "AAVE/USD" {
    median {
      min_values = 2
      indirect {
        origin "binance" { query = "AAVE/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "AAVE/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "hitbtc" { query = "AAVE/USD" }
      indirect {
        origin "kucoin" { query = "AAVE/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "AAVE/USD" }
    }
  }

  data_model "ARB/USD" {
    median {
      min_values = 3
#      indirect {
#        alias "ETH/ARB" {
#          origin "uniswapV3Arb" { query = "WETH/ARB" }
#        }
#        reference { data_model = "ETH/USD" }
#      }
      origin "coinbase" { query = "ARB/USD" }
      indirect {
        origin "binance" { query = "ARB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "huobi" { query = "ARB/USDT" }
        reference { data_model = "USDT/USD" }
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
      origin "coinbase" { query = "AVAX/USD" }
      origin "kraken" { query = "AVAX/USD" }
      indirect {
        origin "binance" { query = "AVAX/BTC" }
        reference { data_model = "BTC/USD" }
      }
      indirect {
        origin "okx" { query = "AVAX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "hitbtc" { query = "AVAX/USDT" }
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
        origin "huobi" { query = "BNB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "kucoin" { query = "BNB/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "hitbtc" { query = "BNB/BTC" }
        reference { data_model = "BTC/USD" }
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
#
#  data_model "CRV/USD" {
#    median {
#      min_values = 2
#      indirect {
#        origin "binance" { query = "CRV/USDT" }
#        reference { data_model = "USDT/USD" }
#      }
#      indirect {
#        origin "hitbtc" { query = "CRV/USDT" }
#        reference { data_model = "USDT/USD" }
#      }
##      indirect {
##        alias "CRV/ETH" {
##          origin "curve" { query = "CRV/WETH" }
##        }
##        reference { data_model = "ETH/USD" }
##      }
#    }
#  }
#
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
      origin "hitbtc" { query = "ETH/USD" }
    }
  }

  data_model "DAI/USD" {
    median {
      min_values = 3
      indirect {
        alias "DAI/ETH" {
          origin "uniswapV3" { query = "DAI/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
#      indirect {
#        origin "curve" { query = "DAI/USDC" }
#        reference { data_model = "USDC/USD" }
#      }
      indirect {
        origin "huobi" { query = "DAI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "DAI/USD" }
      origin "kraken" { query = "DAI/USD" }
    }
  }

  data_model "GNO/USD" {
    median {
      min_values = 3
      indirect {
        alias "ETH/GNO" {
          origin "balancerV2" { query = "WETH/GNO" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        alias "GNO/ETH" {
          origin "uniswapV3" { query = "GNO/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "kraken" { query = "GNO/BTC" }
        reference { data_model = "BTC/USD" }
      }
      indirect {
        origin "binance" { query = "GNO/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }

  data_model "IBTA/USD" {
    origin "ishares" { query = "IBTA/USD" }
  }

  data_model "LDO/USD" {
    median {
      min_values = 2
      indirect {
        origin "binance" { query = "LDO/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "LDO/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "LDO/USD" }
    }
  }

  data_model "LINK/USD" {
    median {
      min_values = 4
      indirect {
        origin "binance" { query = "LINK/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "LINK/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "kucoin" { query = "LINK/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        alias "LINK/ETH" {
          origin "uniswapV3" { query = "LINK/WETH" }
        }
        reference { data_model = "ETH/USD" }
      }
      origin "coinbase" { query = "LINK/USD" }
      origin "kraken" { query = "LINK/USD" }
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
      origin "gemini" { query = "MATIC/USD" }
      indirect {
        origin "huobi" { query = "MATIC/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query = "MATIC/USD" }
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
        origin "uniswapV3" { query = "MKR/USDC" }
        reference { data_model = "USDC/USD" }
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

  data_model "OP/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "kucoin" { query = "OP/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "OP/USD" }
#      indirect {
#        alias "OP/ETH" {
#          origin "uniswapV3Op" { query = "OP/WETH" }
#        }
#        reference { data_model = "ETH/USD" }
#      }
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

  data_model "RETH/USD" {
    indirect {
      reference { data_model = "RETH/ETH" }
      reference { data_model = "ETH/USD" }
    }
  }

  data_model "SNX/USD" {
    median {
      min_values = 3
      indirect {
        origin "binance" { query = "SNX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "SNX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "hitbtc" { query = "SNX/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "SNX/USD" }
#      indirect {
#        alias "ETH/SNX" {
#          origin "uniswapV3Op" { query = "WETH/SNX" }
#        }
#        reference { data_model = "ETH/USD" }
#      }
    }
  }

  data_model "SOL/USD" {
    median {
      min_values = 4
      origin "coinbase" { query = "SOL/USD" }
      origin "kraken" { query = "SOL/USD" }
      origin "gemini" { query = "SOL/USD" }
      origin "bitfinex" { query = "SOL/USD" }
      indirect {
        origin "kucoin" { query = "SOL/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        origin "okx" { query = "SOL/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "binance" { query = "SOL/USD" }
      indirect {
        origin "hitbtc" { query = "SOL/BTC" }
        reference { data_model = "BTC/USD" }
      }
    }
  }

  data_model "STETH/ETH" {
    median {
      min_values = 2
      alias "STETH/ETH" {
        origin "balancerV2" { query = "STETH/WETH" }
      }
      origin "curve" { query = "STETH/ETH" }
    }
  }

  data_model "UNI/USD" {
    median {
      min_values = 3
      indirect {
        origin "okx" { query = "UNI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      indirect {
        alias "UNI/ETH" {
          origin "uniswapV3" { query = "UNI/WETH"}
        }
        reference { data_model = "ETH/USD" }
      }
      origin "coinbase" { query = "UNI/USD" }
      indirect {
        origin "binance" { query = "UNI/USDT" }
        reference { data_model = "USDT/USD" }
      }
    }
  }

  data_model "USDC/USD" {
    median {
      min_values = 2
#      indirect {
#        alias "USDC/ETH" {
#          origin "uniswapV3" { query = "USDC/WETH" }
#        }
#        reference { data_model = "ETH/USD" }
#      }
#      indirect {
#        origin "curve" { query = "USDT/USDC" }
#        reference { data_model = "USDT/USD" }
#      }
      indirect {
        origin "okx" { query = "USDT/USDC" }
        reference { data_model = "USDT/USD" }
      }
      origin "kraken" { query ="USDC/USD" }
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
#      indirect {
#        alias "WBTC/ETH" {
#          origin "curve" { query = "WBTC/WETH" }
#        }
#        reference { data_model = "ETH/USD" }
#      }
      indirect {
        origin "binance" { query = "WBTC/BTC" }
        reference { data_model = "BTC/USD" }
      }
      indirect {
        origin "coinbase" { query = "WBTC/BTC" }
        reference { data_model = "BTC/USD" }
      }
    }
  }

  data_model "WSTETH/ETH" {
    indirect {
      origin "wsteth" { query = "WSTETH/STETH" }
      reference { data_model = "STETH/ETH" }
    }
  }

  data_model "WSTETH/USD" {
    indirect {
      reference { data_model = "WSTETH/ETH" }
      reference { data_model = "ETH/USD" }
    }
  }

  data_model "YFI/USD" {
    median {
      min_values = 2
      indirect {
        alias "ETH/YFI" {
          origin "balancerV2" { query = "WETH/YFI" }
        }
        reference { data_model = "ETH/USD" }
      }
      indirect {
        origin "binance" { query = "YFI/USDT" }
        reference { data_model = "USDT/USD" }
      }
      origin "coinbase" { query = "YFI/USD" }
      origin "kraken" { query = "YFI/USD" }
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
    }
  }
}
