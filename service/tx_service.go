package service

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
)

type TransactionInfo struct {
	Hash      string       `json:"hash"`
	Nonce     uint64       `json:"nonce"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	Value     string       `json:"value"`
	Gas       uint64       `json:"gas"`
	GasPrice  string       `json:"gasPrice"`
	DataLen   int          `json:"dataLen"`
	IsPending bool         `json:"isPending"`
	Receipt   *ReceiptInfo `json:"receipt,omitempty"`
}

type ReceiptInfo struct {
	Status      uint64 `json:"status"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	TxIndex     uint   `json:"txIndex"`
	GasUsed     uint64 `json:"gasUsed"`
	LogsCount   int    `json:"logsCount"`
}

type TxService struct {
	client *client.EthClient
}

func NewTxService(c *client.EthClient) *TxService {
	return &TxService{client: c}
}

func (s *TxService) GetTransactionByHash(ctx context.Context, hash string) (*TransactionInfo, error) {
	txHash := common.HexToHash(hash)
	tx, isPending, err := s.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		from = common.Address{}
	}

	toAddr := ""
	if tx.To() != nil {
		toAddr = tx.To().Hex()
	}

	txInfo := &TransactionInfo{
		Hash:      tx.Hash().Hex(),
		Nonce:     tx.Nonce(),
		From:      from.Hex(),
		To:        toAddr,
		Value:     tx.Value().String(),
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice().String(),
		DataLen:   len(tx.Data()),
		IsPending: isPending,
	}

	if !isPending {
		receipt, err := s.client.TransactionReceipt(ctx, txHash)
		if err == nil {
			txInfo.Receipt = &ReceiptInfo{
				Status:      receipt.Status,
				BlockNumber: receipt.BlockNumber.Uint64(),
				BlockHash:   receipt.BlockHash.Hex(),
				TxIndex:     receipt.TransactionIndex,
				GasUsed:     receipt.GasUsed,
				LogsCount:   len(receipt.Logs),
			}
		}
	}

	return txInfo, nil
}
