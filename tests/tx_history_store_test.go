package tests

import (
	"testing"
	"time"

	"github.com/meu/go-ether/store"
)

func TestTxHistoryStore_AddAndGet(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	entry := store.TxHistoryEntry{
		TxHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		FromAddr:    "0x1111111111111111111111111111111111111111",
		ToAddr:      "0x2222222222222222222222222222222222222222",
		Value:       "1000000000000000000",
		GasLimit:    21000,
		GasPrice:    "20000000000",
		Nonce:       1,
		Data:        "",
		Status:      store.TxStatusPending,
		BlockNumber: 0,
		Network:     "local",
		TxType:      "eth_transfer",
		CreatedAt:   time.Now(),
	}

	err = txStore.Add(entry)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	retrieved, err := txStore.GetByHash(entry.TxHash)
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetByHash returned nil")
	}

	if retrieved.TxHash != entry.TxHash {
		t.Errorf("TxHash mismatch: got %s, want %s", retrieved.TxHash, entry.TxHash)
	}

	if retrieved.FromAddr != entry.FromAddr {
		t.Errorf("FromAddr mismatch: got %s, want %s", retrieved.FromAddr, entry.FromAddr)
	}

	if retrieved.ToAddr != entry.ToAddr {
		t.Errorf("ToAddr mismatch: got %s, want %s", retrieved.ToAddr, entry.ToAddr)
	}
}

func TestTxHistoryStore_List(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	for i := 0; i < 5; i++ {
		err := txStore.Add(store.TxHistoryEntry{
			TxHash:      "0x" + string(rune('a'+i)) + "234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			FromAddr:    "0x1111111111111111111111111111111111111111",
			ToAddr:      "0x2222222222222222222222222222222222222222",
			Value:       "1000000000000000000",
			GasLimit:    21000,
			GasPrice:    "20000000000",
			Nonce:       uint64(i),
			Data:        "",
			Status:      store.TxStatusPending,
			BlockNumber: 0,
			Network:     "local",
			TxType:      "eth_transfer",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	list, err := txStore.List(3, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("List returned %d items, want 3", len(list))
	}
}

func TestTxHistoryStore_UpdateStatus(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	err = txStore.Add(store.TxHistoryEntry{
		TxHash:      txHash,
		FromAddr:    "0x1111111111111111111111111111111111111111",
		ToAddr:      "0x2222222222222222222222222222222222222222",
		Value:       "1000000000000000000",
		Status:      store.TxStatusPending,
		BlockNumber: 0,
		Network:     "local",
		TxType:      "eth_transfer",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	err = txStore.UpdateStatus(txHash, store.TxStatusSuccess, 12345)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	retrieved, err := txStore.GetByHash(txHash)
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}

	if retrieved.Status != store.TxStatusSuccess {
		t.Errorf("Status mismatch: got %d, want %d", retrieved.Status, store.TxStatusSuccess)
	}

	if retrieved.BlockNumber != 12345 {
		t.Errorf("BlockNumber mismatch: got %d, want %d", retrieved.BlockNumber, 12345)
	}
}

func TestTxHistoryStore_Count(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	count, err := txStore.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Initial count should be 0, got %d", count)
	}

	err = txStore.Add(store.TxHistoryEntry{
		TxHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		FromAddr:    "0x1111111111111111111111111111111111111111",
		ToAddr:      "0x2222222222222222222222222222222222222222",
		Value:       "1000000000000000000",
		Status:      store.TxStatusPending,
		Network:     "local",
		TxType:      "eth_transfer",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, err = txStore.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Count should be 1, got %d", count)
	}
}

// TestTxHistoryStore_ListByType 测试按 tx_type 过滤查询
func TestTxHistoryStore_ListByType(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	for i := 0; i < 3; i++ {
		txStore.Add(store.TxHistoryEntry{
			TxHash:    "0xe" + string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000001",
			FromAddr:  "0x1111111111111111111111111111111111111111",
			ToAddr:    "0x2222222222222222222222222222222222222222",
			Value:     "100",
			Status:    store.TxStatusSuccess,
			Network:   "local",
			TxType:    "eth_transfer",
			CreatedAt: time.Now(),
		})
	}
	for i := 0; i < 2; i++ {
		txStore.Add(store.TxHistoryEntry{
			TxHash:    "0xc" + string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000001",
			FromAddr:  "0x1111111111111111111111111111111111111111",
			ToAddr:    "0x3333333333333333333333333333333333333333",
			Value:     "500",
			Status:    store.TxStatusSuccess,
			Network:   "local",
			TxType:    "erc20_transfer",
			CreatedAt: time.Now(),
		})
	}

	ethTxs, err := txStore.ListByType("eth_transfer", 10, 0)
	if err != nil {
		t.Fatalf("ListByType(eth_transfer) failed: %v", err)
	}
	if len(ethTxs) != 3 {
		t.Errorf("ListByType(eth_transfer) = %d entries, want 3", len(ethTxs))
	}

	erc20Txs, err := txStore.ListByType("erc20_transfer", 10, 0)
	if err != nil {
		t.Fatalf("ListByType(erc20_transfer) failed: %v", err)
	}
	if len(erc20Txs) != 2 {
		t.Errorf("ListByType(erc20_transfer) = %d entries, want 2", len(erc20Txs))
	}
}

// TestTxHistoryStore_ListByTypeAndAddress 测试按 tx_type + 地址过滤
func TestTxHistoryStore_ListByTypeAndAddress(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	entries := []store.TxHistoryEntry{
		{TxHash: "0xa000000000000000000000000000000000000000000000000000000000000001", FromAddr: "0xaaa0000000000000000000000000000000000000", ToAddr: "0xbbb0000000000000000000000000000000000000", Value: "100", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
		{TxHash: "0xa000000000000000000000000000000000000000000000000000000000000002", FromAddr: "0xaaa0000000000000000000000000000000000000", ToAddr: "0xccc0000000000000000000000000000000000000", Value: "200", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
		{TxHash: "0xa000000000000000000000000000000000000000000000000000000000000003", FromAddr: "0xbbb0000000000000000000000000000000000000", ToAddr: "0xccc0000000000000000000000000000000000000", Value: "300", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		if err := txStore.Add(e); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	results, err := txStore.ListByTypeAndAddress("erc20_transfer", "0xaaa0000000000000000000000000000000000000", 10, 0)
	if err != nil {
		t.Fatalf("ListByTypeAndAddress failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ListByTypeAndAddress for 0xaaa = %d entries, want 2", len(results))
	}

	results, err = txStore.ListByTypeAndAddress("erc20_transfer", "0xccc0000000000000000000000000000000000000", 10, 0)
	if err != nil {
		t.Fatalf("ListByTypeAndAddress failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ListByTypeAndAddress for 0xccc = %d entries, want 2", len(results))
	}

	results, err = txStore.ListByTypeAndAddress("erc20_transfer", "0xddd0000000000000000000000000000000000000", 10, 0)
	if err != nil {
		t.Fatalf("ListByTypeAndAddress failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("ListByTypeAndAddress for unknown address = %d entries, want 0", len(results))
	}
}

// TestTxHistoryStore_CountByType 测试按 tx_type 统计
func TestTxHistoryStore_CountByType(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	for i := 0; i < 4; i++ {
		txStore.Add(store.TxHistoryEntry{
			TxHash:    "0xe" + string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000001",
			FromAddr:  "0x1111111111111111111111111111111111111111",
			ToAddr:    "0x2222222222222222222222222222222222222222",
			Value:     "100",
			Status:    store.TxStatusPending,
			Network:   "local",
			TxType:    "eth_transfer",
			CreatedAt: time.Now(),
		})
	}
	for i := 0; i < 3; i++ {
		txStore.Add(store.TxHistoryEntry{
			TxHash:    "0xc" + string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000001",
			FromAddr:  "0x1111111111111111111111111111111111111111",
			ToAddr:    "0x3333333333333333333333333333333333333333",
			Value:     "500",
			Status:    store.TxStatusSuccess,
			Network:   "local",
			TxType:    "erc20_transfer",
			CreatedAt: time.Now(),
		})
	}

	ethCount, err := txStore.CountByType("eth_transfer")
	if err != nil {
		t.Fatalf("CountByType failed: %v", err)
	}
	if ethCount != 4 {
		t.Errorf("CountByType(eth_transfer) = %d, want 4", ethCount)
	}

	erc20Count, err := txStore.CountByType("erc20_transfer")
	if err != nil {
		t.Fatalf("CountByType failed: %v", err)
	}
	if erc20Count != 3 {
		t.Errorf("CountByType(erc20_transfer) = %d, want 3", erc20Count)
	}
}

// TestTxHistoryStore_CountByTypeAndAddress 测试按 tx_type + 地址统计
func TestTxHistoryStore_CountByTypeAndAddress(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	entries := []store.TxHistoryEntry{
		{TxHash: "0xb000000000000000000000000000000000000000000000000000000000000001", FromAddr: "0xaaa0000000000000000000000000000000000001", ToAddr: "0xbbb0000000000000000000000000000000000002", Value: "100", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
		{TxHash: "0xb000000000000000000000000000000000000000000000000000000000000002", FromAddr: "0xbbb0000000000000000000000000000000000002", ToAddr: "0xaaa0000000000000000000000000000000000001", Value: "200", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
		{TxHash: "0xb000000000000000000000000000000000000000000000000000000000000003", FromAddr: "0xbbb0000000000000000000000000000000000002", ToAddr: "0xccc0000000000000000000000000000000000003", Value: "300", Status: store.TxStatusSuccess, Network: "local", TxType: "erc20_transfer", CreatedAt: time.Now()},
	}
	for _, e := range entries {
		if err := txStore.Add(e); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	count, err := txStore.CountByTypeAndAddress("erc20_transfer", "0xaaa0000000000000000000000000000000000001")
	if err != nil {
		t.Fatalf("CountByTypeAndAddress failed: %v", err)
	}
	if count != 2 {
		t.Errorf("CountByTypeAndAddress for 0xaaa = %d, want 2", count)
	}
}

// TestTxHistoryStore_GetByHash_NotFound 测试查询不存在的hash
func TestTxHistoryStore_GetByHash_NotFound(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	result, err := txStore.GetByHash("0xnonexistent00000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	if result != nil {
		t.Errorf("GetByHash for nonexistent hash should return nil, got %v", result)
	}
}

// TestTxHistoryStore_UpsertOnConflict 测试 ON CONFLICT 更新已有记录
func TestTxHistoryStore_UpsertOnConflict(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	txHash := "0xdup0000000000000000000000000000000000000000000000000000000000000001"

	err = txStore.Add(store.TxHistoryEntry{
		TxHash:      txHash,
		FromAddr:    "0x1111111111111111111111111111111111111111",
		ToAddr:      "0x2222222222222222222222222222222222222222",
		Value:       "100",
		Status:      store.TxStatusPending,
		BlockNumber: 0,
		Network:     "local",
		TxType:      "eth_transfer",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	err = txStore.Add(store.TxHistoryEntry{
		TxHash:      txHash,
		FromAddr:    "0x1111111111111111111111111111111111111111",
		ToAddr:      "0x2222222222222222222222222222222222222222",
		Value:       "100",
		Status:      store.TxStatusSuccess,
		BlockNumber: 12345,
		Network:     "local",
		TxType:      "eth_transfer",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Add (upsert) failed: %v", err)
	}

	retrieved, err := txStore.GetByHash(txHash)
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetByHash returned nil after upsert")
	}
	if retrieved.Status != store.TxStatusSuccess {
		t.Errorf("Upsert Status = %d, want %d", retrieved.Status, store.TxStatusSuccess)
	}
	if retrieved.BlockNumber != 12345 {
		t.Errorf("Upsert BlockNumber = %d, want 12345", retrieved.BlockNumber)
	}
}

// TestTxHistoryStore_ListPagination 测试分页边界
func TestTxHistoryStore_ListPagination(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	for i := 0; i < 10; i++ {
		txStore.Add(store.TxHistoryEntry{
			TxHash:    "0x" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + "0000000000000000000000000000000000000000000000000000000000000001",
			FromAddr:  "0x1111111111111111111111111111111111111111",
			ToAddr:    "0x2222222222222222222222222222222222222222",
			Value:     "100",
			Status:    store.TxStatusPending,
			Network:   "local",
			TxType:    "eth_transfer",
			CreatedAt: time.Now(),
		})
	}

	page1, err := txStore.List(3, 0)
	if err != nil {
		t.Fatalf("List page1 failed: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("List(3, 0) = %d entries, want 3", len(page1))
	}

	page2, err := txStore.List(3, 3)
	if err != nil {
		t.Fatalf("List page2 failed: %v", err)
	}
	if len(page2) != 3 {
		t.Errorf("List(3, 3) = %d entries, want 3", len(page2))
	}

	page4, err := txStore.List(3, 9)
	if err != nil {
		t.Fatalf("List page4 failed: %v", err)
	}
	if len(page4) != 1 {
		t.Errorf("List(3, 9) = %d entries, want 1", len(page4))
	}
}
