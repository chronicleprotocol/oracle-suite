variables {
  spire_keys = explode(var.item_separator, env("CFG_MODELS", env("CFG_SPIRE_KEYS", "")))
}

spire {
  # Ethereum key to use for signing messages. The key must be present in the `ethereum` section.
  # (optional) if not set, the first key in the `ethereum` section is used.
  ethereum_key = "default"

  rpc_listen_addr = env("CFG_SPIRE_RPC_ADDR", ":9100")
  rpc_agent_addr  = env("CFG_SPIRE_RPC_ADDR", "127.0.0.1:9100")

  # List of pairs that are collected by the spire node. Other pairs are ignored.
  pairs = distinct(concat([
    for v in var.contracts : v.wat
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_scribe, false)
    # If CFG_SPIRE_KEYS is set to a list of asset symbols
    && try(length(var.spire_keys) == 0 || contains(var.spire_keys, v.wat), false)
  ], [
    for v in var.contracts : replace(v.wat, "/", "")
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Median compatible contracts
    && try(v.is_median, false)
    # If CFG_SPIRE_KEYS is set to a list of asset symbols
    && try(length(var.spire_keys) == 0 || contains(var.spire_keys, v.wat), false)
  ]))

  # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
  feeds = var.feeds
}
