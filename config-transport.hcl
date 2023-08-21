variables {
  webapi_enable            = tobool(env("CFG_WEBAPI_ENABLE", "1"))
  webapi_listen_addr       = env("CFG_WEBAPI_LISTEN_ADDR", "0.0.0.0:8080")
  webapi_static_addr_book  = explode(",", env("CFG_WEBAPI_STATIC_ADDR_BOOK", "cqsdvjamh6vh5bmavgv6hdb5rrhjqgqtqzy6cfgbmzqhpxfrppblupqd.onion:8888"))
  webapi_eth_addr_book     = env("CFG_WEBAPI_ETH_ADDR_BOOK", "0xd51Fd30C873356b432F766eB55fc599586734a95")
  webapi_socks5_proxy_addr = env("CFG_WEBAPI_SOCKS5_PROXY_ADDR", "") # will not try to connect to a proxy if empty

  libp2p_enable     = tobool(env("CFG_LIBP2P_ENABLE", "1"))
  libp2p_bootstraps = explode(",", env("CFG_LIBP2P_BOOTSTRAP_ADDRS", "/dns/spire-bootstrap1.makerops.services/tcp/8000/p2p/12D3KooWRfYU5FaY9SmJcRD5Ku7c1XMBRqV6oM4nsnGQ1QRakSJi,/dns/spire-bootstrap2.makerops.services/tcp/8000/p2p/12D3KooWBGqjW4LuHUoYZUhbWW1PnDVRUvUEpc4qgWE3Yg9z1MoR"))
  libp2p_peers      = explode(",", env("CFG_LIBP2P_DIRECT_PEERS_ADDRS", ""))
  libp2p_bans       = explode(",", env("CFG_LIBP2P_BLOCKED_ADDRS", ""))
  libp2p_listens    = explode(",", env("CFG_LIBP2P_LISTEN_ADDRS", "/ip4/0.0.0.0/tcp/8000"))
}

transport {
  # LibP2P transport configuration. Enabled if CFG_LIBP2P_ENABLE is set to anything evaluated to `false`.
  dynamic "libp2p" {
    for_each = var.libp2p_enable ? [1] : []
    content {
      feeds              = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(",", env("CFG_FEEDS", "")))
      priv_key_seed      = env("CFG_LIBP2P_PK_SEED", "")
      listen_addrs       = var.libp2p_listens
      bootstrap_addrs    = var.libp2p_bootstraps
      direct_peers_addrs = var.libp2p_peers
      blocked_addrs      = var.libp2p_bans
      disable_discovery  = tobool(env("CFG_LIBP2P_DISABLE_DISCOVERY", "0"))
      ethereum_key       = "default"
    }
  }

  # WebAPI transport configuration. Enabled if CFG_WEBAPI_LISTEN_ADDR is set to a listen address.
  dynamic "webapi" {
    for_each = var.webapi_enable ? [1] : []
    content {
      feeds             = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(",", env("CFG_FEEDS", "")))
      listen_addr       = var.webapi_listen_addr
      socks5_proxy_addr = var.webapi_socks5_proxy_addr # will not try to connect to a proxy if empty
      ethereum_key      = "default"

      # Ethereum based address book. Enabled if CFG_WEBAPI_ETH_ADDR_BOOK is set to a contract address.
      dynamic "ethereum_address_book" {
        for_each = var.webapi_eth_addr_book == "" ? [] : [1]
        content {
          contract_addr   = var.webapi_eth_addr_book
          ethereum_client = "ethereum"
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
