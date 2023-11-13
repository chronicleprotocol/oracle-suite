morph {
  cache_path = env("CFG_CONFIG_CACHE", "config-cache.hcl")
  ethereum_client = "default"
  config_registry = var.contract_map["${var.environment}-${var.chain_name}-ConfigRegistry"]
  interval = tonumber(env("CFG_MORPH_INTERVAL", "3600"))
}