variables {
  spectre_target_network  = try(env.CFG_SPECTRE_TARGET_NETWORK, "ethereum-mainnet")
  spectre_ethereum_client = split("-", try(env.CFG_SPECTRE_TARGET_NETWORK, "ethereum-mainnet"))[0]
  # List of median contracts that will be updated by the Relay.
  median_contracts        = {
    "ethereum-mainnet" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xe0F30cb149fAADC7247E953746Be9BbBB6B5751f",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      },
      "ETH/BTC" : {
        "msgExpiration" : 1800,
        "oracle" : "0x81A679f98b63B3dDf2F17CB5619f4d6775b3c5ED",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x64DE91F5A373Cd4c28de3600cB34C7C6cE410C85",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      },
      "GNO/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x31BFA908637C29707e155Cfac3a50C9823bF8723",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "IBTA/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xa5d4a331125D7Ece7252699e2d3CB1711950fBc8",
        "oracleExpiration" : 86400,
        "oracleSpread" : 10
      },
      "LINK/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xbAd4212d73561B240f10C56F27e6D9608963f17b",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "MANA/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x681c4F8f69cF68852BAd092086ffEaB31F5B812c",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "MATIC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xfe1e93840D286C83cF7401cB021B94b5bc1763d2",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "MKR/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xdbbe5e9b1daa91430cf0772fcebe53f6c6f137df",
        "oracleExpiration" : 86400,
        "oracleSpread" : 3
      },
      "RETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xf86360f0127f8a441cfca332c75992d1c692b3d1",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      },
      "WSTETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x2F73b6567B866302e132273f67661fB89b5a66F2",
        "oracleExpiration" : 86400,
        "oracleSpread" : 2
      },
      "YFI/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x89AC26C0aFCB28EC55B6CD2F6b7DAD867Fa24639",
        "oracleExpiration" : 86400,
        "oracleSpread" : 4
      }
    }
    "ethereum-goerli" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x586409bb88cF89BBAB0e106b0620241a0e4005c9",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "ETH/BTC" : {
        "msgExpiration" : 1800,
        "oracle" : "0xaF495008d177a2E2AD95125b78ace62ef61Ed1f7",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xD81834Aa83504F6614caE3592fb033e4b8130380",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "GNO/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x0cd01b018C355a60B2Cc68A1e3d53853f05A7280",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "IBTA/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x0Aca91081B180Ad76a848788FC76A089fB5ADA72",
        "oracleExpiration" : 14400,
        "oracleSpread" : 10
      },
      "LINK/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xe4919256D404968566cbdc5E5415c769D5EeBcb0",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "MANA/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xCCce898497e139831523cc9D23c948138dDF67f6",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "MATIC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x4b4e2A0b7a560290280F083c8b5174FB706D7926",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "MKR/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x496C851B2A9567DfEeE0ACBf04365F3ba00Eb8dC",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "RETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x7eEE7e44055B6ddB65c6C970B061EC03365FADB3",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "WSTETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x9466e1ffA153a8BdBB5972a7217945eb2E28721f",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "YFI/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x38D27Ba21E1B2995d0ff9C1C070c5c93dd07cB31",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      }
    }
    "arbitrum-mainnet" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xBBF1a875B13E4614645934faA3FEE59258320415",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      }
    }
    "arbitrum-goerli" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xBBF1a875B13E4614645934faA3FEE59258320415",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      }
    }
    "optimism-mainnet" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xdc65E49016ced01FC5aBEbB5161206B0f8063672",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
        "oracleExpiration" : 86400,
        "oracleSpread" : 1
      }
    }
    "optimism-goerli" : {
      "BTC/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      },
      "ETH/USD" : {
        "msgExpiration" : 1800,
        "oracle" : "0xBBF1a875B13E4614645934faA3FEE59258320415",
        "oracleExpiration" : 14400,
        "oracleSpread" : 3
      }
    }
  }
}

spectre {
  # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
  interval = 60

  dynamic "median" {
    for_each = var.median_contracts[var.spectre_target_network]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = var.spectre_ethereum_client

      # Address of the Median contract.
      contract_addr = contract.value.oracle

      # Name of the pair to fetch the price for.
      pair = replace(contract.key, "/", "")

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.oracleSpread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.oracleExpiration
    }
  }

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = try(env.CFG_FEEDS, "")=="*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[try(env.CFG_FEEDS, "prod")], split(",", try(env.CFG_FEEDS, "")))
}
