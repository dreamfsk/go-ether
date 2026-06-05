package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/store"
	"github.com/meu/go-ether/wallet"
)

type TxSendService struct {
	client      *client.EthClient
	signer      wallet.Signer
	network     config.NetworkType
	chainID     *big.Int
	txHistory   *store.TxHistoryStore

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTxSendService(c *client.EthClient, signer wallet.Signer, network config.NetworkType, chainID *big.Int, txHistory *store.TxHistoryStore) *TxSendService {
	ctx, cancel := context.WithCancel(context.Background())
	return &TxSendService{
		client:    c,
		signer:    signer,
		network:   network,
		chainID:   chainID,
		txHistory: txHistory,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Stop 停止后台确认等待
func (s *TxSendService) Stop() {
	s.cancel()
}

type SendTxRequest struct {
	To       string `json:"to"`
	Value    string `json:"value"`
	Gas      uint64 `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	GasTip   string `json:"gasTip,omitempty"`
	Data     string `json:"data,omitempty"`
}

type SendTxResponse struct {
	TxHash     string `json:"txHash"`
	From       string `json:"from"`
	To         string `json:"to"`
	Value      string `json:"value"`
	Nonce      uint64 `json:"nonce"`
	GasLimit   uint64 `json:"gasLimit"`
	GasPrice   string `json:"gasPrice"`
	Status     string `json:"status"`
}

func (s *TxSendService) SendTransaction(ctx context.Context, req SendTxRequest) (*SendTxResponse, error) {
	log.Printf("🔄 [TxSendService] 准备发送交易，目标: %s, 金额: %s", req.To, req.Value)

	if req.To == "" {
		return nil, fmt.Errorf("目标地址不能为空")
	}

	toAddr := common.HexToAddress(req.To)

	value, ok := new(big.Int).SetString(req.Value, 10)
	if !ok {
		return nil, fmt.Errorf("无效的金额格式")
	}

	fromAddr := s.signer.Address()

	nonce, err := s.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		log.Printf("❌ [TxSendService] 获取 nonce 失败: %v", err)
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	log.Printf("📝 [TxSendService] Nonce: %d", nonce)

	gasLimit := req.Gas
	if gasLimit == 0 {
		gasLimit = 21000
	}

	gasTipCap, err := s.client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Printf("⚠️ [TxSendService] 获取建议 Gas Tip 失败，使用默认值: %v", err)
		gasTipCap = big.NewInt(1_000_000_000)
	}

	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %w", err)
	}

	baseFee := header.BaseFee
	// 暂时强制使用 Legacy 交易以避免 EIP-1559 兼容性问题
	supportsEIP1559 := false
	log.Printf("📝 [TxSendService] 强制使用 Legacy 交易模式")

	var txData []byte
	if req.Data != "" {
		txData = common.FromHex(req.Data)
	}

	log.Printf("📝 [TxSendService] 网络支持 EIP-1559: %v, baseFee: %v", supportsEIP1559, baseFee)

	// 用于日志记录的 gas 价格
	gasPriceForLog := ""

	var signedTx *types.Transaction
	if supportsEIP1559 {
		// EIP-1559 交易
		gasFeeCap := new(big.Int).Add(
			new(big.Int).Mul(baseFee, big.NewInt(2)),
			gasTipCap,
		)
		gasPriceForLog = gasFeeCap.String()
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   s.chainID,
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gasLimit,
			To:        &toAddr,
			Value:     value,
			Data:      txData,
		})
		signedTx, err = s.signer.SignTx(ctx, tx, s.chainID)
	} else {
		// Legacy 交易 (用于不支持 EIP-1559 的网络)
		var gasPrice *big.Int
		
		// 如果用户提供了 gasPrice，则使用用户提供的值
		if req.GasPrice != "" {
			gasPrice, ok = new(big.Int).SetString(req.GasPrice, 10)
			if !ok {
				return nil, fmt.Errorf("无效的 gasPrice 格式")
			}
			log.Printf("📝 [TxSendService] 使用用户提供的 Gas Price: %s", gasPrice.String())
		} else {
			gasPrice, err = s.client.SuggestGasPrice(ctx)
			if err != nil {
				log.Printf("⚠️ [TxSendService] 获取 Gas Price 失败，使用默认值: %v", err)
				gasPrice = big.NewInt(1_000_000_000)
			}
		}
		gasPriceForLog = gasPrice.String()
		tx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &toAddr,
			Value:    value,
			Data:     txData,
		})
		signedTx, err = s.signer.SignTx(ctx, tx, s.chainID)
	}

	if err != nil {
		log.Printf("❌ [TxSendService] 签名交易失败: %v", err)
		
		// 保存失败记录
		if s.txHistory != nil {
			_ = s.txHistory.Add(store.TxHistoryEntry{
				TxHash:        "0x",
				FromAddr:      fromAddr.Hex(),
				ToAddr:        toAddr.Hex(),
				Value:         value.String(),
				GasLimit:      gasLimit,
				GasPrice:      gasPriceForLog,
				Nonce:         nonce,
				Data:          req.Data,
				Status:        store.TxStatusFailed,
				Network:       string(s.network),
				TxType:        "eth_transfer",
				FailureReason: fmt.Sprintf("签名失败: %v", err),
				CreatedAt:     time.Now(),
			})
		}
		
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	
	// 先保存交易记录（状态为待确认）
	if s.txHistory != nil {
		err = s.txHistory.Add(store.TxHistoryEntry{
			TxHash:     txHash,
			FromAddr:   fromAddr.Hex(),
			ToAddr:     toAddr.Hex(),
			Value:      value.String(),
			GasLimit:   gasLimit,
			GasPrice:   gasPriceForLog,
			Nonce:      nonce,
			Data:       req.Data,
			Status:     store.TxStatusPending,
			Network:    string(s.network),
			TxType:     "eth_transfer",
			CreatedAt:  time.Now(),
		})
		if err != nil {
			log.Printf("⚠️ [TxSendService] 保存交易记录失败: %v", err)
		}
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Printf("❌ [TxSendService] 发送交易失败: %v", err)
		
		// 更新交易记录为失败状态
		if s.txHistory != nil {
			_ = s.txHistory.UpdateStatusWithReason(txHash, store.TxStatusFailed, 0, fmt.Sprintf("%v", err))
		}
		
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("✅ [TxSendService] 交易发送成功: %s", txHash)

	go s.waitForTxConfirmation(txHash)

	return &SendTxResponse{
		TxHash:   txHash,
		From:     fromAddr.Hex(),
		To:       toAddr.Hex(),
		Value:    value.String(),
		Nonce:    nonce,
		GasLimit: gasLimit,
		GasPrice: gasPriceForLog,
		Status:   "pending",
	}, nil
}

const (
	maxConfirmAttempts = 120   // 最多轮询 120 次（约 10 分钟，5s 间隔）
	confirmBaseDelay   = 5    // 基础延迟（秒）
	confirmMaxDelay    = 30   // 最大延迟（秒）
)

func (s *TxSendService) waitForTxConfirmation(txHash string) {
	log.Printf("⏳ [TxSendService] 等待交易确认: %s", txHash)

	bgCtx := context.Background()
	for attempt := 1; attempt <= maxConfirmAttempts; attempt++ {
		select {
		case <-s.ctx.Done():
			return
		default:
			delay := confirmBaseDelay * attempt
			if delay > confirmMaxDelay {
				delay = confirmMaxDelay
			}
			time.Sleep(time.Duration(delay) * time.Second)

			receipt, err := s.client.TransactionReceipt(bgCtx, common.HexToHash(txHash))
			if err != nil {
				continue
			}

			if receipt != nil {
				blockNum := receipt.BlockNumber.Uint64()

				if receipt.Status == 0 {
					// 交易失败，尝试获取 revert 原因
					reason := s.getRevertReason(bgCtx, txHash, receipt)
					if reason != "" {
						log.Printf("❌ [TxSendService] 交易失败 (revert): %s, 原因: %s, 区块: %d (尝试 %d/%d)",
							txHash, reason, blockNum, attempt, maxConfirmAttempts)
					} else {
						log.Printf("❌ [TxSendService] 交易失败: %s, 区块: %d (尝试 %d/%d)",
							txHash, blockNum, attempt, maxConfirmAttempts)
					}

					if s.txHistory != nil {
						s.txHistory.UpdateStatusWithReason(txHash, store.TxStatusFailed, blockNum, reason)
					}
				} else {
					log.Printf("✅ [TxSendService] 交易成功: %s, 区块: %d (尝试 %d/%d)",
						txHash, blockNum, attempt, maxConfirmAttempts)

					if s.txHistory != nil {
						s.txHistory.UpdateStatus(txHash, store.TxStatusSuccess, blockNum)
					}
				}
				return
			}
		}
	}
	log.Printf("⚠️  [TxSendService] 交易确认超时: %s (已等待约 %d 秒)", txHash, maxConfirmAttempts*confirmBaseDelay)
}

// getRevertReason 通过重放交易获取 revert 原因
func (s *TxSendService) getRevertReason(ctx context.Context, txHash string, receipt *types.Receipt) string {
	tx, _, err := s.client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return ""
	}

	fromAddr, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return ""
	}

	// 在前一个区块重放交易以获取 revert 原因
	blockNum := new(big.Int).SetUint64(receipt.BlockNumber.Uint64())
	callBlock := new(big.Int).Sub(blockNum, big.NewInt(1))
	if callBlock.Sign() < 0 {
		callBlock = blockNum
	}

	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   tx.To(),
		Gas:  tx.Gas(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}

	// 根据交易类型设置相应的 gas 参数
	switch tx.Type() {
	case types.DynamicFeeTxType:
		// EIP-1559 交易
		msg.GasFeeCap = tx.GasFeeCap()
		msg.GasTipCap = tx.GasTipCap()
	case types.LegacyTxType:
		// Legacy 交易
		msg.GasPrice = tx.GasPrice()
	default:
		// 对于其他交易类型，优先使用 GasPrice
		if tx.GasPrice() != nil && tx.GasPrice().Sign() > 0 {
			msg.GasPrice = tx.GasPrice()
		} else {
			msg.GasFeeCap = tx.GasFeeCap()
			msg.GasTipCap = tx.GasTipCap()
		}
	}

	_, callErr := s.client.CallContract(ctx, msg, callBlock)
	if callErr == nil {
		return "unknown (replay succeeded)"
	}

	// 解析 revert 错误信息
	errStr := callErr.Error()
	if strings.Contains(errStr, "execution reverted") {
		// 提取 "execution reverted: <reason>" 或 "execution reverted"
		if idx := strings.Index(errStr, "execution reverted"); idx >= 0 {
			reasonPart := strings.TrimSpace(errStr[idx:])
			// 裁剪末尾的 method/signature 信息
			if methodIdx := strings.Index(reasonPart, "{method:"); methodIdx >= 0 {
				reasonPart = strings.TrimSpace(reasonPart[:methodIdx])
			}
			return reasonPart
		}
	}

	return errStr
}
