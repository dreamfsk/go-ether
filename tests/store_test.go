package tests

import (
	"testing"
	"time"

	"github.com/meu/go-ether/store"
)

func TestEventStore_AddAndList(t *testing.T) {
	eventStore := store.NewEventStore(3)

	event1 := store.TransferEvent{
		BlockNumber: 1,
		TxHash:      "0x111",
		From:        "0xaaa",
		To:          "0xbbb",
		Value:       "100",
		Timestamp:   time.Now(),
	}
	event2 := store.TransferEvent{
		BlockNumber: 2,
		TxHash:      "0x222",
		From:        "0xccc",
		To:          "0xdd",
		Value:       "200",
		Timestamp:   time.Now(),
	}
	event3 := store.TransferEvent{
		BlockNumber: 3,
		TxHash:      "0x333",
		From:        "0xeee",
		To:          "0xfff",
		Value:       "300",
		Timestamp:   time.Now(),
	}
	event4 := store.TransferEvent{
		BlockNumber: 4,
		TxHash:      "0x444",
		From:        "0xggg",
		To:          "0xhhh",
		Value:       "400",
		Timestamp:   time.Now(),
	}

	// Add first event
	eventStore.Add(event1)
	list := eventStore.List()
	if len(list) != 1 {
		t.Errorf("List() after adding 1 event returned %d items, want 1", len(list))
	}

	// Add second event
	eventStore.Add(event2)
	list = eventStore.List()
	if len(list) != 2 {
		t.Errorf("List() after adding 2 events returned %d items, want 2", len(list))
	}

	// Add third event
	eventStore.Add(event3)
	list = eventStore.List()
	if len(list) != 3 {
		t.Errorf("List() after adding 3 events returned %d items, want 3", len(list))
	}

	// Add fourth event (should evict the first)
	eventStore.Add(event4)
	list = eventStore.List()
	if len(list) != 3 {
		t.Errorf("List() after adding 4th event returned %d items, want 3", len(list))
	}

	// Check that the first event was evicted
	if list[0].TxHash != "0x222" {
		t.Errorf("First event should be 0x222 after eviction, got %s", list[0].TxHash)
	}
}

func TestEventStore_Concurrency(t *testing.T) {
	eventStore := store.NewEventStore(100)

	// Run concurrent additions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				eventStore.Add(store.TransferEvent{
					BlockNumber: uint64(id*10 + j),
					TxHash:      "tx-" + string(rune('a'+id)) + "-" + string(rune('0'+j)),
					From:        "0xaaa",
					To:          "0xbbb",
					Value:       "100",
					Timestamp:   time.Now(),
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no data corruption
	list := eventStore.List()
	if len(list) != 100 {
		t.Errorf("List() after concurrent adds returned %d items, want 100", len(list))
	}
}