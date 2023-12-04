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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "bar": 9,
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
    "address": "0xED4C91FC28B48E2Cf98b59668408EAeE44665511",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x7dE6Df8E4c057eD9baE215F347A0339D603B09B2",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xD419f76594d411BD94c71FB0a78c80f71A2290Ce",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x6931FB9C54958f77873ceC4536EaC56F561d2dC4",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xdD5232e76798BEACB69eC310d9b0864b56dD08dD",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x7B6E473f1CeB8b7100C9F7d58879e7211Bc48f32",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x16984396EE0903782Ba8e6ebfA7DD356B0cA3841",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x09f3BfC6b46526045De5F5BE64f5CCe121bbf8B3",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x4E866Ac929374096Afc2715C4e9c40D581A4067e",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x90430C5b8045a1E2A0Fc4e959542a0c75b576439",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xBcC6BFFde7888A3008f17c88D5a5e5F0D7462cf9",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xc52539EfbA58a521d69494D86fc47b9E71D32997",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x3aeF92049C9401094A9f75259430F4771143F0C3",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x4EDdF05CfAd20f1E39ed4CB067bdfa831dAeA9fE",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x06997AadB30d51eAdBAA7836f7a0F177474fc235",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xE61A66f737c32d5Ac8cDea6982635B80447e9404",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x1Ae491D618A667a44D48E0b0BE2Cc0cDBF269BC5",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xEff79d34f24Bb36eD8FB6c4CbaD5De293fdCf66F",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xB6EE756124e88e12585981DdDa9E6E3bf3C4487D",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x20A32F633c1D26fC42A15dc7e6bd12Bf0468cAb1",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x0A154ec276972dBFEA01b13711408Ea6e72Ac36B",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x6Ab51f7E684923CE051e784D382A470b0fa834Be",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x11ceEcca4d49f596E0Df781Af237CDE741ad2106",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xfE051Bc90D3a2a825fA5172181f9124f8541838c",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xfef7a1Eb17A095E1bd7723cBB1092caba34f9b1C",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0xF78A4e093Cd2D9F57Bb363Cc4edEBcf9bF3325ba",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x39C899178F4310705b12888886884b361CeF26C7",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x8Ba43F8Fa2fC13D7EEDCeb9414CDbB6643483C34",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x16978358A8D6C7C8cA758F433685A5E8D988dfD4",
    "chain_id": 100,
    "bar": 3,
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
    "address": "0x3F982a82B4B6bd09b1DAF832140F166b595FEF7F",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "ARB/USD",
    "address": "0x9Bf0C1ba75C9d7b6Bf051cc7f7dCC7bfE5274302",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "AVAX/USD",
    "address": "0x7F56CdaAdB1c5230Fcab3E20D3A15BDE26cb6C2b",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "BNB/USD",
    "address": "0xE4A1EED38F972d05794C740Eae965A7Daa6Ab28c",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "BTC/USD",
    "address": "0x6edF073c4Bd934d3916AA6dDAC4255ccB2b7c0f0",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "CRV/USD",
    "address": "0xDcda58cAAC639C20aed270859109f03E9832a13A",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "DAI/USD",
    "address": "0xaf900d10f197762794C41dac395C5b8112eD13E1",
    "chain_id": 11155111,
    "bar": 3,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
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
    "address": "0xf95d3B8Ae567F4AA9BEC822931976c117cdf836a",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "ETH/USD",
    "address": "0xdd6D76262Fd7BdDe428dcfCd94386EbAe0151603",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "GNO/USD",
    "address": "0x9C9e56AE479f82bcF229F2200420106C93C0A24e",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "IBTA/USD",
    "address": "0x92b7Ab73BA53Bc64b57194242e3a36A6F1209A70",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "LDO/USD",
    "address": "0x4cD2a8c3Fd6329029461A95784051A553f31eb29",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "LINK/USD",
    "address": "0x260c182f0054BF244a8e38d7C475b6d9f67AeAc1",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "MATIC/USD",
    "address": "0xEa00861Dc00eBd246F6E51E52c28aBd9062bc09F",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "MKR/USD",
    "address": "0xE55afC31AFA140597c581Bc32057BF393ba97c5A",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "OP/USD",
    "address": "0x1Be54a524226fc44565747FE221157f4cAE71B80",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "RETH/USD",
    "address": "0x6454753E0909E7F6476BfB78BD6BDC281197A5be",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "SDAI/DAI",
    "address": "0x0B20Fd1c09452FC3F214667073EA8975aB2c55EA",
    "chain_id": 11155111,
    "bar": 3,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
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
    "address": "0x1eFD788C634C59e2c7507b523B3eEfD6CaaE0c4f",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "SOL/USD",
    "address": "0x39eC7D193D1Aa282b8ecCAC9B791b09c75D30491",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "UNI/USD",
    "address": "0x0E9e54244F6585a71d0d1035E7625849B516C817",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "USDC/USD",
    "address": "0xb34d784dc8E7cD240Fe1F318e282dFdD13C389AC",
    "chain_id": 11155111,
    "bar": 3,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
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
    "address": "0x8c852EEC6ae356FeDf5d7b824E254f7d94Ac6824",
    "chain_id": 11155111,
    "bar": 3,
    "is_scribe": true,
    "is_scribe_optimistic": true,
    "challenge_period": 1200,
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
    "address": "0xdc3ef3E31AdAe791d9D5054B575f7396851Fa432",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "WSTETH/ETH",
    "address": "0x2d95B1862279771fcE76823CD777384D8598fB48",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "WSTETH/USD",
    "address": "0x89822dd9D74dF50BFba8764DC9bE25E9B8d554A1",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "sep",
    "wat": "YFI/USD",
    "address": "0xdF54aBf0eF88aB7fFf22e21eDD9AE1DA89A7DefC",
    "chain_id": 11155111,
    "bar": 3,
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
    "env": "stage",
    "chain": "testnet-zkEVM-mango",
    "wat": "AAVE/USD",
    "address": "0x3F982a82B4B6bd09b1DAF832140F166b595FEF7F",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x9bf0c1ba75c9d7b6bf051cc7f7dcc7bfe5274302",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xDcd4c95f9D1f660E31fD516B936388fc9D4117Ea",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xE4A1EED38F972d05794C740Eae965A7Daa6Ab28c",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x6edF073c4Bd934d3916AA6dDAC4255ccB2b7c0f0",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xDcda58cAAC639C20aed270859109f03E9832a13A",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xaf900d10f197762794C41dac395C5b8112eD13E1",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xf95d3B8Ae567F4AA9BEC822931976c117cdf836a",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xdd6D76262Fd7BdDe428dcfCd94386EbAe0151603",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x9C9e56AE479f82bcF229F2200420106C93C0A24e",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x92b7Ab73BA53Bc64b57194242e3a36A6F1209A70",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x4cD2a8c3Fd6329029461A95784051A553f31eb29",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x260c182f0054BF244a8e38d7C475b6d9f67AeAc1",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xEa00861Dc00eBd246F6E51E52c28aBd9062bc09F",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xE55afC31AFA140597c581Bc32057BF393ba97c5A",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x1Be54a524226fc44565747FE221157f4cAE71B80",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x6454753E0909E7F6476BfB78BD6BDC281197A5be",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x0B20Fd1c09452FC3F214667073EA8975aB2c55EA",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x1eFD788C634C59e2c7507b523B3eEfD6CaaE0c4f",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x39eC7D193D1Aa282b8ecCAC9B791b09c75D30491",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x0E9e54244F6585a71d0d1035E7625849B516C817",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xb34d784dc8E7cD240Fe1F318e282dFdD13C389AC",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x8c852EEC6ae356FeDf5d7b824E254f7d94Ac6824",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xdc3ef3E31AdAe791d9D5054B575f7396851Fa432",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x2d95B1862279771fcE76823CD777384D8598fB48",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0x89822dd9D74dF50BFba8764DC9bE25E9B8d554A1",
    "chain_id": 1442,
    "bar": 3,
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
    "address": "0xdF54aBf0eF88aB7fFf22e21eDD9AE1DA89A7DefC",
    "chain_id": 1442,
    "bar": 3,
    "is_scribe": true,
    "is_scribe_optimistic": false,
    "poke": {
      "spread": 1,
      "expiration": 21600
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
variables {
  environment    = env("CFG_ENVIRONMENT", "prod")
  item_separator = env("CFG_ITEM_SEPARATOR", "\n")
  feeds          = try(var.feed_sets[env("CFG_FEEDS", var.environment)], explode(var.item_separator, env("CFG_FEEDS", "")))

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

  static_address_books = {
    "prod" : [
      "66thskfs35yclgmvmp3z47vaewo62vedzdwoboygm7bn5s7m7paa6cqd.onion:8888",
    ]
    "stage" : [
      "cqsdvjamh6vh5bmavgv6hdb5rrhjqgqtqzy6cfgbmzqhpxfrppblupqd.onion:8888",
    ]
  }

  libp2p_bootstraps = {
    "prod" : [
      "/dns/spire-bootstrap1.chroniclelabs.io/tcp/8000/p2p/12D3KooWFYkJ1SghY4KfAkZY9Exemqwnh4e4cmJPurrQ8iqy2wJG",
      "/dns/spire-bootstrap2.chroniclelabs.io/tcp/8000/p2p/12D3KooWD7eojGbXT1LuqUZLoewRuhNzCE2xQVPHXNhAEJpiThYj",
    ]
    "stage" : [
      "/dns/spire-bootstrap1.staging.chroniclelabs.io/tcp/8000/p2p/12D3KooWHoSyTgntm77sXShoeX9uNkqKNMhHxKtskaHqnA54SrSG",
      "/ip4/178.128.141.30/tcp/8000/p2p/12D3KooWLaMPReGaxFc6Z7BKWTxZRbxt3ievW8Np7fpA6y774W9T",
    ]
  }
}
variables {
  chain_rpc_urls = explode(var.item_separator, env("CFG_CHAIN_RPC_URLS", env("CFG_RPC_URLS", "")))
  chain_name     = env("CFG_CHAIN_NAME", "eth")

  # RPC URLs for specific blockchain clients. Gofer is chain type aware.
  # See: config-gofer.hcl: origin.<name>.contracts.<client>
  eth_rpc_urls = explode(var.item_separator, env("CFG_ETH_RPC_URLS", "https://eth.public-rpc.com"))
  arb_rpc_urls = explode(var.item_separator, env("CFG_ARB_RPC_URLS", ""))
  opt_rpc_urls = explode(var.item_separator, env("CFG_OPT_RPC_URLS", ""))
}

ethereum {
  # Labels for generating random ethereum keys on every app boot.
  # The labels are used to reference ethereum keys in other sections.
  # (optional)
  #
  # If you want to use a specific key, you can set the CFG_ETH_FROM
  # environment variable along with CFG_ETH_KEYS and CFG_ETH_PASS.
  rand_keys = env("CFG_ETH_FROM", "") == "" ? ["default"] : []

  dynamic "key" {
    for_each = env("CFG_ETH_FROM", "") == "" ? [] : [1]
    labels   = ["default"]
    content {
      address         = env("CFG_ETH_FROM", "")
      keystore_path   = env("CFG_ETH_KEYS", "")
      passphrase_file = env("CFG_ETH_PASS", "")
    }
  }

  dynamic "client" {
    for_each = length(var.chain_rpc_urls) == 0 ? [] : [1]
    labels   = ["default"]
    content {
      rpc_urls                    = var.chain_rpc_urls
      chain_id                    = tonumber(env("CFG_CHAIN_ID", "1"))
      ethereum_key                = "default"
      tx_type                     = env("CFG_CHAIN_TX_TYPE", "eip1559")
      gas_priority_fee_multiplier = tonumber(env("CFG_CHAIN_GAS_FEE_MULTIPLIER", "1"))
      gas_fee_multiplier          = tonumber(env("CFG_CHAIN_GAS_PRIORITY_FEE_MULTIPLIER", "1"))
      max_gas_fee                 = tonumber(env("CFG_CHAIN_MAX_GAS_FEE", "0"))
      max_gas_priority_fee        = tonumber(env("CFG_CHAIN_MAX_GAS_PRIORITY_FEE", "0"))
      max_gas_limit               = tonumber(env("CFG_CHAIN_MAX_GAS_LIMIT", "0"))
    }
  }
  dynamic "client" {
    for_each = length(var.eth_rpc_urls) == 0 ? [] : [1]
    labels   = ["ethereum"]
    content {
      rpc_urls     = var.eth_rpc_urls
      chain_id     = tonumber(env("CFG_ETH_CHAIN_ID", "1"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.arb_rpc_urls) == 0 ? [] : [1]
    labels   = ["arbitrum"]
    content {
      rpc_urls     = var.arb_rpc_urls
      chain_id     = tonumber(env("CFG_ARB_CHAIN_ID", "42161"))
      ethereum_key = "default"
    }
  }
  dynamic "client" {
    for_each = length(var.opt_rpc_urls) == 0 ? [] : [1]
    labels   = ["optimism"]
    content {
      rpc_urls     = var.opt_rpc_urls
      chain_id     = tonumber(env("CFG_OPT_CHAIN_ID", "10"))
      ethereum_key = "default"
    }
  }
}
variables {
  ghost_pairs = explode(var.item_separator, env("CFG_MODELS", env("CFG_GHOST_PAIRS", "")))
}

ghost {
  ethereum_key = "default"
  interval     = tonumber(env("CFG_GHOST_INTERVAL", "60"))
  data_models  = distinct(concat([
    for v in var.contracts : v.wat
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_scribe, false)
    # If CFG_GHOST_PAIRS is set to a list of asset symbols, only for those assets will the signatures be created
    && try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v.wat), false)
  ], [
    for v in var.contracts : replace(v.wat, "/", "")
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_median, false)
    # If CFG_GHOST_PAIRS is set to a list of asset symbols, only for those assets will the signatures be created
    && try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v.wat), false)
  ], [
    for v in var.models : v
    if try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v), false)
  ]))
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

  origin "dsr" {
    type = "dsr"
    contracts "ethereum" {
      addresses = {
        "DSR/RATE" = "0x197E90f9FAD81970bA7976f33CbD77088E5D7cf7" # address to pot contract
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

  origin "okx" {
    type = "tick_generic_jq"
    url  = "https://www.okx.com/api/v5/market/ticker?instId=$${ucbase}-$${ucquote}&instType=SPOT"
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
        "MATIC/WETH"  = "0x290A6a7460B308ee3F19023D2D00dE604bcf5B42"
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
      min_values = 2
      alias "RETH/ETH" {
        origin "balancerV2" { query = "RETH/WETH" }
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
        origin "balancerV2" { query = "WSTETH/WETH" }
      }
      indirect {
        origin "curve" { query = "RETH/WSTETH" }
        reference { data_model = "RETH/ETH" }
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

variables {
  spectre_pairs = explode(var.item_separator, env("CFG_MODELS", env("CFG_SPECTRE_PAIRS", "")))
}

spectre {
  dynamic "median" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.is_median, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = replace(contract.value.wat, "/", "")

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration
    }
  }

  dynamic "scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.is_scribe, false) && !try(v.is_scribe_optimistic, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration
    }
  }

  dynamic "optimistic_scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.is_scribe, false) && try(v.is_scribe_optimistic, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration

      # Spread in percent points above which the price is considered stale.
      optimistic_spread = contract.value.poke_optimistic.spread

      # Time in seconds after which the price is considered stale.
      optimistic_expiration = contract.value.poke_optimistic.expiration
    }
  }
}
variables {
  spire_keys = explode(var.item_separator, env("CFG_MODELS", env("CFG_SPIRE_KEYS", "")))
}

spire {
  # Ethereum key to use for signing messages. The key must be present in the `ethereum` section.
  # (optional) if not set, the first key in the `ethereum` section is used.
  ethereum_key = "default"

  rpc_listen_addr = env("CFG_SPIRE_RPC_ADDR", ":9100")
  rpc_agent_addr  = env("CFG_SPIRE_RPC_ADDR", "127.0.0.1:9100")

  # List of pairs that are collected by the spire node. Other pairs are ignored.
  pairs = distinct(concat([
    for v in var.contracts : v.wat
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_scribe, false)
    # If CFG_SPIRE_KEYS is set to a list of asset symbols
    && try(length(var.spire_keys) == 0 || contains(var.spire_keys, v.wat), false)
  ], [
    for v in var.contracts : replace(v.wat, "/", "")
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Median compatible contracts
    && try(v.is_median, false)
    # If CFG_SPIRE_KEYS is set to a list of asset symbols
    && try(length(var.spire_keys) == 0 || contains(var.spire_keys, v.wat), false)
  ]))

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = var.feeds
}
variables {
  libp2p_enable          = tobool(env("CFG_LIBP2P_ENABLE", "1"))
  libp2p_bootstrap_addrs = explode(var.item_separator, env("CFG_LIBP2P_BOOTSTRAP_ADDRS", join(
    var.item_separator,
    try(var.libp2p_bootstraps[var.environment], [])
  )))

  webapi_enable              = tobool(env("CFG_WEBAPI_ENABLE", "0"))
  webapi_eth_address_book    = env("CFG_WEBAPI_ETH_ADDR_BOOK", try(var.contract_map["${var.environment}-${var.chain_name}-TorAddressRegister"], ""))
  webapi_static_address_book = explode(var.item_separator, env("CFG_WEBAPI_STATIC_ADDR_BOOK", join(
    var.item_separator,
    try(var.static_address_books[var.environment], [])
  )))
}

transport {
  # LibP2P transport configuration. Enabled if CFG_LIBP2P_ENABLE is set to anything evaluated to `false`.
  dynamic "libp2p" {
    for_each = var.libp2p_enable ? [1] : []
    content {
      feeds                = var.feeds
      feeds_filter_disable = tobool(env("CFG_LIBP2P_FEEDS_FILTER_DISABLE", "0"))
      priv_key_seed        = env("CFG_LIBP2P_PK_SEED", "")
      listen_addrs         = explode(var.item_separator, env("CFG_LIBP2P_LISTEN_ADDRS", "/ip4/0.0.0.0/tcp/8000"))
      bootstrap_addrs      = var.libp2p_bootstrap_addrs
      direct_peers_addrs   = explode(var.item_separator, env("CFG_LIBP2P_DIRECT_PEERS_ADDRS", ""))
      blocked_addrs        = explode(var.item_separator, env("CFG_LIBP2P_BLOCKED_ADDRS", ""))
      disable_discovery    = tobool(env("CFG_LIBP2P_DISABLE_DISCOVERY", "0"))
      ethereum_key         = "default"
      external_addr        = env("CFG_LIBP2P_EXTERNAL_ADDR", env("CFG_LIBP2P_EXTERNAL_IP", ""))
    }
  }

  # WebAPI transport configuration. Enabled if CFG_WEBAPI_LISTEN_ADDR is set to a listen address.
  dynamic "webapi" {
    for_each = var.webapi_enable ? [1] : []
    content {
      feeds             = var.feeds
      listen_addr       = env("CFG_WEBAPI_LISTEN_ADDR", "")
      socks5_proxy_addr = env("CFG_WEBAPI_SOCKS5_PROXY_ADDR", "")
      ethereum_key      = "default"

      # Ethereum based address book. Enabled if CFG_WEBAPI_ETH_ADDR_BOOK is set to a contract address.
      dynamic "ethereum_address_book" {
        for_each = var.webapi_eth_address_book == "" ? [] : [1]
        content {
          contract_addr   = var.webapi_eth_address_book
          ethereum_client = "default"
        }
      }

      # Static address book. Enabled if CFG_WEBAPI_STATIC_ADDR_BOOK is set.
      dynamic "static_address_book" {
        for_each = var.webapi_static_address_book =="" ? [] : [1]
        content {
          addresses = var.webapi_static_address_book
        }
      }
    }
  }
}
variables {
  musig_pairs = explode(var.item_separator, env("CFG_MODELS", env("CFG_MUSIG_PAIRS", "")))
}

info {
  web_url = env("CFG_WEB_URL", "")
}

musig {
  ethereum_key = "default"

  dynamic "tick" {
    for_each = [
      for v in var.models : v
      if try(length(var.musig_pairs) == 0 || contains(var.musig_pairs, v), false)
    ]
    iterator = contract
    labels   = [contract.value]
    content {
      ethereum_client = "default"

      chainlog_contract_addr     = env("CFG_MUSIG_CHAINLOG", var.contract_map["${var.environment}-${var.chain_name}-Chainlog"])
      wat_registry_contract_addr = env("CFG_MUSIG_WAT_REGISTRY", var.contract_map["${var.environment}-${var.chain_name}-WatRegistry"])

      interval = tonumber(env("CFG_MUSIG_INTERVAL", "600"))
      max_age  = tonumber(env("CFG_MUSIG_MAX_AGE", "3600"))

      optimistic = true
    }
  }
}
