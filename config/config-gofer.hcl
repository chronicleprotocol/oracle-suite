gofer {
  origin "balancerV2" {
    type = "balancerV2"
    contracts "ethereum" {
      addresses = {
        "WETH/GNO"    = "0xF4C0DD9B82DA36C07605df83c8a416F11724d88b" # WeightedPool2Tokens
        "RETH/WETH"   = "0x1E19CF2D73a72Ef1332C882F20534B6519Be0276" # MetaStablePool
      }
      references = {
        "RETH/WETH"   = "0xae78736Cd615f374D3085123A210448E74Fc6393" # token0 of RETH/WETH
      }
    }
  }

  origin "composableBalancerV2" {
    type = "composable_balancerV2"
    contracts "ethereum" {
      addresses = {
        "GHO/LUSD"                    = "0x3FA8C89704e5d07565444009e5d9e624B40Be813"
        "WSTETH/WSTETH_WETH_BPT/WETH" = "0x93d199263632a4EF4Bb438F1feB99e57b4b5f0BD"
      }
    }
  }

  origin "weightedBalancerV2" {
    type = "weighted_balancerV2"
    contracts "ethereum" {
      addresses = {
        "WUSDM/WSTETH" = "0x54ca50EE86616379420Cc56718E12566aa75Abbe"
        "SD/ETHX"      = "0x034E2d995B39A88aB9a532A9BF0deDDac2c576eA"
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
    url = "https://api.bybit.com/v5/market/tickers?category=spot&symbol=$${ucbase}$${ucquote}"
    jq = "{price: .result.list[0].lastPrice|tonumber, volume: .result.list[0].volume24h|tonumber, time: (.time/1000)|round}"
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
        # int256, stableswap
        "ETH/STETH"     = "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022",
        "DAI/USDC/USDT" = "0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7",
        "FRAX/USDC"     = "0xDcEF968d416a41Cdac0ED8702fAC8128A64241A2",
        "ETH/ETHX"      = "0x59Ab5a5b5d617E478a2479B0cAD80DA7e2831492"
      }
      addresses2 = {
        # uint256, cryptoswap
        "WETH/LDO"       = "0x9409280DC1e6D33AB7A8C6EC03e5763FB61772B5",
        "USDT/WBTC/WETH" = "0xD51a44d3FaE010294C616388b506AcdA1bfAAE46",
        "WETH/YFI"       = "0xC26b89A667578ec7b3f11b2F98d6Fd15C07C54ba",
        "WETH/RETH"      = "0x0f3159811670c117c372428D4E69AC32325e4D0F",
        "WSTETH/ETHX"    = "0x14756A5eD229265F86990e749285bDD39Fe0334F"
      }
    }
  }

  origin "dsr" {
    type = "dsr"
    contracts "ethereum" {
      addresses = {
        "DSR/RATE" = "0x197E90f9FAD81970bA7976f33CbD77088E5D7cf7" # address to pot contract
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
        # query in data model must be `1DAY` or `nDAYS`, n <= 7, i.e. "STETH/7DAYS"
        "STETH/ERC20" = "0xae7ab96520DE3A18E5e111B5EaAb095312D7fE84"
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
        "WSTETH/WETH" = "0x3a1b97Fc25fA45832F588ED3bFb2A0f74ddBD4F8",
        "RETH/WETH"   = "0x2201d2400d30BFD8172104B4ad046d019CA4E7bd"
      }
    }
  }

  origin "rocketpool" {
    type = "rocketpool"
    contracts "ethereum" {
      addresses = {
        "RETH/ETH" = "0xae78736Cd615f374D3085123A210448E74Fc6393"
      }
    }
  }

  origin "sdai" {
    type = "sdai"
    contracts "ethereum" {
      addresses = {
        "SDAI/DAI" = "0x83F20F44975D03b1b09e64809B757c47f942BEeA"
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
        "MATIC/WETH"  = "0x290A6a7460B308ee3F19023D2D00dE604bcf5B42",
        "ETHX/WETH"   = "0x1b9669b12959Ad51B01FaBcF01EaBDFADB82f578",
        "SD/USDC"     = "0xc72AbB13B6BDfA64770cb5B1F57Bebd36a91A29E",
        "RETH/WETH"   = "0xa4e0faA58465A2D369aa21B3e42d43374c6F9613",
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
      freshness_threshold = 3600 * 8
      expiry_threshold    = 3600 * 24
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

  dynamic "data_model" {
    for_each = distinct([
      for v in var.contracts : v.wat
      # Limit the list only to a specific environment but take all chains
      if v.env == var.environment
      # Only Median compatible contracts
      && try(v.is_median, false)
    ])
    iterator = symbol
    labels   = [replace(symbol.value, "/", "")]
    content {
      reference { data_model = symbol.value }
    }
  }
}

