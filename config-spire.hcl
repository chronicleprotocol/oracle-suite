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
  pairs = concat(length(var.spire_keys) == 0 ? keys(var.median_contracts[var.chain_name]) : var.spire_keys, [
    for p in (length(var.spire_keys) == 0 ? keys(var.median_contracts[var.chain_name]) : var.spire_keys) :
    replace(p, "/", "")
  ])

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = try(var.feed_sets[env("CFG_FEEDS", var.environment)], explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_FEEDS", "")))
}
