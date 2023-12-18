package kafka

type ConfigKafka struct {
	// RPCListenAddr is an address to listen for RPC requests.
	Brokers string `hcl:"brokers"`
}
