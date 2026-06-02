package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/meu/go-ether/service"
)

type ContractHandlers struct {
	contractService *service.ContractService
	tokenService    *service.TokenService
}

func NewContractHandlers(contractService *service.ContractService, tokenService *service.TokenService) *ContractHandlers {
	return &ContractHandlers{
		contractService: contractService,
		tokenService:    tokenService,
	}
}

func (h *ContractHandlers) ContractView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.ContractCallRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析合约调用请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/contract/view - 方法: %s, 合约: %s", req.Method, req.ContractAddr)

	if h.contractService == nil {
		http.Error(w, "Contract service not available", http.StatusServiceUnavailable)
		return
	}

	resp, err := h.contractService.CallViewMethod(r.Context(), req)
	if err != nil {
		log.Printf("❌ [API] 合约视图调用失败: %v", err)
		http.Error(w, "Failed to call contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 合约视图调用成功: %s", resp.Result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ContractHandlers) ContractCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.ContractCallRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析合约调用请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/contract/call - 方法: %s, 合约: %s", req.Method, req.ContractAddr)

	if h.contractService == nil {
		http.Error(w, "Contract service not available", http.StatusServiceUnavailable)
		return
	}

	resp, err := h.contractService.SendTransaction(r.Context(), req)
	if err != nil {
		log.Printf("❌ [API] 合约交易调用失败: %v", err)
		http.Error(w, "Failed to send transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 合约交易发送成功: %s", resp.TxHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ContractHandlers) TokenInfo(w http.ResponseWriter, r *http.Request) {
	tokenAddr := r.URL.Query().Get("token")
	if tokenAddr == "" {
		http.Error(w, "token address is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] GET /api/token/info - token: %s", tokenAddr)

	if h.tokenService == nil {
		http.Error(w, "Token service not available", http.StatusServiceUnavailable)
		return
	}

	info, err := h.tokenService.GetTokenInfo(r.Context(), tokenAddr)
	if err != nil {
		log.Printf("❌ [API] 查询代币信息失败: %v", err)
		http.Error(w, "Failed to get token info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 查询代币信息成功: %s (%s)", info.Name, info.Symbol)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

func (h *ContractHandlers) TokenBalance(w http.ResponseWriter, r *http.Request) {
	tokenAddr := r.URL.Query().Get("token")
	holderAddr := r.URL.Query().Get("holder")

	if tokenAddr == "" {
		http.Error(w, "token address is required", http.StatusBadRequest)
		return
	}

	if holderAddr == "" {
		http.Error(w, "holder address is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] GET /api/token/balance - token: %s, holder: %s", tokenAddr, holderAddr)

	if h.tokenService == nil {
		http.Error(w, "Token service not available", http.StatusServiceUnavailable)
		return
	}

	balance, err := h.tokenService.GetBalance(r.Context(), tokenAddr, holderAddr)
	if err != nil {
		log.Printf("❌ [API] 查询代币余额失败: %v", err)
		http.Error(w, "Failed to get balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token":   tokenAddr,
		"holder":  holderAddr,
		"balance": balance,
	})
}

func (h *ContractHandlers) TokenTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.TokenTransferRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析代币转账请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/token/transfer - token: %s, to: %s, amount: %s", req.TokenAddr, req.To, req.Amount)

	if h.tokenService == nil {
		http.Error(w, "Token service not available", http.StatusServiceUnavailable)
		return
	}

	resp, err := h.tokenService.Transfer(r.Context(), req)
	if err != nil {
		log.Printf("❌ [API] 代币转账失败: %v", err)
		http.Error(w, "Failed to transfer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 代币转账交易发送成功: %s", resp.TxHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}