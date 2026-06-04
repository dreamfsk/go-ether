package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

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
	To     string `json:"to"`
	Value  string `json:"value"`
	Gas    uint64 `json:"gas,omitempty"`
	GasTip string `json:"gasTip,omitempty"`
	Data   string `json:"data,omitempty"`
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
	if baseFee == nil {
		baseFee = big.NewInt(1_000_000_000)
	}

	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	var txData []byte
	if req.Data != "" {
		txData = common.FromHex(req.Data)
	}

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

	signedTx, err := s.signer.SignTx(ctx, tx, s.chainID)
	if err != nil {
		log.Printf("❌ [TxSendService] 签名交易失败: %v", err)
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Printf("❌ [TxSendService] 发送交易失败: %v", err)
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	log.Printf("✅ [TxSendService] 交易发送成功: %s", txHash)

	if s.txHistory != nil {
		err = s.txHistory.Add(store.TxHistoryEntry{
			TxHash:      txHash,
			FromAddr:    fromAddr.Hex(),
			ToAddr:      toAddr.Hex(),
			Value:       value.String(),
			GasLimit:    gasLimit,
			GasPrice:    gasFeeCap.String(),
			Nonce:       nonce,
			Data:        req.Data,
			Status:      store.TxStatusPending,
			BlockNumber: 0,
			Network:     string(s.network),
			TxType:      "eth_transfer",
			CreatedAt:   time.Now(),
		})
		if err != nil {
			log.Printf("⚠️ [TxSendService] 保存交易历史失败: %v", err)
		}
	}

	go s.waitForTxConfirmation(txHash)

	return &SendTxResponse{
		TxHash:   txHash,
		From:     fromAddr.Hex(),
		To:       toAddr.Hex(),
		Value:    value.String(),
		Nonce:    nonce,
		GasLimit: gasLimit,
		GasPrice: gasFeeCap.String(),
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
				status := store.TxStatusSuccess
				if receipt.Status == 0 {
					status = store.TxStatusFailed
				}

				if s.txHistory != nil {
					s.txHistory.UpdateStatus(txHash, status, receipt.BlockNumber.Uint64())
				}

				log.Printf("✅ [TxSendService] 交易已确认: %s, 状态: %d, 区块: %d (尝试 %d/%d)",
					txHash, receipt.Status, receipt.BlockNumber.Uint64(), attempt, maxConfirmAttempts)
				return
			}
		}
	}
	log.Printf("⚠️  [TxSendService] 交易确认超时: %s (已等待约 %d 秒)", txHash, maxConfirmAttempts*confirmBaseDelay)
}
