package cmd

import (
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/guyuxiang/multi-chain-faucet/internal/chain"
	"github.com/guyuxiang/multi-chain-faucet/internal/config"
	"github.com/guyuxiang/multi-chain-faucet/internal/server"
)

var (
	appVersion = "v1.2.0"

	httpPortFlag       = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag       = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	versionFlag        = flag.Bool("version", false, "Print version number")
	listNetworksFlag   = flag.Bool("list-networks", false, "List all supported networks and exit")
	multiChainFlag     = flag.String("multichain", "", "Path to multi-chain configuration file")
	generateConfigFlag = flag.Bool("generate-config", false, "Generate sample multi-chain configuration file")

	payoutFlag   = flag.Float64("faucet.amount", 1, "Number of Ethers to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 1440, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "testnet", "Network name to display on the frontend")
	symbolFlag   = flag.String("faucet.symbol", "ETH", "Token symbol to display on the frontend")

	keyJSONFlag  = flag.String("wallet.keyjson", os.Getenv("KEYSTORE"), "Keystore file to fund user requests with")
	keyPassFlag  = flag.String("wallet.keypass", "password.txt", "Passphrase text file to decrypt keystore")
	privKeyFlag  = flag.String("wallet.privkey", os.Getenv("PRIVATE_KEY"), "Private key hex to fund user requests with")
	providerFlag = flag.String("wallet.provider", os.Getenv("WEB3_PROVIDER"), "Endpoint for Ethereum JSON-RPC connection")

	hcaptchaSiteKeyFlag = flag.String("hcaptcha.sitekey", os.Getenv("HCAPTCHA_SITEKEY"), "hCaptcha sitekey")
	hcaptchaSecretFlag  = flag.String("hcaptcha.secret", os.Getenv("HCAPTCHA_SECRET"), "hCaptcha secret")
)

const defaultMultiChainConfigPath = "multichain-config.json"

// ListSupportedNetworks prints all supported networks (useful for CLI help)
func ListSupportedNetworks(configPath string) {
	cfg, err := config.LoadMultiChainConfigFile(configPath)
	if err != nil {
		fmt.Printf("failed to load %s: %v\n", configPath, err)
		return
	}

	fmt.Println("Supported networks:")
	fmt.Println("==================")

	for _, network := range cfg.Networks {
		testnetStr := ""
		if network.IsTestnet {
			testnetStr = " (Testnet)"
		}
		displayName := network.DisplayName
		if displayName == "" {
			displayName = network.Name
		}
		fmt.Printf("- %-20s %s%s (Chain ID: %d, Symbol: %s)\n",
			network.Name, displayName, testnetStr, network.ChainID, network.Symbol)
	}
}

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	if *listNetworksFlag {
		configPath := *multiChainFlag
		if configPath == "" {
			configPath = defaultMultiChainConfigPath
		}
		ListSupportedNetworks(configPath)
		os.Exit(0)
	}
	if *generateConfigFlag {
		if err := GenerateMultiChainConfig("multichain-config.json"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if *multiChainFlag != "" {
		ExecuteMultiChain(*multiChainFlag)
		return
	}
}

func Execute() {
	networkName := strings.ToLower(*netnameFlag)
	fileNetwork, err := loadNetworkFromConfig(networkName)
	if err != nil {
		panic(fmt.Errorf("failed to read network config: %w", err))
	}

	privateKey, err := getPrivateKeyFromFlags()
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}

	var chainID *big.Int
	var chainIDValue int64
	var symbol string
	var displayName string
	var explorerURL string

	if fileNetwork != nil {
		if fileNetwork.ChainID != 0 {
			chainID = big.NewInt(fileNetwork.ChainID)
			chainIDValue = fileNetwork.ChainID
		}
		explorerURL = fileNetwork.ExplorerURL
		displayName = fileNetwork.Name
		if *symbolFlag == "ETH" {
			symbol = fileNetwork.Symbol
		} else {
			symbol = *symbolFlag
		}
		if *providerFlag == "" {
			*providerFlag = fileNetwork.DefaultRPC
		}
	} else {
		symbol = *symbolFlag
		displayName = *netnameFlag
	}

	// Validate provider is set
	if *providerFlag == "" {
		panic(fmt.Errorf("web3 provider is required. Set via -wallet.provider flag or WEB3_PROVIDER environment variable"))
	}

	txBuilder, err := chain.NewTxBuilder(*providerFlag, privateKey, chainID)
	if err != nil {
		panic(fmt.Errorf("cannot connect to web3 provider: %w", err))
	}

	config := server.NewConfig(displayName, symbol, chainIDValue, explorerURL, *httpPortFlag, *intervalFlag, *proxyCntFlag, *payoutFlag, *hcaptchaSiteKeyFlag, *hcaptchaSecretFlag)
	go server.NewServer(txBuilder, config).Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func loadNetworkFromConfig(networkName string) (*config.NetworkConfig, error) {
	cfg, err := config.LoadMultiChainConfigFile(defaultMultiChainConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	entry := config.FindNetworkInFile(cfg, networkName)
	if entry == nil {
		return nil, nil
	}

	networkCfg, err := config.BuildNetworkConfigFromFile(*entry)
	if err != nil {
		return nil, err
	}

	return &networkCfg, nil
}

func getPrivateKeyFromFlags() (*ecdsa.PrivateKey, error) {
	if *privKeyFlag != "" {
		hexkey := *privKeyFlag
		if chain.Has0xPrefix(hexkey) {
			hexkey = hexkey[2:]
		}
		return crypto.HexToECDSA(hexkey)
	} else if *keyJSONFlag == "" {
		return nil, errors.New("missing private key or keystore")
	}

	keyfile, err := chain.ResolveKeyfilePath(*keyJSONFlag)
	if err != nil {
		return nil, err
	}
	password, err := os.ReadFile(*keyPassFlag)
	if err != nil {
		return nil, err
	}

	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}
