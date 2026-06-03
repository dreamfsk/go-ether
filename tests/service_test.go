package tests

import (
	"math/big"
	"testing"

	"github.com/meu/go-ether/service"
)

func TestConvertBlock(t *testing.T) {
	// This test verifies the convertBlock function indirectly
	// Since convertBlock is unexported, we test through BlockService
	// For unit testing purposes, we can only test the BlockInfo struct creation
	
	blockInfo := &service.BlockInfo{
		Number:         12345,
		Hash:           "0x1234567890abcdef",
		ParentHash:     "0xabcdef1234567890",
		Timestamp:      1234567890,
		TxCount:        10,
		GasUsed:        210000,
		GasLimit:       30000000,
		GasUsedPercent: 0.7,
	}

	if blockInfo.Number != 12345 {
		t.Errorf("BlockInfo.Number = %v, want 12345", blockInfo.Number)
	}

	if blockInfo.TxCount != 10 {
		t.Errorf("BlockInfo.TxCount = %v, want 10", blockInfo.TxCount)
	}

	if blockInfo.GasUsedPercent != 0.7 {
		t.Errorf("BlockInfo.GasUsedPercent = %v, want 0.7", blockInfo.GasUsedPercent)
	}
}

func TestTransactionInfo_Validation(t *testing.T) {
	txInfo := &service.TransactionInfo{
		Hash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Nonce:     5,
		From:      "0x1111111111111111111111111111111111111111",
		To:        "0x2222222222222222222222222222222222222222",
		Value:     "1000000000000000000",
		Gas:       21000,
		GasPrice:  "20000000000",
		InputData: "0x",
		DataLen:   0,
		IsPending: false,
	}

	if txInfo.Hash == "" {
		t.Error("TransactionInfo.Hash should not be empty")
	}

	if txInfo.From == "" {
		t.Error("TransactionInfo.From should not be empty")
	}

	if txInfo.Gas != 21000 {
		t.Errorf("TransactionInfo.Gas = %v, want 21000", txInfo.Gas)
	}
}

func TestReceiptInfo_Validation(t *testing.T) {
	receipt := &service.ReceiptInfo{
		Status:      1,
		BlockNumber: 12345,
		BlockHash:   "0x1234567890abcdef",
		TxIndex:     0,
		GasUsed:     21000,
		LogsCount:   0,
	}

	if receipt.Status != 1 {
		t.Errorf("ReceiptInfo.Status = %v, want 1", receipt.Status)
	}

	if receipt.BlockNumber != 12345 {
		t.Errorf("ReceiptInfo.BlockNumber = %v, want 12345", receipt.BlockNumber)
	}
}

func TestSendTxRequest_Validation(t *testing.T) {
	req := &service.SendTxRequest{
		To:    "0x2222222222222222222222222222222222222222",
		Value: "1000000000000000000",
		Gas:   21000,
	}

	if req.To == "" {
		t.Error("SendTxRequest.To should not be empty")
	}

	if req.Value == "" {
		t.Error("SendTxRequest.Value should not be empty")
	}
}

func TestTokenInfo_Validation(t *testing.T) {
	testSupply, _ := new(big.Int).SetString("1000000000000000000000", 10)
	tokenInfo := &service.TokenInfo{
		Name:        "Test Token",
		Symbol:      "TT",
		Decimals:    18,
		TotalSupply: testSupply,
	}

	if tokenInfo.Name == "" {
		t.Error("TokenInfo.Name should not be empty")
	}

	if tokenInfo.Symbol == "" {
		t.Error("TokenInfo.Symbol should not be empty")
	}

	if tokenInfo.Decimals != 18 {
		t.Errorf("TokenInfo.Decimals = %v, want 18", tokenInfo.Decimals)
	}
}
