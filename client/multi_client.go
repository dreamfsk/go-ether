package client

import (
	"context"
	"fmt"

	"github.com/meu/go-ether/config"
)

type MultiClient struct {
	Sepolia *EthClient
	Local   *EthClient
}

func NewMultiClient(ctx context.Context, cfg *config.Config) (*MultiClient, error) {
	clients := &MultiClient{}

	// 初始化默认网络客户端
	defaultClient, err := New(ctx, cfg.GetNodeURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default network: %w", err)
	}

	// 根据网络类型分配
	switch cfg.Network {
	case config.NetworkSepolia:
		clients.Sepolia = defaultClient
	case config.NetworkLocal:
		clients.Local = defaultClient
	}

	return clients, nil
}

func (mc *MultiClient) GetClient(network config.NetworkType) (*EthClient, error) {
	switch network {
	case config.NetworkSepolia:
		if mc.Sepolia == nil {
			return nil, fmt.Errorf("sepolia client not initialized")
		}
		return mc.Sepolia, nil
	case config.NetworkLocal:
		if mc.Local == nil {
			return nil, fmt.Errorf("local client not initialized")
		}
		return mc.Local, nil
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

func (mc *MultiClient) Close() {
	if mc.Sepolia != nil {
		mc.Sepolia.Client.Close()
	}
	if mc.Local != nil {
		mc.Local.Client.Close()
	}
}
