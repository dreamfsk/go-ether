package client

import (
	"context"
	"fmt"
	"log"

	"github.com/meu/go-ether/config"
)

// MultiClient 管理 RPC (HTTP) 和 WebSocket 双通道客户端。
// RPC 客户端用于合约调用、交易查询、Gas 估算等；
// WebSocket 客户端用于事件订阅。
type MultiClient struct {
	rpc *EthClient
	ws  *EthClient
}

// NewMultiClient 创建多通道客户端，至少需要 RPC 或 WS 之一可用。
// 如果 WS URL 与 RPC URL 相同，则共享同一个客户端。
func NewMultiClient(ctx context.Context, cfg *config.Config) (*MultiClient, error) {
	mc := &MultiClient{}

	rpcURL := cfg.GetRPCURL()
	wsURL := cfg.GetWSURL()

	var err error
	if rpcURL != "" {
		mc.rpc, err = New(ctx, rpcURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect RPC (%s): %w", rpcURL, err)
		}
		log.Printf("✅ RPC 客户端连接成功: %s", rpcURL)
	}

	if wsURL != "" {
		if wsURL == rpcURL && mc.rpc != nil {
			// 同一地址，共享客户端
			mc.ws = mc.rpc
			log.Println("✅ WebSocket 客户端与 RPC 共享连接")
		} else {
			mc.ws, err = New(ctx, wsURL)
			if err != nil {
				log.Printf("⚠️  WebSocket 连接失败 (%s): %v，事件监听不可用", wsURL, err)
				mc.ws = nil
			} else {
				log.Printf("✅ WebSocket 客户端连接成功: %s", wsURL)
			}
		}
	}

	if mc.rpc == nil && mc.ws == nil {
		return nil, fmt.Errorf("both RPC and WebSocket connections failed")
	}

	// 如果只有 WS 没有 RPC，用 WS 兜底
	if mc.rpc == nil {
		mc.rpc = mc.ws
	}

	return mc, nil
}

// RPC 返回 HTTP RPC 客户端（合约调用、交易查询等）
func (mc *MultiClient) RPC() *EthClient {
	return mc.rpc
}

// WS 返回 WebSocket 客户端（事件订阅），可能为 nil
func (mc *MultiClient) WS() *EthClient {
	return mc.ws
}

// Close 关闭所有连接（注意 RPC 和 WS 可能指向同一个客户端）
func (mc *MultiClient) Close() {
	if mc.ws != nil && mc.ws != mc.rpc {
		mc.ws.Client.Close()
	}
	if mc.rpc != nil {
		mc.rpc.Client.Close()
	}
}
