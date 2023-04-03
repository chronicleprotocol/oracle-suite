libp2p {
  feeds              = ["0x1234567890123456789012345678901234567890", "0x2345678901234567890123456789012345678901"]
  listen_addrs       = ["/ip4/0.0.0.0/tcp/6000"]
  priv_key_seed      = "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2"
  bootstrap_addrs    = ["/ip4/0.0.0.0/tcp/7000"]
  direct_peers_addrs = ["/ip4/0.0.0.0/tcp/8000"]
  blocked_addrs      = ["/ip4/0.0.0.0/tcp/9000"]
  disable_discovery  = true
  ethereum_key       = "eth_key"
}

webapi {
  feeds             = ["0x3456789012345678901234567890123456789012", "0x4567890123456789012345678901234567890123"]
  listen_addr       = "localhost:8080"
  socks5_proxy_addr = "localhost:9050"
  ethereum_key      = "eth_key"

  ethereum_address_book {
    contract_addr   = "0x5678901234567890123456789012345678901234"
    ethereum_client = "default"
  }

  static_address_book {
    addresses = ["https://example.com/api/v1/endpoint"]
  }
}
