package config

import (
	"os"
)

type Config struct {
	EthRPCURL     string
	EthWSURL      string
	ERC20Contract string
}

func Load() *Config {
	wsURL := os.Getenv("ETH_WS_URL")
	rpcURL := os.Getenv("ETH_RPC_URL")
	contractAddr := os.Getenv("ERC20_CONTRACT")

	return &Config{
		EthRPCURL:     rpcURL,
		EthWSURL:      wsURL,
		ERC20Contract: contractAddr,
	}
}

func (c *Config) GetNodeURL() string {
	if c.EthWSURL != "" {
		return c.EthWSURL
	}
	return c.EthRPCURL
}
