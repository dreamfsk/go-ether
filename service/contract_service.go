package service

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/wallet"
)

// 预计算的 ERC20 方法 selector，由 sync.Once 保证线程安全的一次性初始化。
var (
	selectorTransfer     []byte
	selectorMint         []byte
	selectorApprove      []byte
	selectorTransferFrom []byte
	selectorOnce         sync.Once
)

func initSelectors() {
	hash := crypto.Keccak256Hash([]byte("transfer(address,uint256)"))
	selectorTransfer = hash[:4]
	hash = crypto.Keccak256Hash([]byte("mint(address,uint256)"))
	selectorMint = hash[:4]
	hash = crypto.Keccak256Hash([]byte("approve(address,uint256)"))
	selectorApprove = hash[:4]
	hash = crypto.Keccak256Hash([]byte("transferFrom(address,address,uint256)"))
	selectorTransferFrom = hash[:4]
}

// getMethodSelector 根据方法签名返回 4 字节 selector（仅支持已注册的 ERC20 方法，未知签名则动态计算）
func getMethodSelector(methodSig string) ([]byte, error) {
	selectorOnce.Do(initSelectors)
	switch methodSig {
	case "transfer(address,uint256)":
		return selectorTransfer, nil
	case "mint(address,uint256)":
		return selectorMint, nil
	case "approve(address,uint256)":
		return selectorApprove, nil
	case "transferFrom(address,address,uint256)":
		return selectorTransferFrom, nil
	default:
		hash := crypto.Keccak256Hash([]byte(methodSig))
		return hash[:4], nil
	}
}

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

// CallViewMethod 调用合约视图方法（只读）。
// 注意：ERC20 方法已由 ERC20Service 在 API 层处理，此方法仅作为回退用于非 ERC20 合约。
func (s *ContractService) CallViewMethod(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	return nil, fmt.Errorf("unsupported view method: %s. ERC20 methods should be handled by ERC20Service", req.Method)
}

// SendTransaction 发送合约交易（写方法）。
// 注意：ERC20 方法已由 ERC20Service 在 API 层处理，此方法仅作为回退用于非 ERC20 合约。
func (s *ContractService) SendTransaction(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	return nil, fmt.Errorf("unsupported write method: %s. ERC20 methods should be handled by ERC20Service", req.Method)
}

// getCachedSelector 获取缓存的 selector（被 erc20_service.go 引用）
func getCachedSelector(methodSig string) []byte {
	selector, _ := getMethodSelector(methodSig)
	return selector
}
