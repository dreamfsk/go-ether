package tests

import (
	"testing"
	"time"

	"github.com/meu/go-ether/store"
)

func TestContractStore_AddAndGet(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	entry := store.ContractEntry{
		Address:   "0x1234567890abcdef1234567890abcdef12345678",
		Name:      "TestToken",
		Symbol:    "TEST",
		Network:   "local",
		Deployer:  "0x1111111111111111111111111111111111111111",
		TxHash:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	err = cs.Add(entry)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	retrieved, err := cs.GetByAddress(entry.Address)
	if err != nil {
		t.Fatalf("GetByAddress failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetByAddress returned nil")
	}
	if retrieved.Address != entry.Address {
		t.Errorf("Address mismatch: got %s, want %s", retrieved.Address, entry.Address)
	}
	if retrieved.Name != entry.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, entry.Name)
	}
	if retrieved.Symbol != entry.Symbol {
		t.Errorf("Symbol mismatch: got %s, want %s", retrieved.Symbol, entry.Symbol)
	}
}

func TestContractStore_GetActive(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	// 激活条目应该不存在
	active, err := cs.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active != nil {
		t.Error("GetActive should return nil when no active contract")
	}

	// 添加两个合约，一个活跃
	cs.Add(store.ContractEntry{
		Address:   "0xaaa0000000000000000000000000000000000001",
		Name:      "TokenA",
		Network:   "local",
		IsActive:  true,
		CreatedAt: time.Now(),
	})
	cs.Add(store.ContractEntry{
		Address:   "0xbbb0000000000000000000000000000000000002",
		Name:      "TokenB",
		Network:   "local",
		IsActive:  false,
		CreatedAt: time.Now(),
	})

	active, err = cs.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active == nil {
		t.Fatal("GetActive returned nil")
	}
	if active.Address != "0xaaa0000000000000000000000000000000000001" {
		t.Errorf("Active address = %s, want 0xaaa...0001", active.Address)
	}
}

func TestContractStore_SetActive(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	cs.Add(store.ContractEntry{
		Address:   "0xaaa0000000000000000000000000000000000001",
		Name:      "TokenA",
		Network:   "local",
		IsActive:  true,
		CreatedAt: time.Now(),
	})
	cs.Add(store.ContractEntry{
		Address:   "0xbbb0000000000000000000000000000000000002",
		Name:      "TokenB",
		Network:   "local",
		IsActive:  false,
		CreatedAt: time.Now(),
	})

	// 切换活跃合约
	err = cs.SetActive("0xbbb0000000000000000000000000000000000002")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	active, err := cs.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active == nil {
		t.Fatal("GetActive returned nil")
	}
	if active.Address != "0xbbb0000000000000000000000000000000000002" {
		t.Errorf("After SetActive, active = %s, want 0xbbb...0002", active.Address)
	}
}

func TestContractStore_SetActive_NotFound(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	err = cs.SetActive("0xnonexistent0000000000000000000000000000000000")
	if err == nil {
		t.Error("SetActive with nonexistent address should return error")
	}
}

func TestContractStore_List(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	list, err := cs.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("Initial List = %d entries, want 0", len(list))
	}

	cs.Add(store.ContractEntry{Address: "0xaaa0000000000000000000000000000000000001", Name: "A", Network: "local", CreatedAt: time.Now()})
	cs.Add(store.ContractEntry{Address: "0xbbb0000000000000000000000000000000000002", Name: "B", Network: "local", CreatedAt: time.Now()})

	list, err = cs.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List after add = %d entries, want 2", len(list))
	}
}

func TestContractStore_Upsert(t *testing.T) {
	cs, err := store.NewContractStore(":memory:")
	if err != nil {
		t.Fatalf("NewContractStore failed: %v", err)
	}
	defer cs.Close()

	addr := "0x1234000000000000000000000000000000000000"

	// 首次插入
	cs.Add(store.ContractEntry{
		Address:  addr,
		Name:     "OldName",
		Symbol:   "OLD",
		Network:  "local",
		IsActive: false,
	})

	// 同地址重复插入，更新 name/symbol
	cs.Add(store.ContractEntry{
		Address:  addr,
		Name:     "NewName",
		Symbol:   "NEW",
		Network:  "local",
		IsActive: true,
	})

	retrieved, err := cs.GetByAddress(addr)
	if err != nil {
		t.Fatalf("GetByAddress failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetByAddress returned nil after upsert")
	}
	if retrieved.Name != "NewName" {
		t.Errorf("Upsert Name = %s, want NewName", retrieved.Name)
	}
	if retrieved.Symbol != "NEW" {
		t.Errorf("Upsert Symbol = %s, want NEW", retrieved.Symbol)
	}
}
