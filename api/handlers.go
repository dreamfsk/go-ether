package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/meu/go-ether/service"
	"github.com/meu/go-ether/store"
)

type Handlers struct {
	blockService *service.BlockService
	txService    *service.TxService
	eventStore   *store.EventStore
}

func NewHandlers(bs *service.BlockService, ts *service.TxService, es *store.EventStore) *Handlers {
	return &Handlers{
		blockService: bs,
		txService:    ts,
		eventStore:   es,
	}
}

func (h *Handlers) GetBlock(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id := strings.TrimPrefix(r.URL.Path, "/api/block/")
	log.Printf("📥 [API] GET /api/block/%s from %s", id, r.RemoteAddr)
	
	if id == "" {
		log.Printf("❌ [API] GET /api/block/: Bad request - 缺少区块 ID")
		http.Error(w, "block ID required", http.StatusBadRequest)
		return
	}

	block, err := h.blockService.GetBlockByID(r.Context(), id)
	if err != nil {
		log.Printf("❌ [API] GET /api/block/%s: 错误 - %v", id, err)
		http.Error(w, "failed to get block: "+err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("✅ [API] GET /api/block/%s: 成功获取 - 区块 #%d, 耗时 %v", id, block.Number, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	hash := strings.TrimPrefix(r.URL.Path, "/api/tx/")
	log.Printf("📥 [API] GET /api/tx/%s from %s", hash, r.RemoteAddr)
	
	if hash == "" {
		log.Printf("❌ [API] GET /api/tx/: Bad request - 缺少交易哈希")
		http.Error(w, "transaction hash required", http.StatusBadRequest)
		return
	}

	tx, err := h.txService.GetTransactionByHash(r.Context(), hash)
	if err != nil {
		log.Printf("❌ [API] GET /api/tx/%s: 错误 - %v", hash, err)
		http.Error(w, "failed to get transaction: "+err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("✅ [API] GET /api/tx/%s: 成功获取 - 从 %s 到 %s, 耗时 %v", hash, tx.From, tx.To, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("📥 [API] GET /api/events from %s", r.RemoteAddr)
	
	events := h.eventStore.List()
	
	log.Printf("✅ [API] GET /api/events: 成功获取 %d 条事件, 耗时 %v", len(events), time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
