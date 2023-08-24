variables {
  spectre_target_network = env("CFG_SPECTRE_TARGET_NETWORK", "")
  spectre_pairs          = explode(",", env("CFG_SPECTRE_PAIRS", ""))
}

spectre {
  dynamic "median" {
    for_each = {
      for p in length(var.spectre_pairs) == 0 ? var.data_symbols : var.spectre_pairs : p=>
      var.median_contracts[var.spectre_target_network][p]
      if contains(try(keys(var.median_contracts[var.spectre_target_network]), []), p)
    }
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.oracle

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(",", env("CFG_FEEDS", "")))

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
    for_each = {
      for p in length(var.spectre_pairs) == 0 ? var.data_symbols : var.spectre_pairs : p=>
      var.scribe_contracts[var.spectre_target_network][p]
      if contains(try(keys(var.scribe_contracts[var.spectre_target_network]), []), p)
    }
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(",", env("CFG_FEEDS", "")))

      # Name of the pair to fetch the price for.
      data_model = contract.key

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.fallbackSpread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.fallbackExpiration

      # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
      interval = contract.value.fallbackInterval
    }
  }

  dynamic "optimistic_scribe" {
    for_each = {
      for p in length(var.spectre_pairs) == 0 ? var.data_symbols : var.spectre_pairs : p=>
      var.scribe_contracts[var.spectre_target_network][p]
      if contains(try(keys(var.scribe_contracts[var.spectre_target_network]), []), p)
    }
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = "default"

      # Address of the Median contract.
      contract_addr = contract.value.address

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = env("CFG_FEEDS", "") == "*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[env("CFG_FEEDS", "prod")], explode(",", env("CFG_FEEDS", "")))

      # Name of the pair to fetch the price for.
      data_model = contract.key

      # Spread in percent points above which the price is considered stale.
      spread = contract.value.opSpread

      # Time in seconds after which the price is considered stale.
      expiration = contract.value.opExpiration

      # Specifies how often in seconds Spectre should check if Oracle contract needs to be updated.
      interval = contract.value.opInterval
    }
  }
}
