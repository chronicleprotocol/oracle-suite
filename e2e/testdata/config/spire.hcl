variables {
  feeds = [
    "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
    "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
  ]
}

transport {
  libp2p {
    feeds         = var.feeds
    priv_key_seed = "8c8eba62d853d3abdd7f3298341a622a8a9df37c3aba788028c646bdd915227c"
    listen_addrs  = ["/ip4/0.0.0.0/tcp/30100"]
  }
}

ethereum {
  rand_keys = ["default"]
}

spire {
  ethereum_key    = "default"
  rpc_listen_addr = "127.0.0.1:9101"
  rpc_agent_addr  = "127.0.0.1:9101"
  feeds           = var.feeds
  pairs           = [
    "BTC/USD",
    "ETH/BTC",
  ]
}
