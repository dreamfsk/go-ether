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

func TestTxHistoryStore_ListByType(t *testing.T) {
	txStore, err := store.NewTxHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer txStore.Close()

	txStore.Add(store.TxHistoryEntry{
		TxHash:    "0xerc20-1",
		FromAddr:  "0xaaa",
		ToAddr:    "0xbbb",
		Value:     "100",
		Status:    store.TxStatusSuccess,
		Network:   "local",
		TxType:    "erc20_transfer",
		CreatedAt: time.Now(),
	})

	txStore.Add(store.TxHistoryEntry{
		TxHash:    "0xeth-1",
		FromAddr:  "0xccc",
		ToAddr:    "0xddd",
		Value:     "200",
		Status:    store.TxStatusPending,
		Network:   "local",
		TxType:    "eth_transfer",
		CreatedAt: time.Now(),
	})

	list, err := txStore.ListByType("erc20_transfer", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 erc20 event, got %d", len(list))
	}
	if list[0].TxType != "erc20_transfer" {
		t.Errorf("expected tx_type erc20_transfer, got %s", list[0].TxType)
	}

	all, err := txStore.List(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 total, got %d", len(all))
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
