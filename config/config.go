package config

import (
	"os"
	"log"
	dotenv "github.com/joho/godotenv"
)

type Config struct {
	EthRPCURL     string
	EthWSURL      string
	ERC20Contract string
}

func Load() *Config {
		// 加载 .env 文件
	err := dotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	dotenv.Load()
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
