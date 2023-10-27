variables {
  spectre_pairs = explode(var.item_separator, env("CFG_SYMBOLS", env("CFG_SPECTRE_PAIRS", "")))
}

spectre {
  dynamic "median" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.IMedian, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = replace(contract.value.wat, "/", "")

      # Spread in percent points above which the price is considered stale.
      spread = try(contract.value.poke.spread, 1)

      # Time in seconds after which the price is considered stale.
      expiration = try(contract.value.poke.expiration, 32400)
    }
  }

  dynamic "scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.IScribe, false) && try(v.IScribeOptimistic, false) == false
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = try(contract.value.poke.spread, 1)

      # Time in seconds after which the price is considered stale.
      expiration = try(contract.value.poke.expiration, 32400)
    }
  }

  dynamic "optimistic_scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.IScribe, false) && try(v.IScribeOptimistic, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = var.feeds

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = try(contract.value.poke.spread, 1)

      # Time in seconds after which the price is considered stale.
      expiration = try(contract.value.poke.expiration, 32400)

      # Spread in percent points above which the price is considered stale.
      optimistic_spread = try(contract.value.optimistic_poke.spread, 0.5)

      # Time in seconds after which the price is considered stale.
      optimistic_expiration = try(contract.value.optimistic_poke.expiration, 28800)
    }
  }
}
