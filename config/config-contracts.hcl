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
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "arb1",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "BTC/USD",
    "address": "0x9Af8fe1d0c9ED3f176Dd3559B6f4b6FeF3AAb83B",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "BTC/USD",
    "address": "0xe0F30cb149fAADC7247E953746Be9BbBB6B5751f",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "DAI/USD",
    "address": "0xf2dc732221e2b374eBBfd0023EF794c4432E66d8",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "ETH/BTC",
    "address": "0x81A679f98b63B3dDf2F17CB5619f4d6775b3c5ED",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "ETH/USD",
    "address": "0x1174948681bb05748E3682398d9b7a6836B07554",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "ETH/USD",
    "address": "0x64DE91F5A373Cd4c28de3600cB34C7C6cE410C85",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "GNO/USD",
    "address": "0x0b4d1660D9f28203a23C33808112FF44cA7bCE41",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 2,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "GNO/USD",
    "address": "0x31BFA908637C29707e155Cfac3a50C9823bF8723",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "IBTA/USD",
    "address": "0xa5d4a331125D7Ece7252699e2d3CB1711950fBc8",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 10
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "LINK/USD",
    "address": "0xbAd4212d73561B240f10C56F27e6D9608963f17b",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "MATIC/USD",
    "address": "0xfe1e93840D286C83cF7401cB021B94b5bc1763d2",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "MKR/USD",
    "address": "0xc4962E0c282b52d00f995f5C70d4695e4Ac14F57",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "MKR/USD",
    "address": "0xdbbe5e9b1daa91430cf0772fcebe53f6c6f137df",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 3
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "RETH/USD",
    "address": "0x3Fcc752dc6Fb8acc80E3e0843C16ea080240b57F",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 2,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "RETH/USD",
    "address": "0xf86360f0127f8a441cfca332c75992d1c692b3d1",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "SDAI/USD",
    "address": "0xe53e78006d2c3E905d73cBdb31b8E43ec06F27A9",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "USDC/USD",
    "address": "0x209186cd917dfaBd9529935dd7202C755a59f06F",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "WSTETH/ETH",
    "address": "0x84A48F89D5844385C515f43797147D6aF61cE2AE",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "WSTETH/USD",
    "address": "0x12a8Ad45db5085e17aBabb3016bba67cc6Bac5Db",
    "chain_id": 1,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
    "poke": {
      "spread": 3,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "WSTETH/USD",
    "address": "0x2F73b6567B866302e132273f67661fB89b5a66F2",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 2
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "wat": "YFI/USD",
    "address": "0x89AC26C0aFCB28EC55B6CD2F6b7DAD867Fa24639",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 4
    }
  },
  {
    "env": "prod",
    "chain": "gno",
    "wat": "BTC/USD",
    "address": "0x898D1aB819a24880F636416df7D1493C94143262",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.25,
      "expiration": 14400
    }
  },
  {
    "env": "prod",
    "chain": "gno",
    "wat": "DAI/USD",
    "address": "0x64596dEb187A1F4dD73240474A18e854AEAe22f7",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.25,
      "expiration": 14400
    }
  },
  {
    "env": "prod",
    "chain": "gno",
    "wat": "ETH/USD",
    "address": "0x5E16CA75000fb2B9d7B1184Fa24fF5D938a345Ef",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.25,
      "expiration": 14400
    }
  },
  {
    "env": "prod",
    "chain": "gno",
    "wat": "GNO/USD",
    "address": "0x92D2E219f7175dce742Bc1aF65c25D11E0e9095e",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.25,
      "expiration": 14400
    }
  },
  {
    "env": "prod",
    "chain": "gno",
    "wat": "WSTETH/ETH",
    "address": "0xe189932051328bAf256bea646c01D0898258C4A9",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.25,
      "expiration": 14400
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "wat": "BTC/USD",
    "address": "0xdc65E49016ced01FC5aBEbB5161206B0f8063672",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "oeth",
    "wat": "ETH/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "is_median": true,
    "poke": {
      "expiration": 86400,
      "spread": 1
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "BTC/USD",
    "address": "0x9Af8fe1d0c9ED3f176Dd3559B6f4b6FeF3AAb83B",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "DAI/USD",
    "address": "0xf2dc732221e2b374eBBfd0023EF794c4432E66d8",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "DSR/RATE",
    "address": "0xbBC385C209bC4C8E00E3687B51E25E21b0E7B186",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "ETH/USD",
    "address": "0x1174948681bb05748E3682398d9b7a6836B07554",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "MATIC/USD",
    "address": "0xD8569712fc3d447004524896010d4a2FB98C0ef7",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "SDAI/DAI",
    "address": "0xfFcF8e5A12Acc48870D2e8834310aa270dE10fE6",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "SDAI/ETH",
    "address": "0xE6DF058512F99c0C8940736687aDdb38722c73C0",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "SDAI/MATIC",
    "address": "0x6c9571D1dD3e606Ce734Cc558bdd0BE576E01660",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "wat": "WSTETH/ETH",
    "address": "0x84A48F89D5844385C515f43797147D6aF61cE2AE",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "wat": "BTC/USD",
    "address": "0x490d05d7eF82816F47737c7d72D10f5C172e7772",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "arb-goerli",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "AAVE/USD",
    "address": "0xa38C2B5408Eb1DCeeDBEC5d61BeD580589C6e717",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "ARB/USD",
    "address": "0x579BfD0581beD0d18fBb0Ebab099328d451552DD",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "AVAX/USD",
    "address": "0x78C8260AF7C8D0d17Cf3BA91F251E9375A389688",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "BNB/USD",
    "address": "0x26EE3E8b618227C1B735D8D884d52A852410019f",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "BTC/USD",
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "CRV/USD",
    "address": "0xf29a932ae56bB96CcACF8d1f2Da9028B01c8F030",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "DAI/USD",
    "address": "0xa7aA6a860D17A89810dE6e6278c58EB21Fa00fc4",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "DSR/RATE",
    "address": "0x729af3A41AE9E707e7AE421569C4b9c632B66a0c",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "expiration": 32400,
      "spread": 1
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "ETH/BTC",
    "address": "0x1804969b296E89C1ddB1712fA99816446956637e",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "ETH/USD",
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "GNO/USD",
    "address": "0xA28dCaB66FD25c668aCC7f232aa71DA1943E04b8",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "IBTA/USD",
    "address": "0x07487b0Bf28801ECD15BF09C13e32FBc87572e81",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "LDO/USD",
    "address": "0xa53dc5B100f0e4aB593f2D8EcD3c5932EE38215E",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "LINK/USD",
    "address": "0xecB89B57A60ac44E06ab1B767947c19b236760c3",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "MATIC/USD",
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "MKR/USD",
    "address": "0x67ffF0C6abD2a36272870B1E8FE42CC8E8D5ec4d",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "OP/USD",
    "address": "0xfadF055f6333a4ab435D2D248aEe6617345A4782",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "RETH/USD",
    "address": "0xEE02370baC10b3AC3f2e9eebBf8f3feA1228D263",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "SDAI/DAI",
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "SDAI/ETH",
    "address": "0x05aB94eD168b5d18B667cFcbbA795789C750D893",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "expiration": 32400,
      "spread": 1
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "SDAI/MATIC",
    "address": "0x2f0e0dE1F8c11d2380dE093ED15cA6cE07653cbA",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "expiration": 32400,
      "spread": 1
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "SNX/USD",
    "address": "0xD20f1eC72bA46b6126F96c5a91b6D3372242cE98",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "SOL/USD",
    "address": "0x4D1e6f39bbfcce8b471171b8431609b83f3a096D",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "UNI/USD",
    "address": "0x2aFF768F5d6FC63fA456B062e02f2049712a1ED5",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "USDC/USD",
    "address": "0x1173da1811a311234e7Ab0A33B4B7B646Ff42aEC",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "USDT/USD",
    "address": "0x0bd446021Ab95a2ABd638813f9bDE4fED3a5779a",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "WBTC/USD",
    "address": "0xA7226d85CE5F0DE97DCcBDBfD38634D6391d0584",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "WSTETH/USD",
    "address": "0xc9Bb81d3668f03ec9109bBca77d32423DeccF9Ab",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gno",
    "wat": "YFI/USD",
    "address": "0x0893EcE705639112C1871DcE88D87D81540D0199",
    "chain_id": 100,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "BTC/USD",
    "address": "0x586409bb88cF89BBAB0e106b0620241a0e4005c9",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "ETH/BTC",
    "address": "0xaF495008d177a2E2AD95125b78ace62ef61Ed1f7",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "ETH/USD",
    "address": "0xD81834Aa83504F6614caE3592fb033e4b8130380",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "GNO/USD",
    "address": "0x0cd01b018C355a60B2Cc68A1e3d53853f05A7280",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "IBTA/USD",
    "address": "0x0Aca91081B180Ad76a848788FC76A089fB5ADA72",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 10
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "LINK/USD",
    "address": "0xe4919256D404968566cbdc5E5415c769D5EeBcb0",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "MATIC/USD",
    "address": "0x4b4e2A0b7a560290280F083c8b5174FB706D7926",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "MKR/USD",
    "address": "0x496C851B2A9567DfEeE0ACBf04365F3ba00Eb8dC",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "RETH/USD",
    "address": "0x7eEE7e44055B6ddB65c6C970B061EC03365FADB3",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "WSTETH/USD",
    "address": "0x9466e1ffA153a8BdBB5972a7217945eb2E28721f",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "gor",
    "wat": "YFI/USD",
    "address": "0x38D27Ba21E1B2995d0ff9C1C070c5c93dd07cB31",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "wat": "BTC/USD",
    "address": "0x1aBBA7EA800f9023Fa4D1F8F840000bE7e3469a1",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "ogor",
    "wat": "ETH/USD",
    "address": "0xBBF1a875B13E4614645934faA3FEE59258320415",
    "is_median": true,
    "poke": {
      "expiration": 14400,
      "spread": 3
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "AAVE/USD",
    "address": "0xED4C91FC28B48E2Cf98b59668408EAeE44665511",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "ARB/USD",
    "address": "0x7dE6Df8E4c057eD9baE215F347A0339D603B09B2",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "AVAX/USD",
    "address": "0xD419f76594d411BD94c71FB0a78c80f71A2290Ce",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "BNB/USD",
    "address": "0x6931FB9C54958f77873ceC4536EaC56F561d2dC4",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "BTC/USD",
    "address": "0xdD5232e76798BEACB69eC310d9b0864b56dD08dD",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "CRV/USD",
    "address": "0x7B6E473f1CeB8b7100C9F7d58879e7211Bc48f32",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "DAI/USD",
    "address": "0x16984396EE0903782Ba8e6ebfA7DD356B0cA3841",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 0.5,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "ETH/BTC",
    "address": "0x4E866Ac929374096Afc2715C4e9c40D581A4067e",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "ETH/USD",
    "address": "0x90430C5b8045a1E2A0Fc4e959542a0c75b576439",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "GNO/USD",
    "address": "0xBcC6BFFde7888A3008f17c88D5a5e5F0D7462cf9",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "IBTA/USD",
    "address": "0xc52539EfbA58a521d69494D86fc47b9E71D32997",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "LDO/USD",
    "address": "0x3aeF92049C9401094A9f75259430F4771143F0C3",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "LINK/USD",
    "address": "0x4EDdF05CfAd20f1E39ed4CB067bdfa831dAeA9fE",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "MATIC/USD",
    "address": "0x06997AadB30d51eAdBAA7836f7a0F177474fc235",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "MKR/USD",
    "address": "0xE61A66f737c32d5Ac8cDea6982635B80447e9404",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "OP/USD",
    "address": "0x1Ae491D618A667a44D48E0b0BE2Cc0cDBF269BC5",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "RETH/USD",
    "address": "0xEff79d34f24Bb36eD8FB6c4CbaD5De293fdCf66F",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "SDAI/DAI",
    "address": "0xB6EE756124e88e12585981DdDa9E6E3bf3C4487D",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 0.5,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 0.5,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "SNX/USD",
    "address": "0x6Ab51f7E684923CE051e784D382A470b0fa834Be",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "SOL/USD",
    "address": "0x11ceEcca4d49f596E0Df781Af237CDE741ad2106",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "UNI/USD",
    "address": "0xfE051Bc90D3a2a825fA5172181f9124f8541838c",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "USDC/USD",
    "address": "0xfef7a1Eb17A095E1bd7723cBB1092caba34f9b1C",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 0.5,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "USDT/USD",
    "address": "0xF78A4e093Cd2D9F57Bb363Cc4edEBcf9bF3325ba",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 0.5,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "WBTC/USD",
    "address": "0x39C899178F4310705b12888886884b361CeF26C7",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "WSTETH/ETH",
    "address": "0x67E93d37B57747686F22f2F2f0a8aAd253199B38",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "WSTETH/USD",
    "address": "0x8Ba43F8Fa2fC13D7EEDCeb9414CDbB6643483C34",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "wat": "YFI/USD",
    "address": "0x16978358A8D6C7C8cA758F433685A5E8D988dfD4",
    "chain_id": 11155111,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 3600,
    "poke": {
      "spread": 2,
      "expiration": 86400
    },
    "poke_optimistic": {
      "spread": 1,
      "expiration": 43200
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "AAVE/USD",
    "address": "0xED4C91FC28B48E2Cf98b59668408EAeE44665511",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "ARB/USD",
    "address": "0x7dE6Df8E4c057eD9baE215F347A0339D603B09B2",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "AVAX/USD",
    "address": "0xD419f76594d411BD94c71FB0a78c80f71A2290Ce",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "BNB/USD",
    "address": "0x6931FB9C54958f77873ceC4536EaC56F561d2dC4",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "BTC/USD",
    "address": "0xdD5232e76798BEACB69eC310d9b0864b56dD08dD",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "CRV/USD",
    "address": "0x7B6E473f1CeB8b7100C9F7d58879e7211Bc48f32",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "DAI/USD",
    "address": "0x16984396EE0903782Ba8e6ebfA7DD356B0cA3841",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "ETH/BTC",
    "address": "0x4E866Ac929374096Afc2715C4e9c40D581A4067e",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "ETH/USD",
    "address": "0x90430C5b8045a1E2A0Fc4e959542a0c75b576439",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "GNO/USD",
    "address": "0xBcC6BFFde7888A3008f17c88D5a5e5F0D7462cf9",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "IBTA/USD",
    "address": "0xc52539EfbA58a521d69494D86fc47b9E71D32997",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "LDO/USD",
    "address": "0x3aeF92049C9401094A9f75259430F4771143F0C3",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "LINK/USD",
    "address": "0x4EDdF05CfAd20f1E39ed4CB067bdfa831dAeA9fE",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "MATIC/USD",
    "address": "0x06997AadB30d51eAdBAA7836f7a0F177474fc235",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "MKR/USD",
    "address": "0xE61A66f737c32d5Ac8cDea6982635B80447e9404",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "OP/USD",
    "address": "0x1ae491d618a667a44d48e0b0be2cc0cdbf269bc5",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "RETH/USD",
    "address": "0xEff79d34f24Bb36eD8FB6c4CbaD5De293fdCf66F",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "SDAI/DAI",
    "address": "0xB6EE756124e88e12585981DdDa9E6E3bf3C4487D",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "SNX/USD",
    "address": "0x6Ab51f7E684923CE051e784D382A470b0fa834Be",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "SOL/USD",
    "address": "0x11ceEcca4d49f596E0Df781Af237CDE741ad2106",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "UNI/USD",
    "address": "0xfE051Bc90D3a2a825fA5172181f9124f8541838c",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "USDC/USD",
    "address": "0xfef7a1Eb17A095E1bd7723cBB1092caba34f9b1C",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "USDT/USD",
    "address": "0xF78A4e093Cd2D9F57Bb363Cc4edEBcf9bF3325ba",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 0.5,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "WBTC/USD",
    "address": "0x39C899178F4310705b12888886884b361CeF26C7",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "WSTETH/ETH",
    "address": "0x67E93d37B57747686F22f2F2f0a8aAd253199B38",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "WSTETH/USD",
    "address": "0x8Ba43F8Fa2fC13D7EEDCeb9414CDbB6643483C34",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "YFI/USD",
    "address": "0x16978358A8D6C7C8cA758F433685A5E8D988dfD4",
    "chain_id": 1442,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "BTC/USD",
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "DSR/RATE",
    "address": "0x729af3A41AE9E707e7AE421569C4b9c632B66a0c",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "ETH/USD",
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "MATIC/USD",
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "SDAI/DAI",
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "SDAI/ETH",
    "address": "0x05aB94eD168b5d18B667cFcbbA795789C750D893",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "wat": "SDAI/MATIC",
    "address": "0x2f0e0dE1F8c11d2380dE093ED15cA6cE07653cbA",
    "chain_id": 1101,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 32400
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
