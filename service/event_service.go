package service

import (
	"context"
	"fmt"
	"log"
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

	query := ethereum.FilterQuery{
		Addresses: []common.Address{s.contract},
	}

	logsCh := make(chan types.Log)
	sub, err := s.client.SubscribeFilterLogs(ctx, query, logsCh)
	if err != nil {
		log.Printf("❌ [EventService] 事件订阅失败: %v", err)
		return
	}
	defer sub.Unsubscribe()

	log.Printf("✅ [EventService] 事件订阅成功，监听合约: %s", s.contract.Hex())

	for {
		select {
		case vLog := <-logsCh:
			log.Printf("📨 [EventService] 收到日志，区块 #%d，交易: %s", vLog.BlockNumber, vLog.TxHash.Hex())
			s.processLog(vLog)
		case err := <-sub.Err():
			log.Printf("❌ [EventService] 订阅错误: %v", err)
			return
		case <-ctx.Done():
			log.Println("🔄 [EventService] 上下文被取消，停止监听")
			return
		}
	}
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

	if err := s.abi.UnpackIntoInterface(&event, "Transfer", vLog.Data); err != nil {
		log.Printf("❌ [EventService] 日志数据解析失败: %v", err)
		return
	}

	if len(vLog.Topics) >= 3 {
		event.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		event.To = common.BytesToAddress(vLog.Topics[2].Bytes())
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
			BlockNumber: vLog.BlockNumber,
			Status:      store.TxStatusSuccess, // 链上事件已确认
			Network:     s.network,
			TxType:      "erc20_transfer",
			CreatedAt:   time.Now(),
		}); err != nil {
			log.Printf("⚠️  [EventService] 持久化事件失败: %v", err)
		}
	}
}
