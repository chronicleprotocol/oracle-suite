ghostnext {
  ethereum_key = "default"
  interval     = try(tonumber(env.CFG_GHOST_INTERVAL), 60)
  data_models  = [
    "BTC/USD"
  ]
}
