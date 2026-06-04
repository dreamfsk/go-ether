package service

import (
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/store"
	"github.com/meu/go-ether/wallet"
)

// ContractManager 管理合约注册、切换、持久化，内部委托 ServiceBundle 管理服务生命周期。
type ContractManager struct {
	mu sync.RWMutex

	multiClient   *client.MultiClient
	signer        wallet.Signer
	contractStore *store.ContractStore
	txHistory     *store.TxHistoryStore
	network       config.NetworkType
	chainID       *big.Int

	currentContract *store.ContractEntry
	serviceBundle   *ContractServiceBundle
}

func NewContractManager(
	multiClient *client.MultiClient,
	signer wallet.Signer,
	contractStore *store.ContractStore,
	txHistory *store.TxHistoryStore,
	network string,
	chainID *big.Int,
) *ContractManager {
	networkType := config.NetworkType(network)
	log.Println("🔧 [ContractManager] 初始化")

	cm := &ContractManager{
		multiClient:   multiClient,
		signer:        signer,
		contractStore: contractStore,
		txHistory:     txHistory,
		network:       networkType,
		chainID:       chainID,
	}
	cm.serviceBundle = NewContractServiceBundle(multiClient, signer, txHistory, networkType, chainID)
	return cm
}

// Initialize 加载活跃合约并初始化关联服务
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
			Network:  string(m.network),
			IsActive: true,
		}
	} else {
		return nil
	}

	return m.serviceBundle.Init(m.currentContract.Address)
}

// SwitchContract 切换到指定合约地址
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
			Network:  string(m.network),
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

	return m.serviceBundle.Init(address)
}

// AddDeployedContract 注册新部署的合约并切换为活跃合约
func (m *ContractManager) AddDeployedContract(address, name, symbol, deployer, txHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := store.ContractEntry{
		Address:   address,
		Name:      name,
		Symbol:    symbol,
		Network:   string(m.network),
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

	return m.serviceBundle.Init(address)
}

// --- 查询方法 ---

func (m *ContractManager) GetCurrentContract() *store.ContractEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentContract
}

func (m *ContractManager) GetEventService() *EventService {
	return m.serviceBundle.EventService()
}

func (m *ContractManager) GetERC20Service() *ERC20Service {
	return m.serviceBundle.ERC20Service()
}

func (m *ContractManager) GetTxSendService() *TxSendService {
	return m.serviceBundle.TxSendService()
}

func (m *ContractManager) GetTxHistory() *store.TxHistoryStore {
	return m.txHistory
}

func (m *ContractManager) ListContracts() ([]store.ContractEntry, error) {
	return m.contractStore.List()
}

// GetERC20ServiceFor 为指定地址创建临时 ERC20Service
func (m *ContractManager) GetERC20ServiceFor(address string) (*ERC20Service, error) {
	return m.serviceBundle.CreateERC20ServiceFor(address)
}

// --- 生命周期 ---

func (m *ContractManager) StartListening() {
	m.serviceBundle.StartListening()
}

func (m *ContractManager) Stop() {
	m.serviceBundle.Stop()
	log.Println("🛑 [ContractManager] 已停止")
}
