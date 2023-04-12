variables {
  feeds = [
    "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
    "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
  ]
}

ethereum {
  client "default" {
    rpc_urls = ["http://localhost:8080"]
  }

  key "default" {
    address         = "2d800d93b065ce011af83f316cef9f0d005b0aa4"
    keystore_path   = "./e2e/testdata/keys"
    passphrase_file = "./e2e/testdata/keys/pass"
  }
}

transport {
  libp2p {
    feeds           = var.feeds
    priv_key_seed   = "2d800d93b065ce011af83f316cef9f0d005b0aa42d800d93b065ce011af83f31"
    listen_addrs    = ["/ip4/0.0.0.0/tcp/30101"]
    bootstrap_addrs = ["/ip4/127.0.0.1/tcp/30100/p2p/12D3KooWSGCRPjd6dHHjfWYeKnurLcaSYAQsQqDYj7GcPN2uhdis"]
    ethereum_key    = "default"
  }
}

leeloo {
  ethereum_key = "default"
  teleport_evm {
    ethereum_client     = "default"
    contract_addrs      = ["0x20265780907778b4d0e9431c8ba5c7f152707f1d"]
    interval            = 10
    prefetch_period     = 10
    block_confirmations = 0
    block_limit         = 1000
    replay_after        = [try(tonumber(env.REPLAY_AFTER), 10)]
  }
}
