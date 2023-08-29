variables {
  spire_keys = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_SPIRE_KEYS", ""))
}

spire {
  # Ethereum key to use for signing messages. The key must be present in the `ethereum` section.
  # (optional) if not set, the first key in the `ethereum` section is used.
  ethereum_key = "default"

  rpc_listen_addr = env("CFG_SPIRE_RPC_ADDR", ":9100")
  rpc_agent_addr  = env("CFG_SPIRE_RPC_ADDR", "127.0.0.1:9100")

  # List of pairs that are collected by the spire node. Other pairs are ignored.
  pairs = length(var.spire_keys) == 0 ? [
    for s in var.data_symbols : replace(s, "/", "")
  ] : var.spire_keys

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_FEEDS", "")))
}
