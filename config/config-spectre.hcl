variables {
  spectre_pairs = explode(var.item_separator, env("CFG_MODELS", env("CFG_SPECTRE_PAIRS", "")))
}

spectre {
  dynamic "median" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.is_median, false)
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
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration
    }
  }

  dynamic "scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(split(".", v.version)[0] == "v2", false)
      && try(v.is_scribe, false) && try(!v.is_scribe_optimistic, false)
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
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration
    }
  }

  dynamic "optimistic_scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(split(".", v.version)[0] == "v2", false)
      && try(v.is_scribe, false) && try(v.is_scribe_optimistic, false)
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
      spread = contract.value.poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.poke.expiration

      # Spread in percent points above which the price is considered stale.
      optimistic_spread = contract.value.poke_optimistic.spread

      # Time in seconds after which the price is considered stale.
      optimistic_expiration = contract.value.poke_optimistic.expiration
    }
  }
}
