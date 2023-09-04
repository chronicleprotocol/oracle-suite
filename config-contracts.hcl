variables {
  contracts = [
  {
    "env": "prod",
    "chain": "eth",
    "contract": "ScribeOptimistic",
    "address": "0x898D1aB819a24880F636416df7D1493C94143262",
    "i_scribe": {
      "wat": "BTC/USD",
      "bar": 2,
      "decimals": 18
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "contract": "ScribeOptimistic",
    "address": "0x5E16CA75000fb2B9d7B1184Fa24fF5D938a345Ef",
    "i_scribe": {
      "wat": "ETH/USD",
      "bar": 2,
      "decimals": 18
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "contract": "ScribeOptimistic",
    "address": "0xb400027B7C31D67982199Fa48B8228F128691fCb",
    "i_scribe": {
      "wat": "MKR/USD",
      "bar": 2,
      "decimals": 18
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "contract": "ScribeOptimistic",
    "address": "0x608D9cD5aC613EBAC4549E6b8A73954eA64C3660",
    "i_scribe": {
      "wat": "RETH/USD",
      "bar": 2,
      "decimals": 18
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "prod",
    "chain": "eth",
    "contract": "ScribeOptimistic",
    "address": "0x013C5C46db9914A19A58E57AD539eD5B125aFA15",
    "i_scribe": {
      "wat": "WSTETH/USD",
      "bar": 2,
      "decimals": 18
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x898D1aB819a24880F636416df7D1493C94143262",
    "i_scribe": {
      "wat": "BTC/USD",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xbBC385C209bC4C8E00E3687B51E25E21b0E7B186",
    "i_scribe": {
      "wat": "DSR/RATE",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x5E16CA75000fb2B9d7B1184Fa24fF5D938a345Ef",
    "i_scribe": {
      "wat": "ETH/USD",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xE0ECe625B1E128EE00e39BB91A80772D5d4d8Ed5",
    "i_scribe": {
      "wat": "MATIC/USD",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xfFcF8e5A12Acc48870D2e8834310aa270dE10fE6",
    "i_scribe": {
      "wat": "SDAI/DAI",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xE6DF058512F99c0C8940736687aDdb38722c73C0",
    "i_scribe": {
      "wat": "SDAI/ETH",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "prod",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x6c9571D1dD3e606Ce734Cc558bdd0BE576E01660",
    "i_scribe": {
      "wat": "SDAI/MATIC",
      "bar": 2,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xa38C2B5408Eb1DCeeDBEC5d61BeD580589C6e717",
    "i_scribe": {
      "wat": "AAVE/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x579BfD0581beD0d18fBb0Ebab099328d451552DD",
    "i_scribe": {
      "wat": "ARB/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x78C8260AF7C8D0d17Cf3BA91F251E9375A389688",
    "i_scribe": {
      "wat": "AVAX/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x26EE3E8b618227C1B735D8D884d52A852410019f",
    "i_scribe": {
      "wat": "BNB/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "i_scribe": {
      "wat": "BTC/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xf29a932ae56bB96CcACF8d1f2Da9028B01c8F030",
    "i_scribe": {
      "wat": "CRV/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xa7aA6a860D17A89810dE6e6278c58EB21Fa00fc4",
    "i_scribe": {
      "wat": "DAI/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x1804969b296E89C1ddB1712fA99816446956637e",
    "i_scribe": {
      "wat": "ETH/BTC",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "i_scribe": {
      "wat": "ETH/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xA28dCaB66FD25c668aCC7f232aa71DA1943E04b8",
    "i_scribe": {
      "wat": "GNO/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x07487b0Bf28801ECD15BF09C13e32FBc87572e81",
    "i_scribe": {
      "wat": "IBTA/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xa53dc5B100f0e4aB593f2D8EcD3c5932EE38215E",
    "i_scribe": {
      "wat": "LDO/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xecB89B57A60ac44E06ab1B767947c19b236760c3",
    "i_scribe": {
      "wat": "LINK/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "i_scribe": {
      "wat": "MATIC/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x67ffF0C6abD2a36272870B1E8FE42CC8E8D5ec4d",
    "i_scribe": {
      "wat": "MKR/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xfadF055f6333a4ab435D2D248aEe6617345A4782",
    "i_scribe": {
      "wat": "OP/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xEE02370baC10b3AC3f2e9eebBf8f3feA1228D263",
    "i_scribe": {
      "wat": "RETH/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "i_scribe": {
      "wat": "SDAI/DAI",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xD20f1eC72bA46b6126F96c5a91b6D3372242cE98",
    "i_scribe": {
      "wat": "SNX/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x4D1e6f39bbfcce8b471171b8431609b83f3a096D",
    "i_scribe": {
      "wat": "SOL/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x2aFF768F5d6FC63fA456B062e02f2049712a1ED5",
    "i_scribe": {
      "wat": "UNI/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x1173da1811a311234e7Ab0A33B4B7B646Ff42aEC",
    "i_scribe": {
      "wat": "USDC/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x0bd446021Ab95a2ABd638813f9bDE4fED3a5779a",
    "i_scribe": {
      "wat": "USDT/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xA7226d85CE5F0DE97DCcBDBfD38634D6391d0584",
    "i_scribe": {
      "wat": "WBTC/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0xc9Bb81d3668f03ec9109bBca77d32423DeccF9Ab",
    "i_scribe": {
      "wat": "WSTETH/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "sep",
    "contract": "ScribeOptimistic",
    "address": "0x0893EcE705639112C1871DcE88D87D81540D0199",
    "i_scribe": {
      "wat": "YFI/USD",
      "bar": 3,
      "decimals": 18,
      "indexes": {
        "0x75FBD0aaCe74Fb05ef0F6C0AC63d26071Eb750c9": 1,
        "0x5C01f0F08E54B85f4CaB8C6a03c9425196fe66DD": 2,
        "0xC50DF8b5dcb701aBc0D6d1C7C99E6602171Abbc4": 3,
        "0x0c4FC7D66b7b6c684488c1F218caA18D4082da18": 4
      }
    },
    "i_scribe_optimistic": {
      "challenge_period": 3600
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x4B5aBFC0Fe78233b97C80b8410681765ED9fC29c",
    "i_scribe": {
      "wat": "BTC/USD",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x729af3A41AE9E707e7AE421569C4b9c632B66a0c",
    "i_scribe": {
      "wat": "DSR/RATE",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xc8A1F9461115EF3C1E84Da6515A88Ea49CA97660",
    "i_scribe": {
      "wat": "ETH/USD",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xa48c56e48A71966676d0D113EAEbe6BE61661F18",
    "i_scribe": {
      "wat": "MATIC/USD",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0xD93c56Aa71923228cDbE2be3bf5a83bF25B0C491",
    "i_scribe": {
      "wat": "SDAI/DAI",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x05aB94eD168b5d18B667cFcbbA795789C750D893",
    "i_scribe": {
      "wat": "SDAI/ETH",
      "bar": 3,
      "decimals": 18
    }
  },
  {
    "env": "stage",
    "chain": "zkevm",
    "contract": "Scribe",
    "address": "0x2f0e0dE1F8c11d2380dE093ED15cA6cE07653cbA",
    "i_scribe": {
      "wat": "SDAI/MATIC",
      "bar": 3,
      "decimals": 18
    }
  }
]
}
