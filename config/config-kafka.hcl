kafka {
  # Comma separated list of kafka brokers
  brokers  = env("CFG_KAFKA_BROKERS", "127.0.0.1:9092")
}
