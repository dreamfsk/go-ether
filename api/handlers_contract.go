package api

import (
	"encoding/json"
	"log"
	"math/big"
	"net/http"

	"github.com/meu/go-ether/service"
)

type ContractHandlers struct {
	contractService *service.ContractService
	erc20Service    *service.ERC20Service
}

func NewContractHandlers(contractService *service.ContractService, erc20Service *service.ERC20Service) *ContractHandlers {
	return &ContractHandlers{
		contractService: contractService,
		erc20Service:    erc20Service,
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
		RequireSigner(w, "合约交易调用")
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
	if h.erc20Service == nil {
		http.Error(w, "ERC20 service not available", http.StatusServiceUnavailable)
		return
	}

	info, err := h.erc20Service.GetTokenInfo(r.Context())
	if err != nil {
		log.Printf("❌ [API] 查询代币信息失败: %v", err)
		http.Error(w, "Failed to get token info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 查询代币信息成功: %s (%s)", info.Name, info.Symbol)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":         info.Name,
		"symbol":       info.Symbol,
		"decimals":     info.Decimals,
		"totalSupply":  info.TotalSupply.String(),
		"contractAddr": h.erc20Service.ContractAddress().Hex(),
	})
}

func (h *ContractHandlers) TokenBalance(w http.ResponseWriter, r *http.Request) {
	holderAddr := r.URL.Query().Get("holder")

	if holderAddr == "" {
		http.Error(w, "holder address is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] GET /api/token/balance - holder: %s", holderAddr)

	if h.erc20Service == nil {
		http.Error(w, "ERC20 service not available", http.StatusServiceUnavailable)
		return
	}

	balance, err := h.erc20Service.BalanceOf(r.Context(), holderAddr)
	if err != nil {
		log.Printf("❌ [API] 查询代币余额失败: %v", err)
		http.Error(w, "Failed to get balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"holder":        holderAddr,
		"balance":       balance.String(),
		"contractAddr":  h.erc20Service.ContractAddress().Hex(),
	})
}

func (h *ContractHandlers) TokenTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To     string `json:"to"`
		Amount string `json:"amount"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析代币转账请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/token/transfer - to: %s, amount: %s", req.To, req.Amount)

	if h.erc20Service == nil {
		RequireSigner(w, "代币转账")
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "invalid amount format", http.StatusBadRequest)
		return
	}

	txHash, err := h.erc20Service.Transfer(r.Context(), req.To, amount)
	if err != nil {
		log.Printf("❌ [API] 代币转账失败: %v", err)
		http.Error(w, "Failed to transfer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 代币转账交易发送成功: %s", txHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"txHash": txHash,
		"status": "pending",
	})
}

func (h *ContractHandlers) TokenMint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To     string `json:"to"`
		Amount string `json:"amount"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析代币铸造请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/token/mint - to: %s, amount: %s", req.To, req.Amount)

	if h.erc20Service == nil {
		RequireSigner(w, "代币铸造")
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "invalid amount format", http.StatusBadRequest)
		return
	}

	txHash, err := h.erc20Service.Mint(r.Context(), req.To, amount)
	if err != nil {
		log.Printf("❌ [API] 代币铸造失败: %v", err)
		http.Error(w, "Failed to mint: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 代币铸造交易发送成功: %s", txHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"txHash": txHash,
		"status": "pending",
	})
}

func (h *ContractHandlers) TokenDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name          string `json:"name"`
		Symbol        string `json:"symbol"`
		InitialSupply string `json:"initialSupply"`
		Recipient     string `json:"recipient"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("❌ [API] 解析合约部署请求失败: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Symbol == "" || req.InitialSupply == "" || req.Recipient == "" {
		http.Error(w, "name, symbol, initialSupply and recipient are required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/token/deploy - name: %s, symbol: %s, recipient: %s", req.Name, req.Symbol, req.Recipient)

	if h.erc20Service == nil {
		RequireSigner(w, "合约部署")
		return
	}

	initialSupply, ok := new(big.Int).SetString(req.InitialSupply, 10)
	if !ok {
		http.Error(w, "invalid initialSupply format", http.StatusBadRequest)
		return
	}

	result, err := h.erc20Service.Deploy(r.Context(), req.Name, req.Symbol, initialSupply, req.Recipient)
	if err != nil {
		log.Printf("❌ [API] 合约部署失败: %v", err)
		http.Error(w, "Failed to deploy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [API] 合约部署交易发送成功: tx=%s, addr=%s", result.TxHash, result.Address)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
