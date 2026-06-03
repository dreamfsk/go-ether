package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meu/go-ether/service"
	"github.com/meu/go-ether/store"
)

type Handlers struct {
	blockService   *service.BlockService
	txService      *service.TxService
	txHistoryStore *store.TxHistoryStore
}

func NewHandlers(bs *service.BlockService, ts *service.TxService, ths *store.TxHistoryStore) *Handlers {
	return &Handlers{
		blockService:   bs,
		txService:      ts,
		txHistoryStore: ths,
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
	address := r.URL.Query().Get("address")
	
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	log.Printf("📥 [API] GET /api/events from %s, address: %s, limit: %d, offset: %d", r.RemoteAddr, address, limit, offset)
	
	var events []store.TxHistoryEntry
	var total int
	var errList, errCount error
	
	if address != "" {
		events, errList = h.txHistoryStore.ListByTypeAndAddress("erc20_transfer", strings.ToLower(address), limit, offset)
		total, errCount = h.txHistoryStore.CountByTypeAndAddress("erc20_transfer", strings.ToLower(address))
	} else {
		events, errList = h.txHistoryStore.ListByType("erc20_transfer", limit, offset)
		total, errCount = h.txHistoryStore.CountByType("erc20_transfer")
	}
	
	if errList != nil {
		log.Printf("❌ [API] GET /api/events: 查询事件失败 - %v", errList)
		http.Error(w, "failed to get events: "+errList.Error(), http.StatusInternalServerError)
		return
	}
	if errCount != nil {
		log.Printf("❌ [API] GET /api/events: 统计事件数失败 - %v", errCount)
		http.Error(w, "failed to count events: "+errCount.Error(), http.StatusInternalServerError)
		return
	}
	
	log.Printf("✅ [API] GET /api/events: 成功获取 %d/%d 条事件, 耗时 %v", len(events), total, time.Since(start))
	
	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
