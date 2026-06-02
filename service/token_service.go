package service

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/wallet"
)

type TokenService struct {
	client  *client.EthClient
	signer  wallet.Signer
	network config.NetworkType
	chainID *big.Int
}

func NewTokenService(c *client.EthClient, signer wallet.Signer, network config.NetworkType, chainID *big.Int) *TokenService {
	return &TokenService{
		client:  c,
		signer:  signer,
		network: network,
		chainID: chainID,
	}
}

type TokenInfo struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
	TotalSupply string `json:"totalSupply"`
}

type TokenTransferRequest struct {
	TokenAddr string `json:"tokenAddr"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
}

type TokenTransferResponse struct {
	TxHash   string `json:"txHash"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Status   string `json:"status"`
}

func (s *TokenService) GetTokenInfo(ctx context.Context, tokenAddr string) (*TokenInfo, error) {
	log.Printf("🔍 [TokenService] 查询代币信息: %s", tokenAddr)

	contractAddr := common.HexToAddress(tokenAddr)

	name, err := s.callTokenName(ctx, contractAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get name: %w", err)
	}

	symbol, err := s.callTokenSymbol(ctx, contractAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	decimals, err := s.callTokenDecimals(ctx, contractAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals: %w", err)
	}

	totalSupply, err := s.callTokenTotalSupply(ctx, contractAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	return &TokenInfo{
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: totalSupply.String(),
	}, nil
}

func (s *TokenService) GetBalance(ctx context.Context, tokenAddr, holderAddr string) (string, error) {
	log.Printf("🔍 [TokenService] 查询余额: token=%s, holder=%s", tokenAddr, holderAddr)

	contractAddr := common.HexToAddress(tokenAddr)
	holder := common.HexToAddress(holderAddr)

	balance, err := s.callTokenBalanceOf(ctx, contractAddr, holder)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.String(), nil
}

func (s *TokenService) Transfer(ctx context.Context, req TokenTransferRequest) (*TokenTransferResponse, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("signer not initialized")
	}

	log.Printf("💸 [TokenService] 代币转账: token=%s, to=%s, amount=%s", req.TokenAddr, req.To, req.Amount)

	contractAddr := common.HexToAddress(req.TokenAddr)
	toAddr := common.HexToAddress(req.To)
	fromAddr := s.signer.Address()

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	data, err := encodeTransfer(toAddr, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to encode transfer: %w", err)
	}

	nonce, err := s.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasTipCap, err := s.client.SuggestGasTipCap(ctx)
	if err != nil {
		gasTipCap = big.NewInt(1_000_000_000)
	}

	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %w", err)
	}

	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(1_000_000_000)
	}

	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	gasLimit := uint64(100000)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   s.chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &contractAddr,
		Value:     big.NewInt(0),
		Data:      data,
	})

	signedTx, err := s.signer.SignTx(ctx, tx, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	log.Printf("✅ [TokenService] 代币转账交易发送成功: %s", txHash)

	return &TokenTransferResponse{
		TxHash: txHash,
		From:   fromAddr.Hex(),
		To:     toAddr.Hex(),
		Amount: amount.String(),
		Status: "pending",
	}, nil
}

func (s *TokenService) callTokenName(ctx context.Context, contractAddr common.Address) (string, error) {
	data := []byte{0x06, 0xfd, 0xde, 0x03}

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", err
	}

	return parseStringResult(result), nil
}

func (s *TokenService) callTokenSymbol(ctx context.Context, contractAddr common.Address) (string, error) {
	data := []byte{0x95, 0xd8, 0x9b, 0x41}

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", err
	}

	return parseStringResult(result), nil
}

func (s *TokenService) callTokenDecimals(ctx context.Context, contractAddr common.Address) (uint8, error) {
	data := []byte{0x31, 0x3c, 0xe5, 0x67}

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 18, nil
	}

	return uint8(result[31]), nil
}

func (s *TokenService) callTokenTotalSupply(ctx context.Context, contractAddr common.Address) (*big.Int, error) {
	data := []byte{0x18, 0x16, 0x0d, 0xdd}

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

func (s *TokenService) callTokenBalanceOf(ctx context.Context, contractAddr, holder common.Address) (*big.Int, error) {
	data := append([]byte{0x70, 0xa0, 0x82, 0x31}, common.LeftPadBytes(holder.Bytes(), 32)...)

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result), nil
}

func encodeTransfer(to common.Address, amount *big.Int) ([]byte, error) {
	data := make([]byte, 4+32+32)
	copy(data[0:4], []byte{0xa9, 0x05, 0x9c, 0xbb})
	copy(data[4:36], common.LeftPadBytes(to.Bytes(), 32))
	copy(data[36:68], common.LeftPadBytes(amount.Bytes(), 32))
	return data, nil
}

func parseStringResult(data []byte) string {
	if len(data) < 64 {
		return ""
	}

	length := new(big.Int).SetBytes(data[32:64]).Uint64()
	if length == 0 {
		return ""
	}

	if int(length)+64 > len(data) {
		return ""
	}

	return string(data[64 : 64+length])
}
