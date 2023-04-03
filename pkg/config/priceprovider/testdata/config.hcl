rpc_listen_addr = "localhost:8080"
rpc_agent_addr  = "localhost:8081"

origin "origin" {
  type   = "origin"
  params = {
    contract_address = "0x1234567890123456789012345678901234567890"
  }
}

price_model "AAA/BBB" "median" {
  source "AAA/BBB" "origin" { origin = "origin1" }
  source "AAA/BBB" "indirect" {
    source "AAA/XXX" "origin" { origin = "origin2" }
    source "XXX/BBB" "origin" { origin = "origin3" }
  }
  min_sources = 3
}

