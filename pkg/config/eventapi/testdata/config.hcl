listen_addr = "0.0.0.0:8000"

# Storage memory
storage_memory {
  ttl = 86400
}

# Storage Redis
storage_redis {
  ttl              = 86400
  addr             = "localhost:6379"
  user             = "user"
  pass             = "password"
  db               = 0
  memory_limit     = 1048576
  tls              = false
  tls_server_name  = "localhost"
  tls_cert_file    = "./tls_cert.pem"
  tls_key_file     = "./tls_key.pem"
  tls_root_ca_file = "./tls_root_ca.pem"
  cluster          = false
  cluster_addrs    = ["localhost:7000", "localhost:7001"]
}
