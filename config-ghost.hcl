ghost {
  ethereum_key = "default"
  interval     = try(tonumber(env.CFG_GHOST_INTERVAL), 60)
  data_models  = try(env.CFG_GHOST_PAIRS == "" ? [] : split(",", env.CFG_GHOST_PAIRS), var.data_symbols)
}
