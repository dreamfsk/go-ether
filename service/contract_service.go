package service

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/wallet"
)

type ContractService struct {
	client  *client.EthClient
	signer  wallet.Signer
	network config.NetworkType
	chainID *big.Int
}

func NewContractService(c *client.EthClient, signer wallet.Signer, network config.NetworkType, chainID *big.Int) *ContractService {
	return &ContractService{
		client:  c,
		signer:  signer,
		network: network,
		chainID: chainID,
	}
}

type ContractCallRequest struct {
	ContractAddr string   `json:"contractAddr"`
	Method       string   `json:"method"`
	Args         []string `json:"args,omitempty"`
}

type ContractCallResponse struct {
	TxHash   string `json:"txHash,omitempty"`
	Result   string `json:"result,omitempty"`
	Status   string `json:"status"`
	GasUsed  uint64 `json:"gasUsed,omitempty"`
}

func (s *ContractService) CallViewMethod(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	log.Printf("🔍 [ContractService] 调用视图方法: %s, 合约: %s", req.Method, req.ContractAddr)

	contractAddr := common.HexToAddress(req.ContractAddr)

	switch req.Method {
	case "count", "getCount":
		data, err := abiEncodeGetCount()
		if err != nil {
			return nil, fmt.Errorf("failed to encode data: %w", err)
		}

		result, err := s.client.CallContract(ctx, ethereum.CallMsg{
			To:   &contractAddr,
			Data: data,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to call contract: %w", err)
		}

		count := new(big.Int).SetBytes(result)
		return &ContractCallResponse{
			Result: count.String(),
			Status: "success",
		}, nil

	case "owner":
		data, err := abiEncodeOwner()
		if err != nil {
			return nil, fmt.Errorf("failed to encode data: %w", err)
		}

		result, err := s.client.CallContract(ctx, ethereum.CallMsg{
			To:   &contractAddr,
			Data: data,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to call contract: %w", err)
		}

		owner := common.BytesToAddress(result)
		return &ContractCallResponse{
			Result: owner.Hex(),
			Status: "success",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported view method: %s", req.Method)
	}
}

func (s *ContractService) SendTransaction(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("signer not initialized")
	}

	log.Printf("✉️ [ContractService] 发送交易调用: %s, 合约: %s", req.Method, req.ContractAddr)

	contractAddr := common.HexToAddress(req.ContractAddr)
	fromAddr := s.signer.Address()

	var data []byte
	var err error

	switch req.Method {
	case "increment":
		data, err = abiEncodeIncrement()
	case "decrement":
		data, err = abiEncodeDecrement()
	case "setCount":
		if len(req.Args) < 1 {
			return nil, fmt.Errorf("setCount requires 1 argument")
		}
		count, ok := new(big.Int).SetString(req.Args[0], 10)
		if !ok {
			return nil, fmt.Errorf("invalid count value")
		}
		data, err = abiEncodeSetCount(count)
	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
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

	gasLimit := uint64(50000)

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
	log.Printf("✅ [ContractService] 合约交易发送成功: %s", txHash)

	return &ContractCallResponse{
		TxHash: txHash,
		Status: "pending",
	}, nil
}

func abiEncodeGetCount() ([]byte, error) {
	return []byte{0x06, 0xfd, 0xde, 0x03}, nil
}

func abiEncodeOwner() ([]byte, error) {
	return []byte{0x8d, 0xa5, 0xcb, 0x5b}, nil
}

func abiEncodeIncrement() ([]byte, error) {
	return []byte{0x16, 0x21, 0x6f, 0x39}, nil
}

func abiEncodeDecrement() ([]byte, error) {
	return []byte{0x0d, 0xe6, 0x7e, 0x2b}, nil
}

func abiEncodeSetCount(count *big.Int) ([]byte, error) {
	args, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	encoded, err := abi.Arguments{{Type: args}}.Pack(count)
	if err != nil {
		return nil, err
	}

	methodID := []byte{0x67, 0xe0, 0x85, 0x11}
	return append(methodID, encoded...), nil
}
