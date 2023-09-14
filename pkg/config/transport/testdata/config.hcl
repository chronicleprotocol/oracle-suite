libp2p {
  feeds              = ["0x1234567890123456789012345678901234567890", "0x2345678901234567890123456789012345678901"]
  listen_addrs       = ["/ip4/0.0.0.0/tcp/6000"]
  priv_key_seed      = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
  bootstrap_addrs    = ["/ip4/0.0.0.0/tcp/7000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi"]
  direct_peers_addrs = ["/ip4/0.0.0.0/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi"]
  blocked_addrs      = ["/ip4/0.0.0.0/tcp/9000"]
  disable_discovery  = true
  ethereum_key       = "key"
  external_ip        = "10.1.2.3"
}

webapi {
  feeds             = ["0x3456789012345678901234567890123456789012", "0x4567890123456789012345678901234567890123"]
  listen_addr       = "localhost:8080"
  socks5_proxy_addr = "localhost:9050"
  ethereum_key      = "key"

  ethereum_address_book {
    contract_addr   = "0x5678901234567890123456789012345678901234"
    ethereum_client = "client"
  }

  static_address_book {
    addresses = ["https://example.com/api/v1/endpoint"]
  }
}
