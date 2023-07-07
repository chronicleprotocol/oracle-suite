spire {
  # Ethereum key to use for signing messages. The key must be present in the `ethereum` section.
  # (optional) if not set, the first key in the `ethereum` section is used.
  ethereum_key = "default"

  rpc_listen_addr = try(env.CFG_SPIRE_RPC_ADDR, "0.0.0.0:9100")
  rpc_agent_addr  = try(env.CFG_SPIRE_RPC_ADDR, "127.0.0.1:9100")

  # List of pairs that are collected by the spire node. Other pairs are ignored.
  pairs = try(env.CFG_SPIRE_PAIRS == "" ? [] : split(",", env.CFG_SPIRE_PAIRS), [
    "BTCUSD",
    "ETHBTC",
    "ETHUSD",
    "GNOUSD",
    "IBTAUSD",
    "LINKUSD",
    "MANAUSD",
    "MATICUSD",
    "MKRUSD",
    "RETHUSD",
    "WSTETHUSD",
    "YFIUSD",
  ])

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = try(env.CFG_FEEDS, "")=="*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[try(env.CFG_FEEDS, "prod")], split(",", try(env.CFG_FEEDS, "")))
}
