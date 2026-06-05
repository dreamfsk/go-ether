package service

import (
	"context"
	"log"
	"math/big"
	"sync"

	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/store"
	"github.com/meu/go-ether/wallet"
)

// ContractServiceBundle 管理某个合约地址对应的所有服务实例（事件监听、ERC20 交互、交易发送）
// 将 Service 创建和生命周期从 ContractManager 中分离，ContractManager 专注合约存储与调度。
type ContractServiceBundle struct {
	mu sync.RWMutex

	multiClient *client.MultiClient
	signer      wallet.Signer
	txHistory   *store.TxHistoryStore
	network     config.NetworkType
	chainID     *big.Int

	eventService  *EventService
	erc20Service  *ERC20Service
	txSendService *TxSendService
	txService     *TxService

	ctx    context.Context
	cancel context.CancelFunc
}

func NewContractServiceBundle(
	multiClient *client.MultiClient,
	signer wallet.Signer,
	txHistory *store.TxHistoryStore,
	network config.NetworkType,
	chainID *big.Int,
) *ContractServiceBundle {
	return &ContractServiceBundle{
		multiClient: multiClient,
		signer:      signer,
		txHistory:   txHistory,
		network:     network,
		chainID:     chainID,
	}
}

// Init 为指定合约地址初始化所有服务（替换旧服务并停止旧监听）
func (b *ContractServiceBundle) Init(contractAddr string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 停止旧服务
	b.stopLocked()

	b.ctx, b.cancel = context.WithCancel(context.Background())

	var err error
	b.eventService, err = NewEventService(b.multiClient.WS(), b.txHistory, b.network, contractAddr)
	if err != nil {
		// 事件服务失败不阻断整体初始化（WS 可能不可用）
		log.Printf("⚠️  [ServiceBundle] 事件服务初始化失败（事件监听不可用）: %v", err)
		b.eventService = nil
	}

	if b.signer != nil {
		b.erc20Service, err = NewERC20Service(b.multiClient.RPC(), b.signer, b.network, b.chainID, contractAddr)
		if err != nil {
			b.cancel()
			return err
		}

		b.txSendService = NewTxSendService(b.multiClient.RPC(), b.signer, b.network, b.chainID, b.txHistory)
	}

	// TxService 不需要 signer，始终创建
	b.txService = NewTxService(b.multiClient.RPC())

	log.Printf("[ServiceBundle] 服务初始化完成，合约: %s", contractAddr)
	return nil
}

// StartListening 启动事件监听
func (b *ContractServiceBundle) StartListening() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.eventService != nil && b.ctx != nil {
		go b.eventService.StartListening(b.ctx)
	}
}

// Stop 停止所有服务
func (b *ContractServiceBundle) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopLocked()
}

func (b *ContractServiceBundle) stopLocked() {
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	if b.txSendService != nil {
		b.txSendService.Stop()
	}
}

// EventService 返回事件服务（可能为 nil）
func (b *ContractServiceBundle) EventService() *EventService {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.eventService
}

// ERC20Service 返回 ERC20 服务（无 signer 时为 nil）
func (b *ContractServiceBundle) ERC20Service() *ERC20Service {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.erc20Service
}

// TxSendService 返回交易发送服务（无 signer 时为 nil）
func (b *ContractServiceBundle) TxSendService() *TxSendService {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.txSendService
}

// TxService 返回交易查询服务
func (b *ContractServiceBundle) TxService() *TxService {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.txService
}

// CreateERC20ServiceFor 为指定地址创建临时 ERC20Service
func (b *ContractServiceBundle) CreateERC20ServiceFor(address string) (*ERC20Service, error) {
	return NewERC20Service(b.multiClient.RPC(), b.signer, b.network, b.chainID, address)
}
