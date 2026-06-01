package store

import (
	"log"
	"sync"
	"time"
)

type TransferEvent struct {
	BlockNumber uint64    `json:"blockNumber"`
	TxHash      string    `json:"txHash"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Value       string    `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

type EventStore struct {
	mu     sync.RWMutex
	events []TransferEvent
	limit  int
}

func NewEventStore(limit int) *EventStore {
	log.Printf("📦 [EventStore] 初始化，事件存储容量: %d", limit)
	return &EventStore{
		events: make([]TransferEvent, 0, limit),
		limit:  limit,
	}
}

func (s *EventStore) Add(event TransferEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	evicted := false
	if len(s.events) >= s.limit {
		evicted = true
		s.events = s.events[1:]
	}
	s.events = append(s.events, event)
	
	if evicted {
		log.Printf("💾 [EventStore] 添加事件，已淘汰最旧事件，当前数量: %d/%d", len(s.events), s.limit)
	} else {
		log.Printf("💾 [EventStore] 添加事件，当前数量: %d/%d", len(s.events), s.limit)
	}
}

func (s *EventStore) List() []TransferEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Printf("📋 [EventStore] 查询事件列表，返回 %d 条记录", len(s.events))
	
	result := make([]TransferEvent, len(s.events))
	copy(result, s.events)
	return result
}
