rand_keys = ["rand_key"]

# Without optionals
key "key1" {
  address       = "0xd18d7f6d9e349d1d6bf33702192019f166a7201e"
  keystore_path = "./testdata/keystore"
}

# With optionals
key "key2" {
  address         = "0x2d800d93b065ce011af83f316cef9f0d005b0aa4"
  keystore_path   = "./testdata/keystore"
  passphrase_file = "./testdata/keystore/passphrase"
}

# Without optionals
client "client1" {
  rpc_urls     = ["https://rpc1.example"]
  chain_id     = 1
  ethereum_key = "key1"
}

# With optionals
client "client2" {
  rpc_urls                    = ["https://rpc2.example"]
  timeout                     = 10
  graceful_timeout            = 5
  max_blocks_behind           = 100
  ethereum_key                = "key2"
  chain_id                    = 1
  tx_type                     = "eip1559"
  gas_fee_multiplier          = 1.5
  gas_priority_fee_multiplier = 1.25
  max_gas_fee                 = 1000000000000
  max_gas_priority_fee        = 1000000000000
  max_gas_limit               = 10000000
}
