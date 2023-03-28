# List of feeds that are allowed to send price updates and event attestations.
feeds = [
  "0xDA1d2961Da837891f43235FddF66BAD26f41368b",
  "0x4b0E327C08e23dD08cb87Ec994915a5375619aa2",
  "0x75ef8432566A79C86BBF207A47df3963B8Cf0753",
  "0x83e23C207a67a9f9cB680ce84869B91473403e7d",
  "0xFbaF3a7eB4Ec2962bd1847687E56aAEE855F5D00",
  "0xfeEd00AA3F0845AFE52Df9ECFE372549B74C69D2",
  "0x71eCFF5261bAA115dcB1D9335c88678324b8A987",
  "0x8ff6a38A1CD6a42cAac45F08eB0c802253f68dfD",
  "0x16655369Eb59F3e1cAFBCfAC6D3Dd4001328f747",
  "0xD09506dAC64aaA718b45346a032F934602e29cca",
  "0xc00584B271F378A0169dd9e5b165c0945B4fE498",
  "0x60da93D9903cb7d3eD450D4F81D402f7C4F71dd9",
  "0xa580BBCB1Cee2BCec4De2Ea870D20a12A964819e",
  "0xD27Fa2361bC2CfB9A591fb289244C538E190684B",
  "0x8de9c5F1AC1D4d02bbfC25fD178f5DAA4D5B26dC",
  "0xE6367a7Da2b20ecB94A25Ef06F3b551baB2682e6",
  "0xA8EB82456ed9bAE55841529888cDE9152468635A",
  "0x130431b4560Cd1d74A990AE86C337a33171FF3c6",
  "0x8aFBD9c3D794eD8DF903b3468f4c4Ea85be953FB",
  "0xd94BBe83b4a68940839cD151478852d16B3eF891",
  "0xC9508E9E3Ccf319F5333A5B8c825418ABeC688BA",
  "0x77EB6CF8d732fe4D92c427fCdd83142DB3B742f7",
  "0x3CB645a8f10Fb7B0721eaBaE958F77a878441Cb9",
  "0x4f95d9B4D842B2E2B1d1AC3f2Cf548B93Fd77c67",
  "0xaC8519b3495d8A3E3E44c041521cF7aC3f8F63B3",
  "0xd72BA9402E9f3Ff01959D6c841DDD13615FFff42"
]

ethereum {
  client "default" {
    rpc_urls     = try(split(env.CFG_ETH_RPC_URLS, ","), ["https://eth.public-rpc.com"])
    chain_id     = try(env.CFG_ETH_CHAIN_ID, 1)
    ethereum_key = try(env.CFG_ETH_FROM, "") == "" ? "" : "default"
  }

  dynamic "key" {
    for_each = try(env.CFG_ETH_FROM, "") == "" ? [] : [1]
    labels   = ["default"]
    content {
      address         = try(env.CFG_ETH_FROM, "")
      keystore_path   = try(env.CFG_ETH_KEYS, "")
      passphrase_file = try(env.CFG_ETH_PASS, "")
    }
  }
}

transport {
  # LibP2P transport configuration. Always enabled.
  libp2p {
    priv_key_seed   = try(env.CFG_LIBP2P_PK_SEED, "")
    listen_addrs    = try(split(env.CFG_LIBP2P_LISTEN_ADDRS, ","), ["/ip4/0.0.0.0/tcp/8000"])
    bootstrap_addrs = try(split(env.CFG_LIBP2P_BOOTSTRAP_ADDRS, ","), [
      "/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi",
      "/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR"
    ])
    direct_peers_addrs = []
    blocked_addrs      = []
    disable_discovery  = false
    ethereum_key       = try(env.CFG_ETH_FROM, "") == "" ? "" : "default"
  }

  # WebAPI transport configuration. Enabled if CFG_WEBAPI_LISTEN_ADDR is set to a listen address.
  dynamic "webapi" {
    for_each = try(env.CFG_WEBAPI_LISTEN_ADDR, "") == "" ? [] : [1]
    content {
      listen_addr       = try(env.CFG_WEBAPI_LISTEN_ADDR, "0.0.0.0.8080")
      socks5_proxy_addr = try(env.CFG_WEBAPI_SOCKS5_PROXY_ADDR, "127.0.0.1:9050")

      # Ethereum based address book. Enabled if CFG_WEBAPI_ETH_ADDR_BOOK is set to a contract address.
      dynamic "ethereum_address_book" {
        for_each = try(env.CFG_WEBAPI_ETH_ADDR_BOOK, "") == "" ? [] : [1]
        content {
          contract_addr   = try(env.CFG_WEBAPI_ETH_ADDR_BOOK, "")
          ethereum_client = "default"
        }
      }

      # Static address book. Enabled if CFG_WEBAPI_STATIC_ADDR_BOOK is set to a comma separated list of addresses.
      dynamic "static_address_book" {
        for_each = try(env.CFG_WEBAPI_STATIC_ADDR_BOOK, "") == "" ? [] : [1]
        content {
          addresses = try(split(env.CFG_WEBAPI_STATIC_ADDR_BOOK, ","), "")
        }
      }
    }
  }
}

spire {
  rpc_listen_addr = try(env.CFG_SPIRE_RPC_ADDR, "127.0.0.1:9100")
  rpc_agent_addr  = try(env.CFG_SPIRE_RPC_ADDR, "127.0.0.1:9100")

  # List of pairs that are collected by the spire node. Other pairs are ignored.
  pairs = [
    "AAVEUSD",
    "AVAXUSD",
    "BALUSD",
    "BATUSD",
    "BTCUSD",
    "COMPUSD",
    "CRVUSD",
    "DOTUSD",
    "ETHBTC",
    "ETHUSD",
    "FILUSD",
    "GNOUSD",
    "IBTAUSD",
    "LINKUSD",
    "LRCUSD",
    "MANAUSD",
    "MKRETH",
    "MKRUSD",
    "PAXGUSD",
    "RETHUSD",
    "SNXUSD",
    "SOLUSD",
    "UNIUSD",
    "USDTUSD",
    "WNXMUSD",
    "XRPUSD",
    "XTZUSD",
    "YFIUSD",
    "ZECUSD",
    "ZRXUSD",
    "STETHUSD",
    "WSTETHUSD",
    "MATICUSD"
  ]
}

gofer {
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

  origin "binance_us" {
    type   = "binance"
    params = {
      url = "https://www.binance.us"
    }
  }

  origin "bittrex" {
    type   = "bittrex"
    params = {
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

  origin "openexchangerates" {
    type   = "openexchangerates"
    params = {
      api_key = try(env.GOFER_OPENEXCHANGERATES_API_KEY, "")
    }
  }

  origin "poloniex" {
    type   = "poloniex"
    params = {
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

  price_model "ETH/USD" "median" {
    source "ETH/USD" "indirect" {
      source "ETH/BTC" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "ETH/USD" "origin" { origin = "bitstamp" }
    source "ETH/USD" "origin" { origin = "coinbasepro" }
    source "ETH/USD" "origin" { origin = "gemini" }
    source "ETH/USD" "origin" { origin = "kraken" }
    source "ETH/USD" "origin" { origin = "uniswapV3" }
    min_sources = 3
  }

  price_model "GNO/USD" "median" {
    source "GNO/USD" "indirect" {
      source "ETH/GNO" "origin" { origin = "balancerV2" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "GNO/USD" "indirect" {
      source "GNO/ETH" "origin" { origin = "uniswapV3" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "GNO/USD" "indirect" {
      source "GNO/BTC" "origin" { origin = "kraken" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "GNO/USD" "indirect" {
      source "GNO/USDT" "origin" { origin = "binance" }
      source "USDT/USD" "origin" { origin = "." }
    }
    min_sources = 3
  }

  price_model "IBTA/USD" "origin" {
    origin = "ishares"
  }

  price_model "LINK/USD" "median" {
    source "LINK/USD" "indirect" {
      source "LINK/BTC" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "LINK/USD" "origin" { origin = "bitstamp" }
    source "LINK/USD" "origin" { origin = "coinbasepro" }
    source "LINK/USD" "origin" { origin = "gemini" }
    source "LINK/USD" "origin" { origin = "kraken" }
    source "LINK/USD" "indirect" {
      source "LINK/ETH" "origin" { origin = "uniswapV3" }
      source "ETH/USD" "origin" { origin = "." }
    }
    min_sources = 3
  }

  price_model "MANA/USD" "median" {
    source "MANA/USD" "indirect" {
      source "MANA/BTC" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "MANA/USD" "origin" { origin = "binance_us" }
    source "MANA/USD" "origin" { origin = "coinbasepro" }
    source "MANA/USD" "origin" { origin = "kraken" }
    source "MANA/USD" "indirect" {
      source "MANA/USDT" "origin" { origin = "okx" }
      source "USDT/USD" "origin" { origin = "." }
    }
    source "MANA/USD" "indirect" {
      source "MANA/KRW" "origin" { origin = "upbit" }
      source "KRW/USD" "origin" { origin = "openexchangerates" }
    }
    min_sources = 3
  }

  price_model "MATIC/USD" "median" {
    source "MATIC/USD" "indirect" {
      source "MATIC/USDT" "origin" { origin = "binance" }
      source "USDT/USD" "origin" { origin = "." }
    }
    source "MATIC/USD" "origin" { origin = "coinbasepro" }
    source "MATIC/USD" "origin" { origin = "gemini" }
    source "MATIC/USD" "indirect" {
      source "MATIC/USDT" "origin" { origin = "huobi" }
      source "USDT/USD" "origin" { origin = "." }
    }
    source "MATIC/USD" "origin" { origin = "kraken" }
    min_sources = 3
  }

  price_model "MKR/USD" "median" {
    source "MKR/USD" "indirect" {
      source "MKR/BTC" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "MKR/USD" "origin" { origin = "bitstamp" }
    source "MKR/USD" "origin" { origin = "coinbasepro" }
    source "MKR/USD" "origin" { origin = "gemini" }
    source "MKR/USD" "origin" { origin = "kraken" }
    source "MKR/USD" "indirect" {
      source "MKR/ETH" "origin" { origin = "uniswapV3" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "MKR/USD" "indirect" {
      source "MKR/USDC" "origin" { origin = "uniswapV3" }
      source "USDC/USD" "origin" { origin = "." }
    }
    min_sources = 3
  }

  price_model "MKR/ETH" "median" {
    source "MKR/ETH" "indirect" {
      source "MKR/BTC" "origin" { origin = "binance" }
      source "ETH/BTC" "origin" { origin = "." }
    }
    source "MKR/ETH" "indirect" {
      source "MKR/USD" "origin" { origin = "bitstamp" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "MKR/ETH" "indirect" {
      source "MKR/USD" "origin" { origin = "coinbasepro" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "MKR/ETH" "indirect" {
      source "MKR/USD" "origin" { origin = "gemini" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "MKR/ETH" "indirect" {
      source "MKR/USD" "origin" { origin = "kraken" }
      source "ETH/USD" "origin" { origin = "." }
    }
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

  price_model "RETH/USD" "indirect" {
    source "RETH/ETH" "origin" { origin = "." }
    source "ETH/USD" "origin" { origin = "." }
  }

  price_model "STETH/ETH" "median" {
    source "STETH/ETH" "origin" { origin = "balancerV2" }
    source "STETH/ETH" "origin" { origin = "curve" }
    min_sources = 2
  }

  price_model "USDC/USD" "median" {
    source "USDC/USD" "origin" { origin = "coinbasepro" }
    source "USDC/USD" "origin" { origin = "gemini" }
    source "USDC/USD" "origin" { origin = "kraken" }
    min_sources = 2
  }

  price_model "USDT/USD" "median" {
    source "USDT/USD" "indirect" {
      source "BTC/USDT" "origin" { origin = "binance" }
      source "BTC/USD" "origin" { origin = "." }
    }
    source "USDT/USD" "origin" { origin = "bitfinex" }
    source "USDT/USD" "origin" { origin = "coinbasepro" }
    source "USDT/USD" "origin" { origin = "kraken" }
    source "USDT/USD" "indirect" {
      source "BTC/USDT" "origin" { origin = "okx" }
      source "BTC/USD" "origin" { origin = "." }
    }
    min_sources = 3
  }

  price_model "WSTETH/ETH" "indirect" {
    source "WSTETH/STETH" "origin" { origin = "wsteth" }
    source "STETH/ETH" "origin" { origin = "." }
  }

  price_model "WSTETH/USD" "indirect" {
    source "WSTETH/ETH" "origin" { origin = "." }
    source "ETH/USD" "origin" { origin = "." }
  }

  price_model "YFI/USD" "median" {
    source "YFI/USD" "indirect" {
      source "ETH/YFI" "origin" { origin = "balancerV2" }
      source "ETH/USD" "origin" { origin = "." }
    }
    source "YFI/USD" "indirect" {
      source "YFI/USDT" "origin" { origin = "binance" }
      source "USDT/USD" "origin" { origin = "." }
    }
    source "YFI/USD" "origin" { origin = "coinbasepro" }
    source "YFI/USD" "origin" { origin = "kraken" }
    source "YFI/USD" "indirect" {
      source "YFI/USDT" "origin" { origin = "okx" }
      source "USDT/USD" "origin" { origin = "." }
    }
    source "YFI/USD" "indirect" {
      source "YFI/ETH" "origin" { origin = "sushiswap" }
      source "ETH/USD" "origin" { origin = "." }
    }
    min_sources = 2
  }
}

leeloo {
  ethereum_key = "default"

  teleport_starknet {
    sequencer       = "https://alpha-mainnet.starknet.io"
    contract_addrs  = ["0x070077337f82db40b34adc7458761ec193d6ab7444f3da5b44d750afdd065d4f"]
    interval        = 10
    prefetch_period = 3600
    replay_after    = [60]
  }
}

lair {
  listen_addr = try(env.CFG_LAIR_LISTEN_ADDR, "0.0.0.0:8082")

  # Configuration for memory storage. Enabled if CFG_LAIR_STORAGE is "memory" or unset.
  dynamic "storage_memory" {
    for_each = try(env.CFG_LAIR_STORAGE, "memory") == "memory" ? [1] : []
    content {}
  }

  # Configuration for redis storage. Enabled if CFG_LAIR_STORAGE is "redis".
  dynamic "storage_redis" {
    for_each = try(env.CFG_LAIR_STORAGE, "") == "redis" ? [1] : []
    content {
      addr                     = try(env.CFG_LAIR_REDIS_ADDR, "127.0.0.1:6379")
      user                     = try(env.CFG_LAIR_REDIS_USER, "")
      pass                     = try(env.CFG_LAIR_REDIS_PASS, "")
      db                       = try(env.CFG_LAIR_REDIS_DB, 0)
      memory_limit             = try(env.CFG_LAIR_REDIS_MEMORY_LIMIT, 0)
      tls                      = try(env.CFG_LAIR_REDIS_TLS == "true", false)
      tls_server_name          = try(env.CFG_LAIR_REDIS_TLS_SERVER_NAME, "")
      tls_cert_file            = try(env.CFG_LAIR_REDIS_TLS_CERT_FILE, "")
      tls_key_file             = try(env.CFG_LAIR_REDIS_TLS_KEY_FILE, "")
      tls_root_ca_file         = try(env.CFG_LAIR_REDIS_TLS_ROOT_CA_FILE, "")
      tls_insecure_skip_verify = try(env.CFG_LAIR_REDIS_TLS_INSECURE == "true", false)
      cluster                  = try(env.CFG_LAIR_REDIS_CLUSTER == "true", false)
      cluster_addrs            = try(split(env.CFG_LAIR_REDIS_CLUSTER_ADDRS, ","), [])
    }
  }
}
