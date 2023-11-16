morph {
  cache_path = env("CFG_CONFIG_CACHE", "config-cache.hcl")
  ethereum_client = "default"
  config_registry = var.contract_map["${var.environment}-${var.chain_name}-ConfigRegistry"]
  interval = tonumber(env("CFG_MORPH_INTERVAL", "3600"))
  work_dir = env("CFG_WORK_DIR", "")
  executable_binary = env("CFG_EXECUTEABLE_BINARY", "")
  waiting_app_running = tonumber(env("CFG_RUN_APP_DURATION", "60"))
  waiting_app_quiting = tonumber(env("CFG_QUIT_APP_DURATION", "60"))
}