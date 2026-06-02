package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/contracts"
	"github.com/meu/go-ether/wallet"
)

// ERC20Service 提供与 ERC20 代币合约的交互
type ERC20Service struct {
	client       *client.EthClient
	signer       wallet.Signer
	contract     *contracts.MyERC20
	contractAddr common.Address
	chainID      *big.Int
}

// NewERC20Service 创建 ERC20 服务实例
func NewERC20Service(c *client.EthClient, signer wallet.Signer, network config.NetworkType, chainID *big.Int, contractAddr string) (*ERC20Service, error) {
	addr := common.HexToAddress(contractAddr)
	erc20, err := contracts.NewMyERC20(addr, c)
	if err != nil {
		return nil, fmt.Errorf("failed to bind contract: %w", err)
	}

	return &ERC20Service{
		client:       c,
		signer:        signer,
		contract:      erc20,
		contractAddr:  addr,
		chainID:       chainID,
	}, nil
}

// TokenInfo 代币信息
type TokenInfo struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}

// GetTokenInfo 获取代币基本信息
func (s *ERC20Service) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	name, err := s.contract.Name(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get name: %w", err)
	}

	symbol, err := s.contract.Symbol(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	decimals, err := s.contract.Decimals(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals: %w", err)
	}

	totalSupply, err := s.contract.TotalSupply(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	return &TokenInfo{
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: totalSupply,
	}, nil
}

// BalanceOf 查询代币余额
func (s *ERC20Service) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := s.contract.BalanceOf(nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// Transfer 发送代币
func (s *ERC20Service) Transfer(ctx context.Context, to string, amount *big.Int) (string, error) {
	if s.signer == nil {
		return "", fmt.Errorf("signer not initialized")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.signer.(*wallet.EnvSigner).PrivateKey(), s.chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}
	auth.Nonce = big.NewInt(int64(s.getNonce(ctx)))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 65000
	auth.GasPrice = s.getGasPrice(ctx)

	toAddr := common.HexToAddress(to)

	tx, err := s.contract.Transfer(auth, toAddr, amount)
	if err != nil {
		return "", fmt.Errorf("failed to transfer: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("✅ [ERC20Service] 代币转账交易已发送: %s", txHash)

	return txHash, nil
}

// Mint 铸造代币
func (s *ERC20Service) Mint(ctx context.Context, to string, amount *big.Int) (string, error) {
	if s.signer == nil {
		return "", fmt.Errorf("signer not initialized")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.signer.(*wallet.EnvSigner).PrivateKey(), s.chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}
	auth.Nonce = big.NewInt(int64(s.getNonce(ctx)))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 100000
	auth.GasPrice = s.getGasPrice(ctx)

	toAddr := common.HexToAddress(to)

	tx, err := s.contract.Mint(auth, toAddr, amount)
	if err != nil {
		return "", fmt.Errorf("failed to mint: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("✅ [ERC20Service] 代币铸造交易已发送: %s", txHash)

	return txHash, nil
}

// DeployResult 合约部署结果
type DeployResult struct {
	Address string `json:"address"`
	TxHash  string `json:"txHash"`
}

// DeployERC20 部署 MyERC20 合约（包级函数）
func DeployERC20(ctx context.Context, c *client.EthClient, signer wallet.Signer, name, symbol string, initialSupply *big.Int, recipient string) (*DeployResult, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer not initialized")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(signer.(*wallet.EnvSigner).PrivateKey(), big.NewInt(1))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	auth.Value = big.NewInt(0)
	auth.GasLimit = 2000000

	nonce, err := c.PendingNonceAt(ctx, signer.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		gasPrice = big.NewInt(1_000_000_000)
	}
	auth.GasPrice = gasPrice

	recipientAddr := common.HexToAddress(recipient)

	addr, tx, _, err := contracts.DeployMyERC20(auth, c, name, symbol, initialSupply, recipientAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("✅ [DeployERC20] 合约部署交易已发送: %s, 合约地址: %s", txHash, addr.Hex())

	return &DeployResult{
		Address: addr.Hex(),
		TxHash:  txHash,
	}, nil
}

// Deploy 部署 MyERC20 合约（实例方法）
func (s *ERC20Service) Deploy(ctx context.Context, name, symbol string, initialSupply *big.Int, recipient string) (*DeployResult, error) {
	return DeployERC20(ctx, s.client, s.signer, name, symbol, initialSupply, recipient)
}

// getNonce 获取待处理交易数
func (s *ERC20Service) getNonce(ctx context.Context) uint64 {
	nonce, err := s.client.PendingNonceAt(ctx, s.signer.Address())
	if err != nil {
		return 0
	}
	return nonce
}

// getGasPrice 获取 Gas 价格
func (s *ERC20Service) getGasPrice(ctx context.Context) *big.Int {
	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return big.NewInt(1_000_000_000)
	}
	return gasPrice
}

// Allowance 查询授权额度
func (s *ERC20Service) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	ownerAddr := common.HexToAddress(owner)
	spenderAddr := common.HexToAddress(spender)

	allowance, err := s.contract.Allowance(nil, ownerAddr, spenderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}
	return allowance, nil
}

// Approve 授权代币
func (s *ERC20Service) Approve(ctx context.Context, spender string, amount *big.Int) (string, error) {
	if s.signer == nil {
		return "", fmt.Errorf("signer not initialized")
	}

	auth := bind.NewKeyedTransactor(s.signer.(*wallet.EnvSigner).PrivateKey())
	auth.Nonce = big.NewInt(int64(s.getNonce(ctx)))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 65000
	auth.GasPrice = s.getGasPrice(ctx)

	spenderAddr := common.HexToAddress(spender)

	tx, err := s.contract.Approve(auth, spenderAddr, amount)
	if err != nil {
		return "", fmt.Errorf("failed to approve: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("✅ [ERC20Service] 代币授权交易已发送: %s", txHash)

	return txHash, nil
}

// TransferFrom 从授权转账
func (s *ERC20Service) TransferFrom(ctx context.Context, from, to string, amount *big.Int) (string, error) {
	if s.signer == nil {
		return "", fmt.Errorf("signer not initialized")
	}

	auth := bind.NewKeyedTransactor(s.signer.(*wallet.EnvSigner).PrivateKey())
	auth.Nonce = big.NewInt(int64(s.getNonce(ctx)))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 65000
	auth.GasPrice = s.getGasPrice(ctx)

	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	tx, err := s.contract.TransferFrom(auth, fromAddr, toAddr, amount)
	if err != nil {
		return "", fmt.Errorf("failed to transfer from: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf("✅ [ERC20Service] 代币授权转账交易已发送: %s", txHash)

	return txHash, nil
}

// ContractAddress 返回合约地址
func (s *ERC20Service) ContractAddress() common.Address {
	return s.contractAddr
}

// WaitForConfirmation 等待交易确认
func (s *ERC20Service) WaitForConfirmation(ctx context.Context, txHash string, confirmations uint64) (*types.Receipt, error) {
	receipt, err := s.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	if confirmations > 1 && receipt.Status == 1 {
		currentBlock := receipt.BlockNumber
		for {
			select {
			case <-ctx.Done():
				return receipt, ctx.Err()
			default:
				header, err := s.client.HeaderByNumber(ctx, nil)
				if err != nil {
					return receipt, nil
				}
				if new(big.Int).Sub(header.Number, currentBlock).Uint64() >= confirmations {
					return receipt, nil
				}
				time.Sleep(5 * time.Second)
			}
		}
	}

	return receipt, nil
}
