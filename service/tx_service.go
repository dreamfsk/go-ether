package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

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
	InputData string       `json:"inputData"`
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
	log.Println("🔧 [TxService] 初始化")
	return &TxService{client: c}
}

func (s *TxService) GetTransactionByHash(ctx context.Context, hash string) (*TransactionInfo, error) {
	log.Printf("🔍 [TxService] 查询交易: %s", hash)
	
	txHash := common.HexToHash(hash)
	tx, isPending, err := s.client.TransactionByHash(ctx, txHash)
	if err != nil {
		log.Printf("❌ [TxService] 查询交易失败: %v", err)
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
		InputData: "0x" + hex.EncodeToString(tx.Data()),
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

	status := "pending"
	if !isPending {
		status = "confirmed"
	}
	log.Printf("✅ [TxService] 交易查询成功 (%s): 从 %s 到 %s", status, txInfo.From, txInfo.To)
	
	return txInfo, nil
}
