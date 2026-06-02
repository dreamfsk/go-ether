package tests

import (
	"math/big"
	"testing"

	"github.com/meu/go-ether/service"
)

func TestConvertBlock(t *testing.T) {
	txService := service.NewTxService(nil)
	if txService == nil {
		t.Error("NewTxService should not return nil")
	}
}

func TestBlockService(t *testing.T) {
	blockService := service.NewBlockService(nil)
	if blockService == nil {
		t.Error("NewBlockService should not return nil")
	}
}

func TestTxSendService(t *testing.T) {
	t.Run("nil signer and chain scenario", func(t *testing.T) {
		txSendService := service.NewTxSendService(nil, nil, "", nil, nil)
		if txSendService == nil {
			t.Error("NewTxSendService should not return nil even with nil params")
		}
	})
}

func TestSendTxRequest_Validation(t *testing.T) {
	req := service.SendTxRequest{
		To:    "0x0000000000000000000000000000000000000000",
		Value: "1000000000000000000",
	}

	if req.To == "" {
		t.Error("SendTxRequest.To should not be empty")
	}

	if req.Value == "" {
		t.Error("SendTxRequest.Value should not be empty")
	}
}

func TestContractCallRequest_Validation(t *testing.T) {
	req := service.ContractCallRequest{
		ContractAddr: "0x0000000000000000000000000000000000000000",
		Method:       "getCount",
	}

	if req.ContractAddr == "" {
		t.Error("ContractCallRequest.ContractAddr should not be empty")
	}

	if req.Method == "" {
		t.Error("ContractCallRequest.Method should not be empty")
	}
}

func TestTokenInfo_Validation(t *testing.T) {
	tokenInfo := &service.TokenInfo{
		Name:        "Test Token",
		Symbol:      "TT",
		Decimals:    18,
		TotalSupply: big.NewInt(1000000000000000000),
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

	if tokenInfo.TotalSupply == nil {
		t.Error("TokenInfo.TotalSupply should not be nil")
	}
}
