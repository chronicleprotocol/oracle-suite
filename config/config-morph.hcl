morph {
  cache_path = env("CFG_CONFIG_CACHE", "config-cache.hcl")
  ethereum_client = "default"
  config_registry = var.contract_map["${var.environment}-${var.chain_name}-ConfigRegistry"]
  interval = tonumber(env("CFG_MORPH_INTERVAL", "3600"))

  app {
    work_dir = env("CFG_APP_WORK_DIR", "")
    bin = env("CFG_APP_BIN", "")
    use = env("CFG_APP_USE", "run")
    args = env("CFG_APP_ARGS", "")
    waiting_running = tonumber(env("CFG_APP_RUN_DURATION", "60"))
    waiting_quiting = tonumber(env("CFG_APP_QUIT_DURATION", "60"))
  }
}