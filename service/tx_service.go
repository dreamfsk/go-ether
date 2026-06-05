package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum"
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

type GasFeeSuggestion struct {
	GasPrice       string `json:"gasPrice"`       // Legacy 交易的 gas price (wei)
	GasTipCap      string `json:"gasTipCap"`      // EIP-1559 的优先费用 (wei)
	GasFeeCap      string `json:"gasFeeCap"`      // EIP-1559 的最大费用 (wei)
	BaseFee        string `json:"baseFee"`        // 当前区块基础费用 (wei)
	EstimatedCost  string `json:"estimatedCost"`  // 预估交易费用 (wei)
	EstimatedCostETH string `json:"estimatedCostETH"` // 预估交易费用 (ETH)
	SupportsEIP1559 bool  `json:"supportsEIP1559"` // 是否支持 EIP-1559
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

// GetClient 返回 RPC 客户端
func (s *TxService) GetClient() *client.EthClient {
	return s.client
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

func (s *TxService) GetGasFeeSuggestion(ctx context.Context) (*GasFeeSuggestion, error) {
	log.Println("🔍 [TxService] 获取 Gas 费用建议")

	// 获取建议的 gas price（用于 Legacy 交易）
	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		log.Printf("⚠️ [TxService] 获取建议 Gas Price 失败: %v", err)
		gasPrice = big.NewInt(1_000_000_000) // 默认 1 Gwei
	}

	// 获取建议的 gas tip cap（用于 EIP-1559 交易）
	gasTipCap, err := s.client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Printf("⚠️ [TxService] 获取建议 Gas Tip 失败: %v", err)
		gasTipCap = big.NewInt(1_000_000_000) // 默认 1 Gwei
	}

	// 获取当前区块信息
	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("⚠️ [TxService] 获取区块头失败: %v", err)
		return nil, fmt.Errorf("failed to get header: %w", err)
	}

	// 判断是否支持 EIP-1559
	supportsEIP1559 := header.BaseFee != nil
	
	var baseFee, gasFeeCap *big.Int
	if supportsEIP1559 {
		baseFee = header.BaseFee
		// 计算 gasFeeCap = baseFee * 2 + gasTipCap
		gasFeeCap = new(big.Int).Add(
			new(big.Int).Mul(baseFee, big.NewInt(2)),
			gasTipCap,
		)
	} else {
		baseFee = big.NewInt(0)
		gasFeeCap = big.NewInt(0)
	}

	// 预估交易费用（使用默认 gas limit 21000）
	gasLimit := uint64(21000)
	estimatedCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	
	// 转换为 ETH
	ethDivisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	estimatedCostETH := new(big.Float).Quo(
		new(big.Float).SetInt(estimatedCost),
		new(big.Float).SetInt(ethDivisor),
	).Text('f', 18)

	suggestion := &GasFeeSuggestion{
		GasPrice:        gasPrice.String(),
		GasTipCap:       gasTipCap.String(),
		GasFeeCap:       gasFeeCap.String(),
		BaseFee:         baseFee.String(),
		EstimatedCost:   estimatedCost.String(),
		EstimatedCostETH: estimatedCostETH,
		SupportsEIP1559: supportsEIP1559,
	}

	log.Printf("✅ [TxService] Gas 费用建议获取成功 - GasPrice: %s wei, SupportsEIP1559: %v", 
		gasPrice.String(), supportsEIP1559)

	return suggestion, nil
}

// EstimateGas 估算交易所需的 gas
func (s *TxService) EstimateGas(ctx context.Context, from, to string, value string) (uint64, error) {
	log.Printf("🔍 [TxService] 估算 Gas - From: %s, To: %s, Value: %s", from, to, value)

	toAddr := common.HexToAddress(to)
	fromAddr := common.HexToAddress(from)

	valueBig, ok := new(big.Int).SetString(value, 10)
	if !ok {
		valueBig = big.NewInt(0)
	}

	msg := ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: valueBig,
	}

	gas, err := s.client.EstimateGas(ctx, msg)
	if err != nil {
		log.Printf("⚠️ [TxService] Gas 估算失败: %v, 使用默认值 21000", err)
		return 21000, nil
	}

	// 增加 20% 的安全余量
	gasWithMargin := gas * 120 / 100

	log.Printf("✅ [TxService] Gas 估算成功: %d (含余量: %d)", gas, gasWithMargin)
	return gasWithMargin, nil
}
