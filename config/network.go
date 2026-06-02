package config

import (
	"fmt"
	"math/big"
)

type NetworkType string

const (
	NetworkSepolia NetworkType = "sepolia"
	NetworkLocal   NetworkType = "local"
)

type NetworkConfig struct {
	Name    NetworkType
	RPCURL  string
	WSURL   string
	ChainID *big.Int
}

var (
	networkConfigs = map[NetworkType]NetworkConfig{
		NetworkSepolia: {
			Name:    NetworkSepolia,
			RPCURL:  "https://sepolia.infura.io/v3/YOUR_INFURA_KEY",
			WSURL:   "wss://sepolia.infura.io/ws/v3/YOUR_INFURA_KEY",
			ChainID: big.NewInt(11155111),
		},
		NetworkLocal: {
			Name:    NetworkLocal,
			RPCURL:  "http://127.0.0.1:8545",
			WSURL:   "ws://127.0.0.1:8545",
			ChainID: big.NewInt(31337),
		},
	}
)

func GetNetworkConfig(network NetworkType) (NetworkConfig, error) {
	cfg, exists := networkConfigs[network]
	if !exists {
		return NetworkConfig{}, fmt.Errorf("unsupported network: %s", network)
	}
	return cfg, nil
}

func GetDefaultNetwork() NetworkType {
	return NetworkLocal
}
