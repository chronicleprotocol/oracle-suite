grafana {
  interval = 60
  endpoint = "https://graphite.example.com"
  api_key  = "your_api_key"

  metric {
    match_message = "message"
    match_fields  = {
      type = "sell"
    }
    value        = "message.path"
    scale_factor = 0.5
    name         = "example.message"
    tags         = {
      environment = ["production"]
    }
    on_duplicate = "sum"
  }
}
