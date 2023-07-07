leeloo {
  ethereum_key = "default"

  # Arbitrum
  # Enabled if CFG_TELEPORT_EVM_ARB_CONTRACT_ADDRS is set.
  dynamic "teleport_evm" {
    for_each = try(env.CFG_TELEPORT_EVM_ARB_CONTRACT_ADDRS, "") == "" ? [] : [1]
    content {
      ethereum_client     = "arbitrum"
      interval            = tonumber(try(env.CFG_TELEPORT_EVM_ARB_INTERVAL, 60))
      prefetch_period     = tonumber(try(env.CFG_TELEPORT_EVM_ARB_PREFETCH_PERIOD, 3600 * 24 * 7))
      block_confirmations = tonumber(try(env.CFG_TELEPORT_EVM_ARB_BLOCK_CONFIRMATIONS, 0))
      block_limit         = tonumber(try(env.CFG_TELEPORT_EVM_ARB_BLOCK_LIMIT, 1000))
      replay_after        = concat(
        [60, 300, 3600, 3600*2, 3600*4],
        [for i in range(3600 * 6, 3600 * 24 * 7, 3600 * 6) :i]
      )
      contract_addrs = try(split(",", env.CFG_TELEPORT_EVM_ARB_CONTRACT_ADDRS), [])
    }
  }

  # Optimism
  # Enabled if CFG_TELEPORT_EVM_OPT_CONTRACT_ADDRS is set.
  dynamic "teleport_evm" {
    for_each = try(env.CFG_TELEPORT_EVM_OPT_CONTRACT_ADDRS, "") == "" ? [] : [1]
    content {
      ethereum_client     = "optimism"
      interval            = tonumber(try(env.CFG_TELEPORT_EVM_OPT_INTERVAL, 60))
      prefetch_period     = tonumber(try(env.CFG_TELEPORT_EVM_OPT_PREFETCH_PERIOD, 3600 * 24 * 7))
      block_confirmations = tonumber(try(env.CFG_TELEPORT_EVM_OPT_BLOCK_CONFIRMATIONS, 0))
      block_limit         = tonumber(try(env.CFG_TELEPORT_EVM_OPT_BLOCK_LIMIT, 1000))
      replay_after        = concat(
        [60, 300, 3600, 3600*2, 3600*4],
        [for i in range(3600 * 6, 3600 * 24 * 7, 3600 * 6) :i]
      )
      contract_addrs = try(split(",", env.CFG_TELEPORT_EVM_OPT_CONTRACT_ADDRS), [])
    }
  }

  # Starknet
  # Enabled if CFG_TELEPORT_STARKNET_CONTRACT_ADDRS is set.
  dynamic "teleport_starknet" {
    for_each = try(env.CFG_TELEPORT_STARKNET_CONTRACT_ADDRS, "") == "" ? [] : [1]
    content {
      sequencer       = try(env.CFG_TELEPORT_STARKNET_SEQUENCER, "https://alpha-mainnet.starknet.io")
      interval        = tonumber(try(env.CFG_TELEPORT_STARKNET_INTERVAL, 60))
      prefetch_period = tonumber(try(env.CFG_TELEPORT_STARKNET_PREFETCH_PERIOD, 3600 * 24 * 7))
      replay_after    = concat(
        [60, 300, 3600, 3600*2, 3600*4],
        [for i in range(3600 * 6, 3600 * 24 * 7, 3600 * 6) :i]
      )
      contract_addrs = try(split(",", env.CFG_TELEPORT_STARKNET_CONTRACT_ADDRS), [])
    }
  }
}
