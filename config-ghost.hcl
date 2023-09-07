variables {
  ghost_pairs = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_GHOST_PAIRS", ""))
}

ghost {
  ethereum_key = "default"
  interval     = tonumber(env("CFG_GHOST_INTERVAL", "60"))
  data_models  = concat(length(var.ghost_pairs) == 0 ? keys(var.median_contracts[var.chain_name]) : var.ghost_pairs, [
    for pair in length(var.ghost_pairs) == 0 ? keys(var.median_contracts[var.chain_name]) : var.ghost_pairs :
    replace(pair, "/", "")
  ])
}
