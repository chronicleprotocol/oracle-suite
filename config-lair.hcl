lair {
  listen_addr = try(env.CFG_LAIR_LISTEN_ADDR, "0.0.0.0:8082")

  # Configuration for memory storage. Enabled if CFG_LAIR_STORAGE is "memory" or unset.
  dynamic "storage_memory" {
    for_each = try(env.CFG_LAIR_STORAGE, "memory") == "memory" ? [1] : []
    content {}
  }

  # Configuration for redis storage. Enabled if CFG_LAIR_STORAGE is "redis".
  dynamic "storage_redis" {
    for_each = try(env.CFG_LAIR_STORAGE, "") == "redis" ? [1] : []
    content {
      addr                     = try(env.CFG_LAIR_REDIS_ADDR, "127.0.0.1:6379")
      user                     = try(env.CFG_LAIR_REDIS_USER, "")
      pass                     = try(env.CFG_LAIR_REDIS_PASS, "")
      db                       = tonumber(try(env.CFG_LAIR_REDIS_DB, 0))
      memory_limit             = tonumber(try(env.CFG_LAIR_REDIS_MEMORY_LIMIT, 0))
      tls                      = tobool(try(env.CFG_LAIR_REDIS_TLS, false))
      tls_server_name          = try(env.CFG_LAIR_REDIS_TLS_SERVER_NAME, "")
      tls_cert_file            = try(env.CFG_LAIR_REDIS_TLS_CERT_FILE, "")
      tls_key_file             = try(env.CFG_LAIR_REDIS_TLS_KEY_FILE, "")
      tls_root_ca_file         = try(env.CFG_LAIR_REDIS_TLS_ROOT_CA_FILE, "")
      tls_insecure_skip_verify = tobool(try(env.CFG_LAIR_REDIS_TLS_INSECURE, false))
      cluster                  = tobool(try(env.CFG_LAIR_REDIS_CLUSTER, false))
      cluster_addrs            = try(env.CFG_LAIR_REDIS_CLUSTER_ADDRS == "" ? [] : split(",", env.CFG_LAIR_REDIS_CLUSTER_ADDRS), [])
    }
  }
}
