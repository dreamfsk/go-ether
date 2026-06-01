package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/meu/go-ether/api"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/service"
	"github.com/meu/go-ether/store"
)

func main() {
	log.Println("=============================================")
	log.Println("  迷你区块浏览器与 ERC-20 监听服务 启动中...")
	log.Println("=============================================")

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables only")
	}

	cfg := config.Load()

	nodeURL := cfg.GetNodeURL()
	if nodeURL == "" {
		log.Fatal("❌ ETH_WS_URL or ETH_RPC_URL must be set")
	}

	if cfg.ERC20Contract == "" {
		log.Fatal("❌ ERC20_CONTRACT must be set")
	}

	log.Printf("✅ 配置加载完成")
	log.Printf("   - 节点 URL: %s", nodeURL)
	log.Printf("   - ERC20 合约: %s", cfg.ERC20Contract)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("🔗 正在连接以太坊节点...")
	ethClient, err := client.New(ctx, nodeURL)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	log.Println("✅ 以太坊节点连接成功")
	defer ethClient.Close()

	log.Println("📦 初始化事件存储 (最多 100 条)...")
	eventStore := store.NewEventStore(100)
	log.Println("✅ 事件存储初始化完成")

	log.Println("🔧 初始化服务组件...")
	blockService := service.NewBlockService(ethClient)
	txService := service.NewTxService(ethClient)
	eventService, err := service.NewEventService(ethClient, eventStore, cfg.ERC20Contract)
	if err != nil {
		log.Fatalf("❌ 事件服务初始化失败: %v", err)
	}
	log.Println("✅ 服务组件初始化完成")

	log.Println("👂 启动 ERC20 Transfer 事件监听...")
	go eventService.StartListening(ctx)

	log.Println("🌐 启动 HTTP API 服务器 (端口: 8080)...")
	handlers := api.NewHandlers(blockService, txService, eventStore)
	server := api.NewServer(handlers, ":8080")

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP 服务器错误: %v", err)
		}
	}()
	log.Println("✅ HTTP API 服务器启动完成")
	log.Println("=============================================")
	log.Println("  🚀 服务已就绪，等待请求...")
	log.Println("  📡 API 端点:")
	log.Println("     - GET /api/block/{id}")
	log.Println("     - GET /api/tx/{hash}")
	log.Println("     - GET /api/events")
	log.Println("=============================================")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	fmt.Printf("\n📡 收到信号 %s，正在优雅关闭...\n", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	log.Println("🔄 关闭 HTTP 服务器...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  HTTP 服务器关闭警告: %v", err)
	}
	log.Println("✅ HTTP 服务器已关闭")

	log.Println("🔄 停止事件监听...")
	cancel()

	log.Println("=============================================")
	log.Println("  ✅ 服务已完全关闭，再见！")
	log.Println("=============================================")
}
