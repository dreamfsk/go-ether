package service

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/store"
	"github.com/meu/go-ether/wallet"
)

type ContractManager struct {
	mu sync.RWMutex

	client        *client.EthClient
	signer        wallet.Signer
	contractStore *store.ContractStore
	txHistory     *store.TxHistoryStore
	network       string
	chainID       *big.Int

	currentContract *store.ContractEntry
	eventService    *EventService
	erc20Service    *ERC20Service

	ctx    context.Context
	cancel context.CancelFunc
}

func NewContractManager(
	ethClient *client.EthClient,
	signer wallet.Signer,
	contractStore *store.ContractStore,
	txHistory *store.TxHistoryStore,
	network string,
	chainID *big.Int,
) *ContractManager {
	log.Println("🔧 [ContractManager] 初始化")
	return &ContractManager{
		client:        ethClient,
		signer:        signer,
		contractStore: contractStore,
		txHistory:     txHistory,
		network:       network,
		chainID:       chainID,
	}
}

func (m *ContractManager) Initialize(defaultAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	activeContract, err := m.contractStore.GetActive()
	if err != nil {
		return err
	}

	if activeContract != nil {
		log.Printf("📋 [ContractManager] 使用数据库中的活跃合约: %s", activeContract.Address)
		m.currentContract = activeContract
	} else if defaultAddress != "" {
		log.Printf("📋 [ContractManager] 使用默认合约地址: %s", defaultAddress)
		m.currentContract = &store.ContractEntry{
			Address:  defaultAddress,
			Network:  m.network,
			IsActive: true,
		}
	} else {
		return nil
	}

	return m.initServices()
}

func (m *ContractManager) initServices() error {
	if m.currentContract == nil {
		return nil
	}

	if m.cancel != nil {
		m.cancel()
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	var err error
	m.eventService, err = NewEventService(m.client, m.txHistory, config.NetworkType(m.network), m.currentContract.Address)
	if err != nil {
		return err
	}

	if m.signer != nil {
		m.erc20Service, err = NewERC20Service(m.client, m.signer, config.NetworkType(m.network), m.chainID, m.currentContract.Address)
		if err != nil {
			return err
		}
	}

	log.Printf("✅ [ContractManager] 服务初始化完成，合约: %s", m.currentContract.Address)
	return nil
}

func (m *ContractManager) SwitchContract(address string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentContract != nil && m.currentContract.Address == address {
		log.Printf("⚠️  [ContractManager] 合约地址未变化: %s", address)
		return nil
	}

	entry, err := m.contractStore.GetByAddress(address)
	if err != nil {
		return err
	}

	if entry == nil {
		entry = &store.ContractEntry{
			Address:  address,
			Network:  m.network,
			IsActive: true,
		}
		if err := m.contractStore.Add(*entry); err != nil {
			return err
		}
	}

	if err := m.contractStore.SetActive(address); err != nil {
		return err
	}

	m.currentContract = entry
	log.Printf("🔄 [ContractManager] 切换合约: %s", address)

	return m.initServices()
}

func (m *ContractManager) GetCurrentContract() *store.ContractEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentContract
}

func (m *ContractManager) GetEventService() *EventService {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.eventService
}

func (m *ContractManager) GetERC20Service() *ERC20Service {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.erc20Service
}

func (m *ContractManager) StartListening() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.eventService != nil && m.ctx != nil {
		log.Println("👂 [ContractManager] 启动事件监听...")
		go m.eventService.StartListening(m.ctx)
	}
}

func (m *ContractManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	log.Println("🛑 [ContractManager] 已停止")
}

func (m *ContractManager) AddDeployedContract(address, name, symbol, deployer, txHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := store.ContractEntry{
		Address:   address,
		Name:      name,
		Symbol:    symbol,
		Network:   m.network,
		Deployer:  deployer,
		TxHash:    txHash,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := m.contractStore.Add(entry); err != nil {
		return err
	}

	if err := m.contractStore.SetActive(address); err != nil {
		return err
	}

	m.currentContract = &entry
	log.Printf("✅ [ContractManager] 已添加部署的合约: %s", address)

	return m.initServices()
}

func (m *ContractManager) ListContracts() ([]store.ContractEntry, error) {
	return m.contractStore.List()
}
