package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// MultiChainConfigFile represents the structure of the multi-chain configuration file
type MultiChainConfigFile struct {
	HTTPPort        int                 `json:"http_port"`
	ProxyCount      int                 `json:"proxy_count"`
	HcaptchaSiteKey string              `json:"hcaptcha_sitekey"`
	HcaptchaSecret  string              `json:"hcaptcha_secret"`
	DefaultNetwork  string              `json:"default_network"`
	Networks        []NetworkConfigFile `json:"networks"`
}

// NetworkConfigFile defines a single network entry inside multichain-config.json
type NetworkConfigFile struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name,omitempty"`
	ChainID     int64   `json:"chain_id"`
	Symbol      string  `json:"symbol"`
	IsTestnet   bool    `json:"is_testnet"`
	ExplorerURL string  `json:"explorer_url,omitempty"`
	Provider    string  `json:"provider"`
	PrivateKey  string  `json:"private_key"`
	Keystore    string  `json:"keystore"`
	KeyPass     string  `json:"key_pass"`
	Payout      float64 `json:"payout"`
	Interval    int     `json:"interval"`
}

// LoadMultiChainConfigFile reads and parses multichain-config.json
func LoadMultiChainConfigFile(path string) (*MultiChainConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg MultiChainConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}
	return &cfg, nil
}

// BuildNetworkConfigFromFile converts a file entry into a runtime NetworkConfig
func BuildNetworkConfigFromFile(file NetworkConfigFile) (NetworkConfig, error) {
	name := strings.TrimSpace(file.DisplayName)
	if name == "" {
		name = strings.TrimSpace(file.Name)
	}
	if name == "" {
		return NetworkConfig{}, fmt.Errorf("network name is required")
	}

	symbol := file.Symbol
	if symbol == "" {
		symbol = "ETH"
	}

	return NetworkConfig{
		ChainID:     file.ChainID,
		Symbol:      symbol,
		Name:        name,
		IsTestnet:   file.IsTestnet,
		DefaultRPC:  file.Provider,
		ExplorerURL: file.ExplorerURL,
	}, nil
}

// FindNetworkInFile searches a parsed multichain config for a specific network
func FindNetworkInFile(cfg *MultiChainConfigFile, network string) *NetworkConfigFile {
	for _, n := range cfg.Networks {
		if strings.EqualFold(n.Name, network) {
			return &n
		}
	}
	return nil
}
