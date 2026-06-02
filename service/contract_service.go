package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/wallet"
)

var (
	selectorCache   = make(map[string][]byte)
	selectorCacheMu sync.RWMutex
)

func getMethodSelector(methodSig string) ([]byte, error) {
	selectorCacheMu.RLock()
	if selector, ok := selectorCache[methodSig]; ok {
		selectorCacheMu.RUnlock()
		return selector, nil
	}
	selectorCacheMu.RUnlock()

	selectorCacheMu.Lock()
	defer selectorCacheMu.Unlock()

	if selector, ok := selectorCache[methodSig]; ok {
		return selector, nil
	}

	hash := crypto.Keccak256Hash([]byte(methodSig))
	selector := hash[:4]
	selectorCache[methodSig] = selector

	return selector, nil
}

type ContractService struct {
	client  *client.EthClient
	signer  wallet.Signer
	network config.NetworkType
	chainID *big.Int
}

func NewContractService(c *client.EthClient, signer wallet.Signer, network config.NetworkType, chainID *big.Int) *ContractService {
	return &ContractService{
		client:  c,
		signer:  signer,
		network: network,
		chainID: chainID,
	}
}

type ContractCallRequest struct {
	ContractAddr string   `json:"contractAddr"`
	Method       string   `json:"method"`
	Args         []string `json:"args,omitempty"`
}

type ContractCallResponse struct {
	TxHash   string `json:"txHash,omitempty"`
	Result   string `json:"result,omitempty"`
	Status   string `json:"status"`
	GasUsed  uint64 `json:"gasUsed,omitempty"`
}

// CallViewMethod 调用合约视图方法（只读）
func (s *ContractService) CallViewMethod(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	log.Printf("🔍 [ContractService] 调用视图方法: %s, 合约: %s", req.Method, req.ContractAddr)

	contractAddr := common.HexToAddress(req.ContractAddr)

	switch req.Method {
	// ERC20 视图方法
	case "name":
		return s.callERC20View(ctx, contractAddr, "name()", "string")
	case "symbol":
		return s.callERC20View(ctx, contractAddr, "symbol()", "string")
	case "decimals":
		return s.callERC20View(ctx, contractAddr, "decimals()", "uint8")
	case "totalSupply":
		return s.callERC20View(ctx, contractAddr, "totalSupply()", "uint256")
	case "balanceOf":
		if len(req.Args) < 1 {
			return nil, fmt.Errorf("balanceOf requires address argument")
		}
		return s.callERC20ViewWithArg(ctx, contractAddr, "balanceOf(address)", req.Args[0])
	case "allowance":
		if len(req.Args) < 2 {
			return nil, fmt.Errorf("allowance requires owner and spender arguments")
		}
		return s.callERC20ViewWithArgs(ctx, contractAddr, "allowance(address,address)", req.Args[0], req.Args[1])
	default:
		return nil, fmt.Errorf("unsupported view method: %s", req.Method)
	}
}

// SendTransaction 发送合约交易（写方法）
func (s *ContractService) SendTransaction(ctx context.Context, req ContractCallRequest) (*ContractCallResponse, error) {
	if s.signer == nil {
		return nil, fmt.Errorf("signer not initialized")
	}

	log.Printf("✉️ [ContractService] 发送交易调用: %s, 合约: %s", req.Method, req.ContractAddr)

	contractAddr := common.HexToAddress(req.ContractAddr)

	var data []byte
	var err error

	switch req.Method {
	case "transfer":
		if len(req.Args) < 2 {
			return nil, fmt.Errorf("transfer requires to and amount arguments")
		}
		data, err = abiEncodeTransfer(req.Args[0], req.Args[1])
	case "mint":
		if len(req.Args) < 2 {
			return nil, fmt.Errorf("mint requires to and amount arguments")
		}
		data, err = abiEncodeMint(req.Args[0], req.Args[1])
	case "approve":
		if len(req.Args) < 2 {
			return nil, fmt.Errorf("approve requires spender and amount arguments")
		}
		data, err = abiEncodeApprove(req.Args[0], req.Args[1])
	case "transferFrom":
		if len(req.Args) < 3 {
			return nil, fmt.Errorf("transferFrom requires from, to and amount arguments")
		}
		data, err = abiEncodeTransferFrom(req.Args[0], req.Args[1], req.Args[2])
	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	return s.sendRawTransaction(ctx, &contractAddr, data)
}

// callERC20View 调用返回单个值的 ERC20 视图方法
func (s *ContractService) callERC20View(ctx context.Context, contractAddr common.Address, methodSig, returnType string) (*ContractCallResponse, error) {
	data, err := getMethodSelector(methodSig)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	var value string
	switch returnType {
	case "string":
		value = parseStringResult(result)
	case "uint8":
		if len(result) == 0 {
			value = "18"
		} else {
			value = fmt.Sprintf("%d", uint8(result[31]))
		}
	case "uint256":
		v := new(big.Int).SetBytes(result)
		value = v.String()
	}

	return &ContractCallResponse{
		Result: value,
		Status: "success",
	}, nil
}

// callERC20ViewWithArg 调用带一个地址参数的 ERC20 视图方法
func (s *ContractService) callERC20ViewWithArg(ctx context.Context, contractAddr common.Address, methodSig, arg string) (*ContractCallResponse, error) {
	addr := common.HexToAddress(arg)
	data := append(getCachedSelector(methodSig), common.LeftPadBytes(addr.Bytes(), 32)...)

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	v := new(big.Int).SetBytes(result)

	return &ContractCallResponse{
		Result: v.String(),
		Status: "success",
	}, nil
}

// callERC20ViewWithArgs 调用带两个地址参数的 ERC20 视图方法
func (s *ContractService) callERC20ViewWithArgs(ctx context.Context, contractAddr common.Address, methodSig, arg1, arg2 string) (*ContractCallResponse, error) {
	addr1 := common.HexToAddress(arg1)
	addr2 := common.HexToAddress(arg2)

	data := getCachedSelector(methodSig)
	data = append(data, common.LeftPadBytes(addr1.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(addr2.Bytes(), 32)...)

	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	v := new(big.Int).SetBytes(result)

	return &ContractCallResponse{
		Result: v.String(),
		Status: "success",
	}, nil
}

// sendRawTransaction 发送原始交易
func (s *ContractService) sendRawTransaction(ctx context.Context, contractAddr *common.Address, data []byte) (*ContractCallResponse, error) {
	fromAddr := s.signer.Address()

	nonce, err := s.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasTipCap, err := s.client.SuggestGasTipCap(ctx)
	if err != nil {
		gasTipCap = big.NewInt(1_000_000_000)
	}

	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %w", err)
	}

	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(1_000_000_000)
	}

	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   s.chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       100000,
		To:        contractAddr,
		Value:     big.NewInt(0),
		Data:      data,
	})

	signedTx, err := s.signer.SignTx(ctx, tx, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	log.Printf("✅ [ContractService] 合约交易发送成功: %s", txHash)

	return &ContractCallResponse{
		TxHash: txHash,
		Status: "pending",
	}, nil
}

// getCachedSelector 获取缓存的 selector
func getCachedSelector(methodSig string) []byte {
	selector, _ := getMethodSelector(methodSig)
	return selector
}

// ABI 编码函数

func abiEncodeTransfer(to, amount string) ([]byte, error) {
	methodID, err := getMethodSelector("transfer(address,uint256)")
	if err != nil {
		return nil, err
	}

	toAddr := common.HexToAddress(to)
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	encoded, err := abi.Arguments{{Type: addressType}, {Type: uint256Type}}.Pack(toAddr, amountBig)
	if err != nil {
		return nil, err
	}

	return append(methodID, encoded...), nil
}

func abiEncodeMint(to, amount string) ([]byte, error) {
	methodID, err := getMethodSelector("mint(address,uint256)")
	if err != nil {
		return nil, err
	}

	toAddr := common.HexToAddress(to)
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	encoded, err := abi.Arguments{{Type: addressType}, {Type: uint256Type}}.Pack(toAddr, amountBig)
	if err != nil {
		return nil, err
	}

	return append(methodID, encoded...), nil
}

func abiEncodeApprove(spender, amount string) ([]byte, error) {
	methodID, err := getMethodSelector("approve(address,uint256)")
	if err != nil {
		return nil, err
	}

	spenderAddr := common.HexToAddress(spender)
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	encoded, err := abi.Arguments{{Type: addressType}, {Type: uint256Type}}.Pack(spenderAddr, amountBig)
	if err != nil {
		return nil, err
	}

	return append(methodID, encoded...), nil
}

func abiEncodeTransferFrom(from, to, amount string) ([]byte, error) {
	methodID, err := getMethodSelector("transferFrom(address,address,uint256)")
	if err != nil {
		return nil, err
	}

	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	encoded, err := abi.Arguments{{Type: addressType}, {Type: addressType}, {Type: uint256Type}}.Pack(fromAddr, toAddr, amountBig)
	if err != nil {
		return nil, err
	}

	return append(methodID, encoded...), nil
}

// parseStringResult 解析 ABI 编码的 string 返回值
func parseStringResult(data []byte) string {
	if len(data) < 64 {
		return ""
	}

	length := new(big.Int).SetBytes(data[32:64]).Uint64()
	if length == 0 {
		return ""
	}

	if int(length)+64 > len(data) {
		return ""
	}

	return string(data[64 : 64+length])
}
