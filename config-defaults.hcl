variables {
  # List of supported asset symbols. This determines Feed behavior.
  data_symbols = [
    "BTC/USD",
    "ETH/BTC",
    "ETH/USD",
    "GNO/USD",
    "IBTA/USD",
    "LINK/USD",
    "MATIC/USD",
    "MKR/USD",
    "RETH/USD",
    "WSTETH/USD",
    "YFI/USD",
  ]

  # Default sets of Feeds to use for the app.
  # CFG_FEEDS environment variable can control which set to use.
  # Set it to one of the keys in the below map to use the Feeds configures therein
  # or use "*" as a wildcard to use both sets of Feeds.
  feed_sets = {
    "prod" : [
      "0x130431b4560Cd1d74A990AE86C337a33171FF3c6",
      "0x16655369Eb59F3e1cAFBCfAC6D3Dd4001328f747",
      "0x3CB645a8f10Fb7B0721eaBaE958F77a878441Cb9",
      "0x4b0E327C08e23dD08cb87Ec994915a5375619aa2",
      "0x4f95d9B4D842B2E2B1d1AC3f2Cf548B93Fd77c67",
      "0x60da93D9903cb7d3eD450D4F81D402f7C4F71dd9",
      "0x71eCFF5261bAA115dcB1D9335c88678324b8A987",
      "0x75ef8432566A79C86BBF207A47df3963B8Cf0753",
      "0x77EB6CF8d732fe4D92c427fCdd83142DB3B742f7",
      "0x83e23C207a67a9f9cB680ce84869B91473403e7d",
      "0x8aFBD9c3D794eD8DF903b3468f4c4Ea85be953FB",
      "0x8de9c5F1AC1D4d02bbfC25fD178f5DAA4D5B26dC",
      "0x8ff6a38A1CD6a42cAac45F08eB0c802253f68dfD",
      "0xa580BBCB1Cee2BCec4De2Ea870D20a12A964819e",
      "0xA8EB82456ed9bAE55841529888cDE9152468635A",
      "0xaC8519b3495d8A3E3E44c041521cF7aC3f8F63B3",
      "0xc00584B271F378A0169dd9e5b165c0945B4fE498",
      "0xC9508E9E3Ccf319F5333A5B8c825418ABeC688BA",
      "0xD09506dAC64aaA718b45346a032F934602e29cca",
      "0xD27Fa2361bC2CfB9A591fb289244C538E190684B",
      "0xd72BA9402E9f3Ff01959D6c841DDD13615FFff42",
      "0xd94BBe83b4a68940839cD151478852d16B3eF891",
      "0xDA1d2961Da837891f43235FddF66BAD26f41368b",
      "0xE6367a7Da2b20ecB94A25Ef06F3b551baB2682e6",
      "0xFbaF3a7eB4Ec2962bd1847687E56aAEE855F5D00",
      "0xfeEd00AA3F0845AFE52Df9ECFE372549B74C69D2",
    ]
    "stage" : [
      "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18",
      "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD",
      "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9",
      "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4",
    ]
  }

  # List of median contracts that will be updated by the Relay.
  median_contracts = {
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

  scribe_contracts = {
    "sep" : {
      "BTC/USD" : {
        "address" : "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "ETH/BTC" : {
        "address" : "0x1804969b296E89C1ddB1712fA99816446956637e",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "CRV/USD" : {
        "address" : "0xf29a932ae56bB96CcACF8d1f2Da9028B01c8F030",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "SDAI/DAI" : {
        "address" : "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "USDC/USD" : {
        "address" : "0x1173da1811a311234e7Ab0A33B4B7B646Ff42aEC",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "DAI/USD" : {
        "address" : "0xa7aA6a860D17A89810dE6e6278c58EB21Fa00fc4",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "AVAX/USD" : {
        "address" : "0x78C8260AF7C8D0d17Cf3BA91F251E9375A389688",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "IBTA/USD" : {
        "address" : "0x07487b0Bf28801ECD15BF09C13e32FBc87572e81",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "AAVE/USD" : {
        "address" : "0xa38C2B5408Eb1DCeeDBEC5d61BeD580589C6e717",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "WBTC/USD" : {
        "address" : "0xA7226d85CE5F0DE97DCcBDBfD38634D6391d0584",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "BNB/USD" : {
        "address" : "0x26EE3E8b618227C1B735D8D884d52A852410019f",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "YFI/USD" : {
        "address" : "0x0893EcE705639112C1871DcE88D87D81540D0199",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "MATIC/USD" : {
        "address" : "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "LINK/USD" : {
        "address" : "0xecB89B57A60ac44E06ab1B767947c19b236760c3",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "UNI/USD" : {
        "address" : "0x2aFF768F5d6FC63fA456B062e02f2049712a1ED5",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "ETH/USD" : {
        "address" : "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "WSTETH/USD" : {
        "address" : "0xc9Bb81d3668f03ec9109bBca77d32423DeccF9Ab",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "RETH/USD" : {
        "address" : "0xEE02370baC10b3AC3f2e9eebBf8f3feA1228D263",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "SOL/USD" : {
        "address" : "0x4D1e6f39bbfcce8b471171b8431609b83f3a096D",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "GNO/USD" : {
        "address" : "0xA28dCaB66FD25c668aCC7f232aa71DA1943E04b8",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "ARB/USD" : {
        "address" : "0x579BfD0581beD0d18fBb0Ebab099328d451552DD",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "SNX/USD" : {
        "address" : "0xD20f1eC72bA46b6126F96c5a91b6D3372242cE98",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "MKR/USD" : {
        "address" : "0x67ffF0C6abD2a36272870B1E8FE42CC8E8D5ec4d",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "OP/USD" : {
        "address" : "0xfadF055f6333a4ab435D2D248aEe6617345A4782",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "USDT/USD" : {
        "address" : "0x0bd446021Ab95a2ABd638813f9bDE4fED3a5779a",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      },
      "LDO/USD" : {
        "address" : "0xa53dc5B100f0e4aB593f2D8EcD3c5932EE38215E",
        "spread" : 0.5,
        "expiration" : 3600,
        "interval" : 10
      }
    }
  }
}
