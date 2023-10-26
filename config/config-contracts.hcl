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
    "wat": "BTC/USD",
    "address": "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "arb1",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "BTC/USD",
    "address": "0xe0F30cb149fAADC7247E953746Be9BbBB6B5751f",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "ETH/BTC",
    "address": "0x81A679f98b63B3dDf2F17CB5619f4d6775b3c5ED",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "ETH/USD",
    "address": "0x64DE91F5A373Cd4c28de3600cB34C7C6cE410C85",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "GNO/USD",
    "address": "0x31BFA908637C29707e155Cfac3a50C9823bF8723",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "IBTA/USD",
    "address": "0xa5d4a331125D7Ece7252699e2d3CB1711950fBc8",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 10,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "LINK/USD",
    "address": "0xbAd4212d73561B240f10C56F27e6D9608963f17b",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "MATIC/USD",
    "address": "0xfe1e93840D286C83cF7401cB021B94b5bc1763d2",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "MKR/USD",
    "address": "0xdbbe5e9b1daa91430cf0772fcebe53f6c6f137df",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "RETH/USD",
    "address": "0xf86360f0127f8a441cfca332c75992d1c692b3d1",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "WSTETH/USD",
    "address": "0x2F73b6567B866302e132273f67661fB89b5a66F2",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 2,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "YFI/USD",
    "address": "0x89AC26C0aFCB28EC55B6CD2F6b7DAD867Fa24639",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 4,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "wat": "BTC/USD",
    "address": "0xdc65E49016ced01FC5aBEbB5161206B0f8063672",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "wat": "ETH/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "IMedian": true,
    "poke": {
      "expiration": 86400,
      "spread": 1,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "wat": "BTC/USD",
    "address": "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "BTC/USD",
    "address": "0x586409bb88cF89BBAB0e106b0620241a0e4005c9",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "ETH/BTC",
    "address": "0xaF495008d177a2E2AD95125b78ace62ef61Ed1f7",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "ETH/USD",
    "address": "0xD81834Aa83504F6614caE3592fb033e4b8130380",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "GNO/USD",
    "address": "0x0cd01b018C355a60B2Cc68A1e3d53853f05A7280",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "IBTA/USD",
    "address": "0x0Aca91081B180Ad76a848788FC76A089fB5ADA72",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 10,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "LINK/USD",
    "address": "0xe4919256D404968566cbdc5E5415c769D5EeBcb0",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "MATIC/USD",
    "address": "0x4b4e2A0b7a560290280F083c8b5174FB706D7926",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "MKR/USD",
    "address": "0x496C851B2A9567DfEeE0ACBf04365F3ba00Eb8dC",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "RETH/USD",
    "address": "0x7eEE7e44055B6ddB65c6C970B061EC03365FADB3",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "WSTETH/USD",
    "address": "0x9466e1ffA153a8BdBB5972a7217945eb2E28721f",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "YFI/USD",
    "address": "0x38D27Ba21E1B2995d0ff9C1C070c5c93dd07cB31",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "wat": "BTC/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "IMedian": true,
    "poke": {
      "expiration": 14400,
      "spread": 3,
      "interval": 60
    }
  }
]
models = [
  "AAVE/USD",
  "ARB/USD",
  "AVAX/USD",
  "BNB/USD",
  "BTC/USD",
  "CRV/USD",
  "DAI/USD",
  "DSR/RATE",
  "ETH/BTC",
  "ETH/USD",
  "FRAX/USD",
  "GNO/ETH",
  "GNO/USD",
  "IBTA/USD",
  "LDO/USD",
  "LINK/USD",
  "MATIC/USD",
  "MKR/ETH",
  "MKR/USD",
  "OP/USD",
  "RETH/ETH",
  "RETH/USD",
  "SDAI/DAI",
  "SDAI/ETH",
  "SDAI/MATIC",
  "SDAI/USD",
  "SNX/USD",
  "SOL/USD",
  "STETH/ETH",
  "STETH/USD",
  "UNI/USD",
  "USDC/USD",
  "USDT/USD",
  "WBTC/USD",
  "WSTETH/ETH",
  "WSTETH/USD",
  "XTZ/USD",
  "YFI/USD"
]
}
