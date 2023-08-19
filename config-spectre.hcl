variables {
  spectre_target_network  = try(env.CFG_SPECTRE_TARGET_NETWORK, "ethereum-mainnet")
  spectre_ethereum_client = split("-", try(env.CFG_SPECTRE_TARGET_NETWORK, "ethereum-mainnet"))[0]
}

spectre {
  dynamic "median" {
    for_each = var.median_contracts[var.spectre_target_network]
    iterator = contract
    content {
      # Ethereum client to use for interacting with the Median contract.
      ethereum_client = var.spectre_ethereum_client

      # Address of the Median contract.
      contract_addr = contract.value.oracle

      # List of feeds that are allowed to be storing messages in storage. Other feeds are ignored.
      feeds = try(env.CFG_FEEDS, "")=="*" ? concat(var.feed_sets["prod"], var.feed_sets["stage"]) : try(var.feed_sets[try(env.CFG_FEEDS, "prod")], split(",", try(env.CFG_FEEDS, "")))

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

}
