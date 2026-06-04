package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/meu/go-ether/service"
)

type TxHandlers struct {
	manager *service.ContractManager
}

func NewTxHandlers(manager *service.ContractManager) *TxHandlers {
	return &TxHandlers{manager: manager}
}

func (h *TxHandlers) getTxSendService() *service.TxSendService {
	if h.manager == nil {
		return nil
	}
	return h.manager.GetTxSendService()
}

func (h *TxHandlers) SendTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.SendTxRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析交易请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/tx/send - 目标: %s, 金额: %s", req.To, req.Value)

	txSendService := h.getTxSendService()
	if txSendService == nil {
		RequireSigner(w, "ETH 交易发送")
		return
	}

	resp, err := txSendService.SendTransaction(r.Context(), req)
	if err != nil {
		log.Printf("❌ [API] 发送交易失败: %v", err)
		http.Error(w, "failed to send transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 交易发送成功: %s", resp.TxHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *TxHandlers) GetTxHistory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}
	}

	log.Printf("📥 [API] GET /api/tx/history - limit: %d, offset: %d", limit, offset)

	entries, err := h.manager.GetTxHistory().List(limit, offset)
	if err != nil {
		log.Printf("❌ [API] 查询交易历史失败: %v", err)
		http.Error(w, "failed to get transaction history", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 返回 %d 条交易记录", len(entries))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func (h *TxHandlers) GetTxByHash(w http.ResponseWriter, r *http.Request) {
	txHash := r.URL.Query().Get("hash")
	if txHash == "" {
		http.Error(w, "tx hash is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] GET /api/tx/detail - hash: %s", txHash)

	entry, err := h.manager.GetTxHistory().GetByHash(txHash)
	if err != nil {
		log.Printf("❌ [API] 查询交易详情失败: %v", err)
		http.Error(w, "failed to get transaction", http.StatusInternalServerError)
		return
	}

	if entry == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Printf("✅ [API] 找到交易: %s", txHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}
