package api

import (
	"encoding/json"
	"net/http"
	"strings"

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
	id := strings.TrimPrefix(r.URL.Path, "/api/block/")
	if id == "" {
		http.Error(w, "block ID required", http.StatusBadRequest)
		return
	}

	block, err := h.blockService.GetBlockByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get block: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimPrefix(r.URL.Path, "/api/tx/")
	if hash == "" {
		http.Error(w, "transaction hash required", http.StatusBadRequest)
		return
	}

	tx, err := h.txService.GetTransactionByHash(r.Context(), hash)
	if err != nil {
		http.Error(w, "failed to get transaction: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	events := h.eventStore.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
