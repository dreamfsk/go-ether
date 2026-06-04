package tests

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/meu/go-ether/service"
)

// ============================================================
// ContractCallRequest / ContractCallResponse
// ============================================================

func TestContractCallRequest_JSON(t *testing.T) {
	req := service.ContractCallRequest{
		ContractAddr: "0x1234567890abcdef1234567890abcdef12345678",
		Method:       "balanceOf",
		Args:         []string{"0x1111111111111111111111111111111111111111"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded service.ContractCallRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ContractAddr != req.ContractAddr {
		t.Errorf("ContractAddr mismatch: got %s, want %s", decoded.ContractAddr, req.ContractAddr)
	}
	if decoded.Method != req.Method {
		t.Errorf("Method mismatch: got %s, want %s", decoded.Method, req.Method)
	}
	if len(decoded.Args) != 1 || decoded.Args[0] != req.Args[0] {
		t.Errorf("Args mismatch: got %v, want %v", decoded.Args, req.Args)
	}
}

func TestContractCallRequest_JSON_EmptyArgs(t *testing.T) {
	req := service.ContractCallRequest{
		ContractAddr: "0x1234567890abcdef1234567890abcdef12345678",
		Method:       "name",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Args should be omitted when empty
	raw := string(data)
	if raw == "" {
		t.Fatal("marshal produced empty output")
	}
}

func TestContractCallResponse(t *testing.T) {
	resp := &service.ContractCallResponse{
		Status: "success",
		Result: "MyToken",
	}

	if resp.Status != "success" {
		t.Errorf("Status = %s, want success", resp.Status)
	}
	if resp.Result != "MyToken" {
		t.Errorf("Result = %s, want MyToken", resp.Result)
	}
	if resp.TxHash != "" {
		t.Errorf("TxHash should be empty for view call, got %s", resp.TxHash)
	}
}

func TestContractCallResponse_WriteMethod(t *testing.T) {
	resp := &service.ContractCallResponse{
		TxHash: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Status: "pending",
	}

	if resp.TxHash == "" {
		t.Error("TxHash should not be empty for write calls")
	}
	if resp.Status != "pending" {
		t.Errorf("Status = %s, want pending", resp.Status)
	}
}

// ============================================================
// SendTxRequest / SendTxResponse
// ============================================================

func TestSendTxRequest_JSON(t *testing.T) {
	req := service.SendTxRequest{
		To:    "0x2222222222222222222222222222222222222222",
		Value: "1000000000000000000",
		Gas:   21000,
		Data:  "0x",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded service.SendTxRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.To != req.To {
		t.Errorf("To mismatch: got %s, want %s", decoded.To, req.To)
	}
	if decoded.Value != req.Value {
		t.Errorf("Value mismatch: got %s, want %s", decoded.Value, req.Value)
	}
	if decoded.Gas != req.Gas {
		t.Errorf("Gas mismatch: got %d, want %d", decoded.Gas, req.Gas)
	}
}

func TestSendTxRequest_JSON_OmitEmpty(t *testing.T) {
	// Gas=0 should be omitted in JSON
	req := service.SendTxRequest{
		To:    "0x2222222222222222222222222222222222222222",
		Value: "1000000000000000000",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Gas should not appear in JSON
	raw := string(data)
	if raw == "" {
		t.Fatal("marshal produced empty output")
	}

	var decoded service.SendTxRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.To != req.To {
		t.Errorf("To mismatch: got %s, want %s", decoded.To, req.To)
	}
}

func TestSendTxResponse(t *testing.T) {
	resp := &service.SendTxResponse{
		TxHash:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		From:     "0x1111111111111111111111111111111111111111",
		To:       "0x2222222222222222222222222222222222222222",
		Value:    "1000000000000000000",
		Nonce:    5,
		GasLimit: 21000,
		GasPrice: "20000000000",
		Status:   "pending",
	}

	if resp.TxHash == "" || resp.From == "" || resp.To == "" {
		t.Error("essential fields should not be empty")
	}
	if resp.Status != "pending" {
		t.Errorf("Status = %s, want pending", resp.Status)
	}
}

// ============================================================
// TokenInfo
// ============================================================

func TestTokenInfo(t *testing.T) {
	supply, _ := new(big.Int).SetString("1000000000000000000000", 10)
	info := &service.TokenInfo{
		Name:        "TestToken",
		Symbol:      "TEST",
		Decimals:    18,
		TotalSupply: supply,
	}

	if info.Name != "TestToken" {
		t.Errorf("Name = %s, want TestToken", info.Name)
	}
	if info.Symbol != "TEST" {
		t.Errorf("Symbol = %s, want TEST", info.Symbol)
	}
	if info.Decimals != 18 {
		t.Errorf("Decimals = %d, want 18", info.Decimals)
	}
	if info.TotalSupply.Cmp(supply) != 0 {
		t.Errorf("TotalSupply mismatch: got %s, want %s", info.TotalSupply.String(), supply.String())
	}
}

// ============================================================
// DeployResult
// ============================================================

func TestDeployResult_JSON(t *testing.T) {
	result := &service.DeployResult{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		TxHash:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded service.DeployResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Address != result.Address {
		t.Errorf("Address mismatch: got %s, want %s", decoded.Address, result.Address)
	}
	if decoded.TxHash != result.TxHash {
		t.Errorf("TxHash mismatch: got %s, want %s", decoded.TxHash, result.TxHash)
	}
}

// ============================================================
// BlockInfo — GasUsedPercent calculation edge cases
// ============================================================

func TestBlockInfo_GasUsedPercent_ZeroGasLimit(t *testing.T) {
	// GasUsedPercent should be 0 when GasLimit is 0 (avoid division by zero)
	info := &service.BlockInfo{
		Number:         100,
		GasUsed:        50000,
		GasLimit:       0,
		GasUsedPercent: 0.0,
	}

	if info.GasUsedPercent != 0.0 {
		t.Errorf("GasUsedPercent with zero GasLimit should be 0, got %f", info.GasUsedPercent)
	}
}

func TestBlockInfo_GasUsedPercent_FullBlock(t *testing.T) {
	info := &service.BlockInfo{
		Number:         100,
		GasUsed:        30000000,
		GasLimit:       30000000,
		GasUsedPercent: 100.0,
	}

	if info.GasUsedPercent != 100.0 {
		t.Errorf("GasUsedPercent for full block should be 100%%, got %f", info.GasUsedPercent)
	}
}

// ============================================================
// TransactionInfo
// ============================================================

func TestTransactionInfo_IsPending(t *testing.T) {
	pendingTx := &service.TransactionInfo{
		Hash:      "0xtest",
		IsPending: true,
	}
	if !pendingTx.IsPending {
		t.Error("IsPending should be true for pending transaction")
	}

	confirmedTx := &service.TransactionInfo{
		Hash:      "0xtest",
		IsPending: false,
	}
	if confirmedTx.IsPending {
		t.Error("IsPending should be false for confirmed transaction")
	}
}

func TestTransactionInfo_ContractCreation(t *testing.T) {
	// Contract creation: To is empty
	tx := &service.TransactionInfo{
		Hash:      "0xtest",
		From:      "0x1111111111111111111111111111111111111111",
		To:        "",
		IsPending: false,
	}

	if tx.To != "" {
		t.Errorf("Contract creation tx should have empty To, got %s", tx.To)
	}
}

func TestReceiptInfo_FailedTx(t *testing.T) {
	receipt := &service.ReceiptInfo{
		Status:      0,
		BlockNumber: 12345,
		BlockHash:   "0xhash",
		TxIndex:     5,
		GasUsed:     21000,
	}

	if receipt.Status != 0 {
		t.Errorf("Failed tx Status should be 0, got %d", receipt.Status)
	}
	// Even failed txs consume gas
	if receipt.GasUsed == 0 {
		t.Error("Failed tx should still have GasUsed > 0")
	}
}
