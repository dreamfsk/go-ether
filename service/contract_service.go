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

var (
	selectorCache   = make(map[string][]byte)
	selectorCacheMu sync.RWMutex
)

func getMethodSelector(methodSig string) ([]byte, error) {
	selectorCacheMu.RLock()
	if selector, ok := selectorCache[methodSig]; ok {
		selectorCacheMu.RUnlock()
		return selector, nil
	}
	selectorCacheMu.RUnlock()

	selectorCacheMu.Lock()
	defer selectorCacheMu.Unlock()

	if selector, ok := selectorCache[methodSig]; ok {
		return selector, nil
	}

	hash := crypto.Keccak256Hash([]byte(methodSig))
	selector := hash[:4]
	selectorCache[methodSig] = selector

	return selector, nil
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
