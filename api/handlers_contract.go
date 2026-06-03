package api

import (
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/meu/go-ether/service"
)

type ContractHandlers struct {
	contractManager *service.ContractManager
}

func NewContractHandlers(contractManager *service.ContractManager) *ContractHandlers {
	return &ContractHandlers{
		contractManager: contractManager,
	}
}

func (h *ContractHandlers) getERC20Service() *service.ERC20Service {
	if h.contractManager == nil {
		return nil
	}
	return h.contractManager.GetERC20Service()
}

func (h *ContractHandlers) ContractList(w http.ResponseWriter, r *http.Request) {
	if h.contractManager == nil {
		http.Error(w, "Contract manager not available", http.StatusServiceUnavailable)
		return
	}

	contracts, err := h.contractManager.ListContracts()
	if err != nil {
		log.Printf("❌ [API] 查询合约列表失败: %v", err)
		http.Error(w, "Failed to list contracts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	current := h.contractManager.GetCurrentContract()
	currentAddr := ""
	if current != nil {
		currentAddr = current.Address
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contracts":      contracts,
		"currentAddress": currentAddr,
	})
}

func (h *ContractHandlers) ContractSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.contractManager == nil {
		http.Error(w, "Contract manager not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Address string `json:"address"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] POST /api/contract/switch - address: %s", req.Address)

	if err := h.contractManager.SwitchContract(req.Address); err != nil {
		log.Printf("❌ [API] 切换合约失败: %v", err)
		http.Error(w, "Failed to switch contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.contractManager.StartListening()

	log.Printf("✅ [API] 合约切换成功: %s", req.Address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"address":  req.Address,
		"message":  "Contract switched successfully",
	})
}

func (h *ContractHandlers) ContractCurrent(w http.ResponseWriter, r *http.Request) {
	if h.contractManager == nil {
		http.Error(w, "Contract manager not available", http.StatusServiceUnavailable)
		return
	}

	current := h.contractManager.GetCurrentContract()
	if current == nil {
		http.Error(w, "No active contract", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(current)
}

// erc20ViewMethods ERC20 视图方法白名单
var erc20ViewMethods = map[string]bool{
	"name": true, "symbol": true, "decimals": true, "totalSupply": true,
	"balanceOf": true, "allowance": true,
}

// erc20WriteMethods ERC20 写方法白名单
var erc20WriteMethods = map[string]bool{
	"transfer": true, "mint": true, "approve": true, "transferFrom": true,
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

	if erc20ViewMethods[req.Method] && h.contractManager != nil {
		resp := h.handleERC20View(r, req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("❌ [API] 不支持的方法: %s", req.Method)
	http.Error(w, "unsupported view method: "+req.Method, http.StatusBadRequest)
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

	if erc20WriteMethods[req.Method] && h.contractManager != nil {
		resp := h.handleERC20Write(r, req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("❌ [API] 不支持的方法: %s", req.Method)
	http.Error(w, "unsupported write method: "+req.Method, http.StatusBadRequest)
}

// handleERC20View 委托 ERC20 视图方法到 ERC20Service
func (h *ContractHandlers) handleERC20View(r *http.Request, req service.ContractCallRequest) *service.ContractCallResponse {
	erc20, err := h.contractManager.GetERC20ServiceFor(req.ContractAddr)
	if err != nil {
		log.Printf("❌ [API] 创建 ERC20 服务失败: %v", err)
		return &service.ContractCallResponse{Status: "error", Result: err.Error()}
	}

	switch req.Method {
	case "name", "symbol", "decimals", "totalSupply":
		info, err := erc20.GetTokenInfo(r.Context())
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		switch req.Method {
		case "name":
			return &service.ContractCallResponse{Result: info.Name, Status: "success"}
		case "symbol":
			return &service.ContractCallResponse{Result: info.Symbol, Status: "success"}
		case "decimals":
			return &service.ContractCallResponse{Result: strconv.Itoa(int(info.Decimals)), Status: "success"}
		default:
			return &service.ContractCallResponse{Result: info.TotalSupply.String(), Status: "success"}
		}
	case "balanceOf":
		if len(req.Args) < 1 {
			return &service.ContractCallResponse{Status: "error", Result: "balanceOf requires address argument"}
		}
		balance, err := erc20.BalanceOf(r.Context(), req.Args[0])
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{Result: balance.String(), Status: "success"}
	case "allowance":
		if len(req.Args) < 2 {
			return &service.ContractCallResponse{Status: "error", Result: "allowance requires owner and spender arguments"}
		}
		allowance, err := erc20.Allowance(r.Context(), req.Args[0], req.Args[1])
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{Result: allowance.String(), Status: "success"}
	default:
		return &service.ContractCallResponse{Status: "error", Result: "unsupported method: " + req.Method}
	}
}

// handleERC20Write 委托 ERC20 写方法到 ERC20Service
func (h *ContractHandlers) handleERC20Write(r *http.Request, req service.ContractCallRequest) *service.ContractCallResponse {
	erc20, err := h.contractManager.GetERC20ServiceFor(req.ContractAddr)
	if err != nil {
		log.Printf("❌ [API] 创建 ERC20 服务失败: %v", err)
		return &service.ContractCallResponse{Status: "error", Result: err.Error()}
	}

	switch req.Method {
	case "transfer":
		if len(req.Args) < 2 {
			return &service.ContractCallResponse{Status: "error", Result: "transfer requires to and amount arguments"}
		}
		amount, ok := new(big.Int).SetString(req.Args[1], 10)
		if !ok {
			return &service.ContractCallResponse{Status: "error", Result: "invalid amount"}
		}
		txHash, err := erc20.Transfer(r.Context(), req.Args[0], amount)
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{TxHash: txHash, Status: "pending"}
	case "mint":
		if len(req.Args) < 2 {
			return &service.ContractCallResponse{Status: "error", Result: "mint requires to and amount arguments"}
		}
		amount, ok := new(big.Int).SetString(req.Args[1], 10)
		if !ok {
			return &service.ContractCallResponse{Status: "error", Result: "invalid amount"}
		}
		txHash, err := erc20.Mint(r.Context(), req.Args[0], amount)
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{TxHash: txHash, Status: "pending"}
	case "approve":
		if len(req.Args) < 2 {
			return &service.ContractCallResponse{Status: "error", Result: "approve requires spender and amount arguments"}
		}
		amount, ok := new(big.Int).SetString(req.Args[1], 10)
		if !ok {
			return &service.ContractCallResponse{Status: "error", Result: "invalid amount"}
		}
		txHash, err := erc20.Approve(r.Context(), req.Args[0], amount)
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{TxHash: txHash, Status: "pending"}
	case "transferFrom":
		if len(req.Args) < 3 {
			return &service.ContractCallResponse{Status: "error", Result: "transferFrom requires from, to and amount arguments"}
		}
		amount, ok := new(big.Int).SetString(req.Args[2], 10)
		if !ok {
			return &service.ContractCallResponse{Status: "error", Result: "invalid amount"}
		}
		txHash, err := erc20.TransferFrom(r.Context(), req.Args[0], req.Args[1], amount)
		if err != nil {
			return &service.ContractCallResponse{Status: "error", Result: err.Error()}
		}
		return &service.ContractCallResponse{TxHash: txHash, Status: "pending"}
	default:
		return &service.ContractCallResponse{Status: "error", Result: "unsupported method: " + req.Method}
	}
}

func (h *ContractHandlers) TokenInfo(w http.ResponseWriter, r *http.Request) {
	// 支持按指定合约地址查询（用于合约详情页）
	contractAddr := r.URL.Query().Get("address")

	var erc20 *service.ERC20Service
	if contractAddr != "" && h.contractManager != nil {
		var err error
		erc20, err = h.contractManager.GetERC20ServiceFor(contractAddr)
		if err != nil {
			log.Printf("❌ [API] 创建指定合约 ERC20Service 失败: %v", err)
			http.Error(w, "Failed to create ERC20 service for address: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		erc20 = h.getERC20Service()
	}

	if erc20 == nil {
		http.Error(w, "ERC20 service not available", http.StatusServiceUnavailable)
		return
	}

	info, err := erc20.GetTokenInfo(r.Context())
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
		"contractAddr": erc20.ContractAddress().Hex(),
	})
}

func (h *ContractHandlers) TokenBalance(w http.ResponseWriter, r *http.Request) {
	holderAddr := r.URL.Query().Get("holder")

	if holderAddr == "" {
		http.Error(w, "holder address is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 [API] GET /api/token/balance - holder: %s", holderAddr)

	if h.getERC20Service() == nil {
		http.Error(w, "ERC20 service not available", http.StatusServiceUnavailable)
		return
	}

	balance, err := h.getERC20Service().BalanceOf(r.Context(), holderAddr)
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
		"contractAddr":  h.getERC20Service().ContractAddress().Hex(),
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

	if h.getERC20Service() == nil {
		RequireSigner(w, "代币转账")
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "invalid amount format", http.StatusBadRequest)
		return
	}

	txHash, err := h.getERC20Service().Transfer(r.Context(), req.To, amount)
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

	if h.getERC20Service() == nil {
		RequireSigner(w, "代币铸造")
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "invalid amount format", http.StatusBadRequest)
		return
	}

	txHash, err := h.getERC20Service().Mint(r.Context(), req.To, amount)
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

	if h.getERC20Service() == nil {
		RequireSigner(w, "合约部署")
		return
	}

	initialSupply, ok := new(big.Int).SetString(req.InitialSupply, 10)
	if !ok {
		http.Error(w, "invalid initialSupply format", http.StatusBadRequest)
		return
	}

	result, err := h.getERC20Service().Deploy(r.Context(), req.Name, req.Symbol, initialSupply, req.Recipient)
	if err != nil {
		log.Printf("❌ [API] 合约部署失败: %v", err)
		http.Error(w, "Failed to deploy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if h.contractManager != nil {
		if err := h.contractManager.AddDeployedContract(result.Address, req.Name, req.Symbol, "", result.TxHash); err != nil {
			log.Printf("⚠️  [API] 保存合约地址失败: %v", err)
		}
		h.contractManager.StartListening()
	}

	log.Printf("✅ [API] 合约部署交易发送成功: tx=%s, addr=%s", result.TxHash, result.Address)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
