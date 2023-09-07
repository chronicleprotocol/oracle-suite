variables {
  spectre_pairs = explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_SPECTRE_PAIRS", ""))
}

spectre {
  dynamic "median" {
    for_each = {
      for p in length(var.spectre_pairs) == 0 ? var.median_contracts[var.chain_name] : var.spectre_pairs : p=>
      var.median_contracts[var.chain_name][p]
      if contains(try(keys(var.median_contracts[var.chain_name]), []), p)
    }
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.oracle

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = try(var.feed_sets[env("CFG_FEEDS", var.environment)], explode(env("CFG_ITEM_SEPARATOR", ","), env("CFG_FEEDS", "")))

      # Name of the pair to fetch the price for.
      data_model = replace(contract.key, "/", "")

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.oracleSpread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.oracleExpiration

      # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
      interval = 60
    }
  }

  dynamic "scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.IScribe, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].poke.expiration

      # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
      interval = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].poke.interval
    }
  }

  dynamic "optimistic_scribe" {
    for_each = [
      for v in var.contracts : v
      if v.env == var.environment
      && v.chain == var.chain_name
      && try(v.IScribe, false)
      && try(length(var.spectre_pairs) == 0 || contains(var.spectre_pairs, v.wat), false)
      && try(v.IScribeOptimistic, false)
    ]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # Name of the pair to fetch the price for.
      data_model = contract.value.wat

      # Spread in percent points above which the price is considered stale.
      spread = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].optimistic_poke.spread

      # Time in seconds after which the price is considered stale.
      expiration = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].optimistic_poke.expiration

      # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
      interval = var.contract_params["${contract.value.env}-${contract.value.chain}-${contract.value.address}"].optimistic_poke.interval
    }
  }
}
