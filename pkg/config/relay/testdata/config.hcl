median {
  ethereum_client = "client1"
  contract_addr   = "0x1234567890123456789012345678901234567890"
  data_model      = "ETH/USD"
  spread          = 1
  expiration      = 300
  feeds           = [
    "0x0011223344556677889900112233445566778899",
    "0x1122334455667788990011223344556677889900",
  ]
}

scribe {
  ethereum_client = "client2"
  contract_addr   = "0x2345678901234567890123456789012345678901"
  data_model      = "BTC/USD"
  spread          = 2
  expiration      = 400
  feeds           = [
    "0x2233445566778899001122334455667788990011",
    "0x3344556677889900112233445566778899001122",
  ]
}

optimistic_scribe {
  ethereum_client       = "client3"
  contract_addr         = "0x3456789012345678901234567890123456789012"
  data_model            = "MKR/USD"
  spread                = 3
  expiration            = 500
  optimistic_spread     = 4
  optimistic_expiration = 600
  feeds                 = [
    "0x4455667788990011223344556677889900112233",
    "0x5566778899001122334455667788990011223344",
  ]
}
