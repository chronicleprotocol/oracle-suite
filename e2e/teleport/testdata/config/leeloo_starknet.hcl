feeds = [
  "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
  "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
]

ethereum {
  client "default" {
    rpc_urls = ["http://localhost:8080"]
  }

  key "default" {
    address         = "2d800d93b065ce011af83f316cef9f0d005b0aa4"
    keystore_path   = "./e2e/teleport/testdata/keys"
    passphrase_file = "./e2e/teleport/testdata/keys/pass"
  }
}

transport {
  libp2p {
    priv_key_seed   = "2d800d93b065ce011af83f316cef9f0d005b0aa42d800d93b065ce011af83f31"
    listen_addrs    = ["/ip4/0.0.0.0/tcp/30101"]
    bootstrap_addrs = ["/ip4/127.0.0.1/tcp/30100/p2p/12D3KooWSGCRPjd6dHHjfWYeKnurLcaSYAQsQqDYj7GcPN2uhdis"]
    ethereum_key    = "default"
  }
}

leeloo {
  ethereum_key = "default"
  teleport_starknet {
    sequencer       = "http://localhost:8080"
    contract_addrs  = ["0x197f9e93cfaf7068ca2daf3ec89c2b91d051505c2231a0a0b9f70801a91fb24"]
    interval        = 10
    prefetch_period = 0
    replay_after    = []
  }
}
