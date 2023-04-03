rand_keys = ["key"]

# Without optionals
key "key1" {
  address         = "0x1234567890123456789012345678901234567890"
  keystore_path   = "./keystore"
}

# With optionals
key "key2" {
  address         = "0x2345678901234567890123456789012345678901"
  keystore_path   = "./keystore2"
  passphrase_file = "./passphrase"
}

# Without optionals
client "client1" {
  rpc_urls     = ["https://rpc1.example"]
  chain_id     = 1
  ethereum_key = "key1"
}

# With optionals
client "client2" {
  rpc_urls          = ["https://rpc2.example"]
  timeout           = 10
  graceful_timeout  = 5
  max_blocks_behind = 100
  ethereum_key      = "key2"
  chain_id          = 1
}
