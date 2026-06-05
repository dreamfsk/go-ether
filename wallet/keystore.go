package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type KeystoreSigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	keyDir     string
}

// NewKeystoreSigner 从 keystore 文件创建签名器
func NewKeystoreSigner(keyDir, password string) (*KeystoreSigner, error) {
	if keyDir == "" {
		// 使用默认路径
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		keyDir = filepath.Join(home, ".ethereum", "keystore")
	}

	// 检查目录是否存在
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("keystore directory does not exist: %s", keyDir)
	}

	// 读取 keystore 文件
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no keystore files found in: %s", keyDir)
	}

	// 找到第一个 keystore 文件
	var keyJSON []byte
	var keyFile string
	for _, f := range files {
		if !f.IsDir() {
			keyFile = filepath.Join(keyDir, f.Name())
			keyJSON, err = ioutil.ReadFile(keyFile)
			if err == nil {
				break
			}
		}
	}

	if len(keyJSON) == 0 {
		return nil, fmt.Errorf("failed to read any keystore file in: %s", keyDir)
	}

	// 解密私钥
	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	address := key.Address

	return &KeystoreSigner{
		privateKey: key.PrivateKey,
		address:    address,
		keyDir:    keyDir,
	}, nil
}

// NewKeystoreSignerFromPath 从指定的 keystore 文件路径创建签名器
func NewKeystoreSignerFromPath(keyPath, password string) (*KeystoreSigner, error) {
	if keyPath == "" {
		return nil, fmt.Errorf("keystore path is required")
	}

	// 检查文件是否存在
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("keystore file does not exist: %s", keyPath)
	}

	// 读取 keystore 文件
	keyJSON, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// 解密私钥
	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	address := key.Address

	return &KeystoreSigner{
		privateKey: key.PrivateKey,
		address:    address,
		keyDir:    filepath.Dir(keyPath),
	}, nil
}

func (s *KeystoreSigner) SignTx(ctx context.Context, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
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

func (s *KeystoreSigner) Address() common.Address {
	return s.address
}

// TransactOpts 创建用于合约交互的 TransactOpts，隐藏私钥细节
func (s *KeystoreSigner) TransactOpts(ctx context.Context, chainID *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.privateKey, chainID)
}
