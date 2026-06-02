package config

import (
	"log"
	"os"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	Network       NetworkType
	NetworkConfig NetworkConfig
	ERC20Contract string
}

func Load() *Config {
	// 加载 .env 文件
	err := dotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// 获取网络配置
	networkName := os.Getenv("NETWORK")
	var network NetworkType
	if networkName == "" {
		network = GetDefaultNetwork()
	} else {
		network = NetworkType(networkName)
	}

	// 加载指定网络的配置
	networkConfig, err := GetNetworkConfig(network)
	if err != nil {
		log.Printf("Warning: failed to load network config for %s, using default", network)
		networkConfig, _ = GetNetworkConfig(GetDefaultNetwork())
	}

	// 允许环境变量覆盖 RPC/WS URL
	if customRPC := os.Getenv("ETH_RPC_URL"); customRPC != "" {
		networkConfig.RPCURL = customRPC
	}
	if customWS := os.Getenv("ETH_WS_URL"); customWS != "" {
		networkConfig.WSURL = customWS
	}

	contractAddr := os.Getenv("ERC20_CONTRACT")

	return &Config{
		Network:       network,
		NetworkConfig: networkConfig,
		ERC20Contract: contractAddr,
	}
}

func (c *Config) GetNodeURL() string {
	if c.NetworkConfig.WSURL != "" {
		return c.NetworkConfig.WSURL
	}
	return c.NetworkConfig.RPCURL
}
