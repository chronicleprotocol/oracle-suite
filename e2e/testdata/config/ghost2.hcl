variables {
  feeds = [
    "0x2D800d93B065CE011Af83f316ceF9F0d005B0AA4",
    "0xe3ced0f62f7eb2856d37bed128d2b195712d2644"
  ]
}

ethereum {
  client "default" {
    rpc_urls     = ["https://eth.public-rpc.com"]
    chain_id     = 1
    ethereum_key = "default"
  }

  key "default" {
    address         = "2d800d93b065ce011af83f316cef9f0d005b0aa4"
    keystore_path   = "./e2e/testdata/keys"
    passphrase_file = "./e2e/testdata/keys/pass"
  }
}

transport {
  libp2p {
    feeds           = var.feeds
    priv_key_seed   = "2d800d93b065ce011af83f316cef9f0d005b0aa42d800d93b065ce011af83f31"
    listen_addrs    = ["/ip4/0.0.0.0/tcp/30101"]
    bootstrap_addrs = ["/ip4/127.0.0.1/tcp/30100/p2p/12D3KooWSGCRPjd6dHHjfWYeKnurLcaSYAQsQqDYj7GcPN2uhdis"]
    ethereum_key    = "default"
  }
}

gofer {
  origin "kraken" {
    type = "tick_generic_jq"
    url  = "http://127.0.0.1:8080/0/public/Ticker?pair=$${ucbase}/$${ucquote}"
    jq   = "($ucbase + \"/\" + $ucquote) as $pair | {price: .result[$pair].c[0]|tonumber, time: now|round, volume: .result[$pair].v[0]|tonumber}"
  }

  data_model "BTC/USD" {
    median {
      min_values = 1
      origin "kraken" { query = "BTC/USD" }
    }
  }

  data_model "ETH/BTC" {
    median {
      min_values = 1
      origin "kraken" { query = "ETH/BTC" }
    }
  }

  data_model "ETH/USD" {
    median {
      min_values = 1
      origin "kraken" { query = "ETH/USD" }
    }
  }

  data_model "BTC/ETH" {
    invert {
      origin "kraken" { query = "ETH/BTC" }
    }
  }
}

ghost {
  ethereum_key = "default"
  interval     = 1
  data_models  = [
    "BTC/USD",
    "ETH/BTC",
    "BTC/ETH",
  ]
}
