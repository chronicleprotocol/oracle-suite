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
