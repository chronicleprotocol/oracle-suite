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
