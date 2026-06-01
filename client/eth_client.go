package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	*ethclient.Client
}

func New(ctx context.Context, url string) (*EthClient, error) {
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to dial Ethereum node: %w", err)
	}
	return &EthClient{Client: client}, nil
}
