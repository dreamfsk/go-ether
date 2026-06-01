package service

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
)

type BlockInfo struct {
	Number         uint64  `json:"number"`
	Hash           string  `json:"hash"`
	ParentHash     string  `json:"parentHash"`
	Timestamp      int64   `json:"timestamp"`
	TxCount        int     `json:"txCount"`
	GasUsed        uint64  `json:"gasUsed"`
	GasLimit       uint64  `json:"gasLimit"`
	GasUsedPercent float64 `json:"gasUsedPercent"`
}

type BlockService struct {
	client *client.EthClient
}

func NewBlockService(c *client.EthClient) *BlockService {
	return &BlockService{client: c}
}

func (s *BlockService) GetBlockByID(ctx context.Context, id string) (*BlockInfo, error) {
	var block *types.Block
	var err error

	if blockNumber, parseErr := strconv.ParseUint(id, 10, 64); parseErr == nil {
		block, err = s.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	} else {
		blockHash := common.HexToHash(id)
		block, err = s.client.BlockByHash(ctx, blockHash)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return convertBlock(block), nil
}

func convertBlock(block *types.Block) *BlockInfo {
	gasUsedPercent := 0.0
	if block.GasLimit() > 0 {
		gasUsedPercent = float64(block.GasUsed()) / float64(block.GasLimit()) * 100
	}

	return &BlockInfo{
		Number:         block.NumberU64(),
		Hash:           block.Hash().Hex(),
		ParentHash:     block.ParentHash().Hex(),
		Timestamp:      int64(block.Time()),
		TxCount:        len(block.Transactions()),
		GasUsed:        block.GasUsed(),
		GasLimit:       block.GasLimit(),
		GasUsedPercent: gasUsedPercent,
	}
}
