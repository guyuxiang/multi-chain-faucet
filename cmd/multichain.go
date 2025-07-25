package cmd

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/guyuxiang/multi-chain-faucet/internal/chain"
	"github.com/guyuxiang/multi-chain-faucet/internal/config"
	"github.com/guyuxiang/multi-chain-faucet/internal/server"
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

type NetworkConfigFile struct {
	Name       string  `json:"name"`
	Provider   string  `json:"provider"`
	PrivateKey string  `json:"private_key"`
	Keystore   string  `json:"keystore"`
	KeyPass    string  `json:"key_pass"`
	Payout     float64 `json:"payout"`
	Interval   int     `json:"interval"`
}

// ExecuteMultiChain starts the multi-chain faucet server
func ExecuteMultiChain(configPath string) {
	// Load configuration
	multiConfig, err := loadMultiChainConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to load multi-chain config: %w", err))
	}

	// Create and start server
	server, err := server.NewMultiChainServer(multiConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create multi-chain server: %w", err))
	}

	go server.Run()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// loadMultiChainConfig loads configuration from JSON file
func loadMultiChainConfig(configPath string) (*config.MultiChainConfig, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var fileConfig MultiChainConfigFile
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Create multi-chain config
	multiConfig := config.NewMultiChainConfig()
	multiConfig.HTTPPort = fileConfig.HTTPPort
	multiConfig.ProxyCount = fileConfig.ProxyCount
	multiConfig.HcaptchaSiteKey = fileConfig.HcaptchaSiteKey
	multiConfig.HcaptchaSecret = fileConfig.HcaptchaSecret

	// Add networks
	for _, netConfig := range fileConfig.Networks {
		chainInput := config.ChainConfigInput{
			Network:  netConfig.Name,
			Provider: netConfig.Provider,
			Payout:   netConfig.Payout,
			Interval: netConfig.Interval,
		}

		// Parse private key or keystore
		var privateKey *ecdsa.PrivateKey
		var err error

		if netConfig.PrivateKey != "" {
			privateKey, err = parsePrivateKeyHex(netConfig.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key for %s: %w", netConfig.Name, err)
			}
		} else if netConfig.Keystore != "" {
			privateKey, err = parseKeystoreFile(netConfig.Keystore, netConfig.KeyPass)
			if err != nil {
				return nil, fmt.Errorf("failed to parse keystore for %s: %w", netConfig.Name, err)
			}
		} else {
			return nil, fmt.Errorf("network %s requires either private_key or keystore", netConfig.Name)
		}

		if err := multiConfig.AddChainWithKey(chainInput, privateKey); err != nil {
			return nil, fmt.Errorf("failed to add network %s: %w", netConfig.Name, err)
		}
	}

	// Set default network
	if fileConfig.DefaultNetwork != "" {
		if _, exists := multiConfig.GetChain(fileConfig.DefaultNetwork); !exists {
			return nil, fmt.Errorf("default network %s is not configured", fileConfig.DefaultNetwork)
		}
		multiConfig.DefaultChain = fileConfig.DefaultNetwork
	}

	return multiConfig, nil
}

// GenerateMultiChainConfig creates a sample configuration file
func GenerateMultiChainConfig(outputPath string) error {
	sampleConfig := MultiChainConfigFile{
		HTTPPort:        8080,
		ProxyCount:      0,
		HcaptchaSiteKey: "",
		HcaptchaSecret:  "",
		DefaultNetwork:  "sepolia",
		Networks: []NetworkConfigFile{
			{
				Name:       "sepolia",
				Provider:   "",                      // Will use default
				PrivateKey: "0x1234567890abcdef...", // Replace with actual key
				Payout:     1.0,
				Interval:   1440,
			},
			{
				Name:       "polygon-mumbai",
				Provider:   "",                      // Will use default
				PrivateKey: "0x1234567890abcdef...", // Replace with actual key
				Payout:     1.0,
				Interval:   1440,
			},
			{
				Name:       "bsc-testnet",
				Provider:   "",                      // Will use default
				PrivateKey: "0x1234567890abcdef...", // Replace with actual key
				Payout:     0.1,
				Interval:   1440,
			},
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(sampleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Sample multi-chain configuration written to %s\n", outputPath)
	fmt.Println("\nPlease update the private keys and other settings before using.")
	return nil
}

// Helper function to parse private key (moved here from config package)
func parsePrivateKeyHex(hexkey string) (*ecdsa.PrivateKey, error) {
	if strings.HasPrefix(hexkey, "0x") {
		hexkey = hexkey[2:]
	}
	return crypto.HexToECDSA(hexkey)
}

// Helper function to parse keystore
func parseKeystoreFile(keystorePath, passwordPath string) (*ecdsa.PrivateKey, error) {
	// Resolve keystore path
	keyfile, err := chain.ResolveKeyfilePath(keystorePath)
	if err != nil {
		return nil, err
	}

	// Read password
	password, err := os.ReadFile(passwordPath)
	if err != nil {
		return nil, err
	}

	// Decrypt keystore
	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}
