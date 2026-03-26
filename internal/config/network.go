package config

// NetworkConfig holds runtime metadata for a blockchain network.
type NetworkConfig struct {
	ChainID     int64
	Symbol      string
	Name        string
	IsTestnet   bool
	DefaultRPC  string
	ExplorerURL string
}
