package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/store"
)

const abiFilePath = "build/MyERC20.abi"

func loadABI() (string, error) {
	absPath, err := filepath.Abs(abiFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read ABI file (%s): %w。请运行 node compile.js 编译合约", absPath, err)
	}

	return string(data), nil
}

type EventService struct {
	client    *client.EthClient
	txHistory *store.TxHistoryStore
	contract  common.Address
	network   string
	abi       abi.ABI
}

func NewEventService(c *client.EthClient, txHistory *store.TxHistoryStore, network config.NetworkType, contractAddr string) (*EventService, error) {
	log.Printf("🔧 [EventService] 初始化，合约地址: %s", contractAddr)

	abiJSON, err := loadABI()
	if err != nil {
		log.Printf("❌ [EventService] 加载 ABI 文件失败: %v", err)
		return nil, fmt.Errorf("failed to load ABI: %w", err)
	}
	log.Println("📄 [EventService] ABI 文件加载成功")

	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		log.Printf("❌ [EventService] ABI 解析失败: %v", err)
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	log.Println("✅ [EventService] ABI 解析成功")
	return &EventService{
		client:    c,
		txHistory: txHistory,
		contract:  common.HexToAddress(contractAddr),
		network:   string(network),
		abi:       parsedABI,
	}, nil
}

func (s *EventService) StartListening(ctx context.Context) {
	log.Println("👂 [EventService] 开始订阅 ERC20 Transfer 事件...")

	// 启动时回填现有记录的 contract_addr
	go s.migrateContractAddr(ctx)

	var attempt int

RECONNECT:
	for {
		select {
		case <-ctx.Done():
			log.Println("🔄 [EventService] 上下文被取消，停止监听")
			return
		default:
		}

		if attempt > 0 {
			// 连接中断后才打印重连信息，首次连接不打印
			log.Printf("🔄 [EventService] 连接尝试 #%d", attempt)
		}

		query := ethereum.FilterQuery{
			Addresses: []common.Address{s.contract},
		}

		logsCh := make(chan types.Log)
		sub, err := s.client.SubscribeFilterLogs(ctx, query, logsCh)
		if err != nil {
			log.Printf("❌ [EventService] 事件订阅失败: %v", err)
			attempt++
			s.sleepWithBackoff(ctx, attempt)
			continue RECONNECT
		}

		// 连接成功后重置计数器
		attempt = 0
		log.Printf("✅ [EventService] 事件订阅成功，监听合约: %s", s.contract.Hex())

		for {
			select {
			case vLog := <-logsCh:
				s.processLog(vLog)
			case err := <-sub.Err():
				log.Printf("❌ [EventService] 订阅连接断开: %v", err)
				sub.Unsubscribe()
				attempt++
				s.sleepWithBackoff(ctx, attempt)
				continue RECONNECT
			case <-ctx.Done():
				log.Println("🔄 [EventService] 上下文被取消，停止监听")
				sub.Unsubscribe()
				return
			}
		}
	}
}

func (s *EventService) sleepWithBackoff(ctx context.Context, attempt int) {
	sec := int(math.Min(60, math.Pow(2, float64(attempt))))
	d := time.Duration(sec) * time.Second
	log.Printf("⏳ [EventService] 将在 %s 后尝试重连", d)

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
	case <-ctx.Done():
	}
}

// migrateContractAddr 回填旧有 ERC-20 记录的 contract_addr（从链上交易收据获取）
func (s *EventService) migrateContractAddr(ctx context.Context) {
	log.Println("🔄 [EventService] 开始回填 contract_addr...")

	entries, err := s.txHistory.ListByContractAddr("", 50, 0)
	if err != nil {
		log.Printf("❌ [EventService] 回填 contract_addr 失败: %v", err)
		return
	}

	if len(entries) == 0 {
		log.Println("✅ [EventService] contract_addr 回填完成（无需回填）")
		return
	}

	updated := 0
	for _, entry := range entries {
		receipt, err := s.client.TransactionReceipt(ctx, common.HexToHash(entry.TxHash))
		if err != nil {
			log.Printf("⚠️  [EventService] 获取交易收据失败 %s: %v", entry.TxHash, err)
			continue
		}

		// 从交易收据的 logs 中找到 Transfer 事件，它的 Address 就是合约地址
		for _, receiptLog := range receipt.Logs {
			if len(receiptLog.Topics) > 0 && receiptLog.Topics[0] == s.abi.Events["Transfer"].ID {
				if err := s.txHistory.UpdateContractAddr(entry.TxHash, receiptLog.Address.Hex()); err != nil {
					log.Printf("⚠️  [EventService] 更新 contract_addr 失败 %s: %v", entry.TxHash, err)
				} else {
					updated++
				}
				break
			}
		}
	}

	log.Printf("✅ [EventService] contract_addr 回填完成: 更新了 %d/%d 条记录", updated, len(entries))
}

func (s *EventService) processLog(vLog types.Log) {
	if len(vLog.Topics) == 0 {
		log.Printf("⚠️  [EventService] 跳过无效日志（无 Topics）")
		return
	}

	// 检查是否为 Transfer 事件 (topic hash)
	transferEventSig := s.abi.Events["Transfer"].ID
	if vLog.Topics[0] != transferEventSig {
		return
	}

	var event struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	// ERC20 Transfer 事件的 from/to 是 indexed 参数，位于 topics[1] 和 topics[2]
	if len(vLog.Topics) >= 3 {
		event.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		event.To = common.BytesToAddress(vLog.Topics[2].Bytes())
	}

	// value 是 non-indexed 参数，需要从 data 中单独解码
	if len(vLog.Data) > 0 {
		var valueOnly struct {
			Value *big.Int
		}
		if err := s.abi.UnpackIntoInterface(&valueOnly, "Transfer", vLog.Data); err != nil {
			log.Printf("❌ [EventService] 日志 value 解码失败: %v", err)
			return
		}
		event.Value = valueOnly.Value
	}

	log.Printf("💸 [EventService] 捕获 Transfer 事件: 从 %s 到 %s, 数量: %s",
		event.From.Hex(), event.To.Hex(), event.Value.String())

	// 写入 SQLite 持久化
	if s.txHistory != nil {
		if err := s.txHistory.Add(store.TxHistoryEntry{
			TxHash:      vLog.TxHash.Hex(),
			FromAddr:    event.From.Hex(),
			ToAddr:      event.To.Hex(),
			Value:       event.Value.String(),
			BlockNumber:  vLog.BlockNumber,
			Status:      store.TxStatusSuccess, // 链上事件已确认
			Network:     s.network,
			TxType:      "erc20_transfer",
			ContractAddr: vLog.Address.Hex(), // 记录事件所属合约地址
			CreatedAt:   time.Now(),
		}); err != nil {
			log.Printf("⚠️  [EventService] 持久化事件失败: %v", err)
		}
	}
}
