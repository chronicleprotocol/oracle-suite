variables {
  webapi_enable            = tobool(env("CFG_WEBAPI_ENABLE", "0"))
  webapi_listen_addr       = env("CFG_WEBAPI_LISTEN_ADDR", "")
  webapi_socks5_proxy_addr = env("CFG_WEBAPI_SOCKS5_PROXY_ADDR", "") # will not try to connect to a proxy if empty
  webapi_static_addr_book  = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_WEBAPI_STATIC_ADDR_BOOK", "cqsdvjamh6vh5bmavgv6hdb5rrhjqgqtqzy6cfgbmzqhpxfrppblupqd.onion:8888"))

  libp2p_enable     = tobool(env("CFG_LIBP2P_ENABLE", "1"))
  libp2p_bootstraps = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_LIBP2P_BOOTSTRAP_ADDRS", join(env("CFG_ITEM_SEPARATOR", ","), [
    "/dns4/spire-bootstrap1.chroniclelabs.io/tcp/8000/p2p/12D3KooWFYkJ1SghY4KfAkZY9Exemqwnh4e4cmJPurrQ8iqy2wJG",
    "/dns4/spire-bootstrap2.chroniclelabs.io/tcp/8000/p2p/12D3KooWD7eojGbXT1LuqUZLoewRuhNzCE2xQVPHXNhAEJpiThYj",
  ])))
  libp2p_peers             = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_LIBP2P_DIRECT_PEERS_ADDRS", ""))
  libp2p_bans              = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_LIBP2P_BLOCKED_ADDRS", ""))
  libp2p_listens           = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_LIBP2P_LISTEN_ADDRS", "/ip4/0.0.0.0/tcp/8000"))
  libp2p_disable_discovery = tobool(env("CFG_LIBP2P_DISABLE_DISCOVERY", "0"))
}

transport {
  # LibP2P transport configuration. Enabled if CFG_LIBP2P_ENABLE is set to anything evaluated to `false`.
  dynamic "libp2p" {
    for_each = var.libp2p_enable ? [1] : []
    content {
      feeds              = try(var.feed_sets[env("CFG_FEEDS", var.environment)], explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_FEEDS", "")))
      priv_key_seed      = env("CFG_LIBP2P_PK_SEED", "")
      listen_addrs       = var.libp2p_listens
      bootstrap_addrs    = var.libp2p_bootstraps
      direct_peers_addrs = var.libp2p_peers
      blocked_addrs      = var.libp2p_bans
      disable_discovery  = var.libp2p_disable_discovery
      ethereum_key       = "default"
    }
  }

  # WebAPI transport configuration. Enabled if CFG_WEBAPI_LISTEN_ADDR is set to a listen address.
  dynamic "webapi" {
    for_each = var.webapi_enable ? [1] : []
    content {
      feeds             = try(var.feed_sets[env("CFG_FEEDS", var.environment)], explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_FEEDS", "")))
      listen_addr       = var.webapi_listen_addr
      socks5_proxy_addr = var.webapi_socks5_proxy_addr # will not try to connect to a proxy if empty
      ethereum_key      = "default"

      # Ethereum based address book. Enabled if CFG_WEBAPI_ETH_ADDR_BOOK is set to a contract address.
      dynamic "ethereum_address_book" {
        for_each = env("CFG_WEBAPI_ETH_ADDR_BOOK", try(var.contract_map["${var.environment}-${var.chain_name}-TorAddressRegister"], "")) == "" ? [] : [
          1
        ]
        content {
          contract_addr   = env("CFG_WEBAPI_ETH_ADDR_BOOK", try(var.contract_map["${var.environment}-${var.chain_name}-TorAddressRegister"], ""))
          ethereum_client = "default"
        }
      }

      # Static address book. Enabled if CFG_WEBAPI_STATIC_ADDR_BOOK is set.
      dynamic "static_address_book" {
        for_each = length(var.webapi_static_addr_book) == 0 ? [] : [1]
        content {
          addresses = var.webapi_static_addr_book
        }
      }
    }
  }
}
