ghost {
  ethereum_key = "default"

  interval = try(tonumber(env.CFG_GHOST_INTERVAL, 60))
  pairs    = try(env.CFG_GHOST_PAIRS == "" ? [] : split(",", env.CFG_GHOST_PAIRS), [
    "BTC/USD",
    "ETH/BTC",
    "ETH/USD",
    "GNO/USD",
    "IBTA/USD",
    "LINK/USD",
    "MANA/USD",
    "MATIC/USD",
    "MKR/USD",
    "RETH/USD",
    "WSTETH/USD",
    "YFI/USD",
  ])
}
