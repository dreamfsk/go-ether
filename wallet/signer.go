package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer interface {
	SignTx(ctx context.Context, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	Address() common.Address
	// TransactOpts 创建用于合约交互的 TransactOpts（内部使用私钥，不暴露）
	TransactOpts(ctx context.Context, chainID *big.Int) (*bind.TransactOpts, error)
}

type EnvSigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewEnvSigner() (*EnvSigner, error) {
	privateKeyHex := os.Getenv("SENDER_PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, fmt.Errorf("SENDER_PRIVATE_KEY environment variable is empty")
	}

	// 移除可能的 0x 前缀
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// 验证私钥长度
	if len(privateKeyHex) != 64 {
		return nil, fmt.Errorf("invalid private key length: expected 64 characters, got %d", len(privateKeyHex))
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &EnvSigner{
		privateKey: privateKey,
		address:    address,
	}, nil
}

func (s *EnvSigner) SignTx(ctx context.Context, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	var signer types.Signer
	
	// 根据交易类型选择合适的签名器
	switch tx.Type() {
	case types.LegacyTxType:
		// Legacy 交易使用 Homestead 签名器
		signer = types.NewEIP155Signer(chainID)
	case types.DynamicFeeTxType:
		// EIP-1559 交易使用最新签名器
		signer = types.LatestSignerForChainID(chainID)
	default:
		// 默认使用最新签名器
		signer = types.LatestSignerForChainID(chainID)
	}
	
	signedTx, err := types.SignTx(tx, signer, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	return signedTx, nil
}

func (s *EnvSigner) Address() common.Address {
	return s.address
}

// TransactOpts 创建用于合约交互的 TransactOpts，隐藏私钥细节
func (s *EnvSigner) TransactOpts(ctx context.Context, chainID *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.privateKey, chainID)
}
