variables {
contract_map = {
  "prod-eth-Chainlog": "0xE10e8f60584e2A2202864d5B35A098EA84761256",
  "prod-eth-TorAddressRegister": "0x16515EEe550Fe7ae3b8f70bdfb737a57811B3C96",
  "prod-eth-WatRegistry": "0x594d52fDB6570F07879Bb2AF8a36c3bF00BC7F00",
  "stage-sep-Chainlog": "0xfc71a2e4497d065416A1BBDA103330a381F8D3b1",
  "stage-sep-TorAddressRegister": "0x504Fdbc4a9386c2C48A5775a6967beB00dAa9E9a",
  "stage-sep-WatRegistry": "0xE5f12C7285518bA5C6eEc15b00855A47C19d9557"
}
contracts = [
  {
    "env": "prod",
    "chain": "arb1",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "arb1",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "chain_id": 1,
    "IScribe": true,
    "wat": "BTC/USD",
    "IScribeOptimistic": true,
    "address": "0x898D1aB819a24880F636416df7D1493C94143262",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "chain_id": 1,
    "IScribe": true,
    "wat": "ETH/USD",
    "IScribeOptimistic": true,
    "address": "0x5E16CA75000fb2B9d7B1184Fa24fF5D938a345Ef",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "chain_id": 1,
    "IScribe": true,
    "wat": "MKR/USD",
    "IScribeOptimistic": true,
    "address": "0xb400027B7C31D67982199Fa48B8228F128691fCb",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "chain_id": 1,
    "IScribe": true,
    "wat": "RETH/USD",
    "IScribeOptimistic": true,
    "address": "0x608D9cD5aC613EBAC4549E6b8A73954eA64C3660",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "chain_id": 1,
    "IScribe": true,
    "wat": "WSTETH/USD",
    "IScribeOptimistic": true,
    "address": "0x013C5C46db9914A19A58E57AD539eD5B125aFA15",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0xe0F30cb149fAADC7247E953746Be9BbBB6B5751f",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "ETH/BTC",
    "address": "0x81A679f98b63B3dDf2F17CB5619f4d6775b3c5ED",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0x64DE91F5A373Cd4c28de3600cB34C7C6cE410C85",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "GNO/USD",
    "address": "0x31BFA908637C29707e155Cfac3a50C9823bF8723",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "IBTA/USD",
    "address": "0xa5d4a331125D7Ece7252699e2d3CB1711950fBc8",
    "poke": {
      "expiration": 86400,
      "spread": 10,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "LINK/USD",
    "address": "0xbAd4212d73561B240f10C56F27e6D9608963f17b",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "MATIC/USD",
    "address": "0xfe1e93840D286C83cF7401cB021B94b5bc1763d2",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "MKR/USD",
    "address": "0xdbbe5e9b1daa91430cf0772fcebe53f6c6f137df",
    "poke": {
      "expiration": 86400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "RETH/USD",
    "address": "0xf86360f0127f8a441cfca332c75992d1c692b3d1",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "WSTETH/USD",
    "address": "0x2F73b6567B866302e132273f67661fB89b5a66F2",
    "poke": {
      "expiration": 86400,
      "spread": 2,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "IMedian": true,
    "wat": "YFI/USD",
    "address": "0x89AC26C0aFCB28EC55B6CD2F6b7DAD867Fa24639",
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0xdc65E49016ced01FC5aBEbB5161206B0f8063672",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "BTC/USD",
    "IScribeOptimistic": false,
    "address": "0x898D1aB819a24880F636416df7D1493C94143262",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "DSR/RATE",
    "IScribeOptimistic": false,
    "address": "0xbBC385C209bC4C8E00E3687B51E25E21b0E7B186",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "ETH/USD",
    "IScribeOptimistic": false,
    "address": "0x5E16CA75000fb2B9d7B1184Fa24fF5D938a345Ef",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "MATIC/USD",
    "IScribeOptimistic": false,
    "address": "0xE0ECe625B1E128EE00e39BB91A80772D5d4d8Ed5",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/DAI",
    "IScribeOptimistic": false,
    "address": "0xfFcF8e5A12Acc48870D2e8834310aa270dE10fE6",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/ETH",
    "IScribeOptimistic": false,
    "address": "0xE6DF058512F99c0C8940736687aDdb38722c73C0",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/MATIC",
    "IScribeOptimistic": false,
    "address": "0x6c9571D1dD3e606Ce734Cc558bdd0BE576E01660",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0x586409bb88cF89BBAB0e106b0620241a0e4005c9",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "ETH/BTC",
    "address": "0xaF495008d177a2E2AD95125b78ace62ef61Ed1f7",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0xD81834Aa83504F6614caE3592fb033e4b8130380",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "GNO/USD",
    "address": "0x0cd01b018C355a60B2Cc68A1e3d53853f05A7280",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "IBTA/USD",
    "address": "0x0Aca91081B180Ad76a848788FC76A089fB5ADA72",
    "poke": {
      "expiration": 14400,
      "spread": 10,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "LINK/USD",
    "address": "0xe4919256D404968566cbdc5E5415c769D5EeBcb0",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "MATIC/USD",
    "address": "0x4b4e2A0b7a560290280F083c8b5174FB706D7926",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "MKR/USD",
    "address": "0x496C851B2A9567DfEeE0ACBf04365F3ba00Eb8dC",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "RETH/USD",
    "address": "0x7eEE7e44055B6ddB65c6C970B061EC03365FADB3",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "WSTETH/USD",
    "address": "0x9466e1ffA153a8BdBB5972a7217945eb2E28721f",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "IMedian": true,
    "wat": "YFI/USD",
    "address": "0x38D27Ba21E1B2995d0ff9C1C070c5c93dd07cB31",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "IMedian": true,
    "wat": "BTC/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "IMedian": true,
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "AAVE/USD",
    "IScribeOptimistic": true,
    "address": "0xa38C2B5408Eb1DCeeDBEC5d61BeD580589C6e717",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "ARB/USD",
    "IScribeOptimistic": true,
    "address": "0x579BfD0581beD0d18fBb0Ebab099328d451552DD",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "AVAX/USD",
    "IScribeOptimistic": true,
    "address": "0x78C8260AF7C8D0d17Cf3BA91F251E9375A389688",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "BNB/USD",
    "IScribeOptimistic": true,
    "address": "0x26EE3E8b618227C1B735D8D884d52A852410019f",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "BTC/USD",
    "IScribeOptimistic": true,
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "CRV/USD",
    "IScribeOptimistic": true,
    "address": "0xf29a932ae56bB96CcACF8d1f2Da9028B01c8F030",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "DAI/USD",
    "IScribeOptimistic": true,
    "address": "0xa7aA6a860D17A89810dE6e6278c58EB21Fa00fc4",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "ETH/BTC",
    "IScribeOptimistic": true,
    "address": "0x1804969b296E89C1ddB1712fA99816446956637e",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "ETH/USD",
    "IScribeOptimistic": true,
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "GNO/USD",
    "IScribeOptimistic": true,
    "address": "0xA28dCaB66FD25c668aCC7f232aa71DA1943E04b8",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "IBTA/USD",
    "IScribeOptimistic": true,
    "address": "0x07487b0Bf28801ECD15BF09C13e32FBc87572e81",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "LDO/USD",
    "IScribeOptimistic": true,
    "address": "0xa53dc5B100f0e4aB593f2D8EcD3c5932EE38215E",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "LINK/USD",
    "IScribeOptimistic": true,
    "address": "0xecB89B57A60ac44E06ab1B767947c19b236760c3",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "MATIC/USD",
    "IScribeOptimistic": true,
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "MKR/USD",
    "IScribeOptimistic": true,
    "address": "0x67ffF0C6abD2a36272870B1E8FE42CC8E8D5ec4d",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "OP/USD",
    "IScribeOptimistic": true,
    "address": "0xfadF055f6333a4ab435D2D248aEe6617345A4782",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "RETH/USD",
    "IScribeOptimistic": true,
    "address": "0xEE02370baC10b3AC3f2e9eebBf8f3feA1228D263",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "SDAI/DAI",
    "IScribeOptimistic": true,
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "SNX/USD",
    "IScribeOptimistic": true,
    "address": "0xD20f1eC72bA46b6126F96c5a91b6D3372242cE98",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "SOL/USD",
    "IScribeOptimistic": true,
    "address": "0x4D1e6f39bbfcce8b471171b8431609b83f3a096D",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "UNI/USD",
    "IScribeOptimistic": true,
    "address": "0x2aFF768F5d6FC63fA456B062e02f2049712a1ED5",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "USDC/USD",
    "IScribeOptimistic": true,
    "address": "0x1173da1811a311234e7Ab0A33B4B7B646Ff42aEC",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "USDT/USD",
    "IScribeOptimistic": true,
    "address": "0x0bd446021Ab95a2ABd638813f9bDE4fED3a5779a",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "WBTC/USD",
    "IScribeOptimistic": true,
    "address": "0xA7226d85CE5F0DE97DCcBDBfD38634D6391d0584",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "WSTETH/USD",
    "IScribeOptimistic": true,
    "address": "0xc9Bb81d3668f03ec9109bBca77d32423DeccF9Ab",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "chain_id": 11155111,
    "IScribe": true,
    "wat": "YFI/USD",
    "IScribeOptimistic": true,
    "address": "0x0893EcE705639112C1871DcE88D87D81540D0199",
    "challenge_period": 3600,
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    },
    "optimistic_poke": {
      "spread": 0.5,
      "expiration": 28800,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "BTC/USD",
    "IScribeOptimistic": false,
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "DSR/RATE",
    "IScribeOptimistic": false,
    "address": "0x729af3A41AE9E707e7AE421569C4b9c632B66a0c",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "ETH/USD",
    "IScribeOptimistic": false,
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "MATIC/USD",
    "IScribeOptimistic": false,
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/DAI",
    "IScribeOptimistic": false,
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/ETH",
    "IScribeOptimistic": false,
    "address": "0x05aB94eD168b5d18B667cFcbbA795789C750D893",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "chain_id": 1101,
    "IScribe": true,
    "wat": "SDAI/MATIC",
    "IScribeOptimistic": false,
    "address": "0x2f0e0dE1F8c11d2380dE093ED15cA6cE07653cbA",
    "poke": {
      "spread": 1,
      "expiration": 32400,
      "interval": 120
    }
  }
]
}
