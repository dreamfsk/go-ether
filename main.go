package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/meu/go-ether/api"
	"github.com/meu/go-ether/client"
	"github.com/meu/go-ether/config"
	"github.com/meu/go-ether/service"
	"github.com/meu/go-ether/store"
	"github.com/meu/go-ether/wallet"
)

func main() {
	log.Println("=============================================")
	log.Println("  迷你区块浏览器与 ERC-20 监听服务 启动中...")
	log.Println("=============================================")

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables only")
	}

	cfg := config.Load()

	networkURL := cfg.GetRPCURL()
	if networkURL == "" {
		log.Fatal("❌ ETH_RPC_URL must be set")
	}

	log.Printf("✅ 配置加载完成")
	log.Printf("   - 当前网络: %s", cfg.Network)
	log.Printf("   - RPC URL: %s", cfg.GetRPCURL())
	log.Printf("   - WS URL: %s", cfg.GetWSURL())
	log.Printf("   - ChainID: %s", cfg.NetworkConfig.ChainID.String())
	log.Printf("   - 默认合约: %s", cfg.ERC20Contract)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("🔗 正在连接以太坊节点...")
	multiClient, err := client.NewMultiClient(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	log.Println("✅ 以太坊节点连接成功")
	defer multiClient.Close()

	log.Println("📦 初始化 SQLite 存储...")
	txHistoryStore, err := store.NewTxHistoryStore("transactions.db")
	if err != nil {
		log.Fatalf("❌ SQLite 初始化失败: %v", err)
	}
	log.Println("✅ SQLite 交易历史存储初始化完成")
	defer txHistoryStore.Close()

	contractStore, err := store.NewContractStore("contracts.db")
	if err != nil {
		log.Fatalf("❌ 合约存储初始化失败: %v", err)
	}
	log.Println("✅ SQLite 合约存储初始化完成")
	defer contractStore.Close()

	var signer wallet.Signer
	log.Println("🔐 初始化钱包...")
	envSigner, err := wallet.NewEnvSigner()
	if err != nil {
		log.Printf("⚠️  钱包初始化失败: %v", err)
		log.Println("   交易发送功能将不可用")
		signer = nil
	} else {
		signer = envSigner
		log.Printf("✅ 钱包初始化成功")
	}

	log.Println("🔧 初始化服务组件...")
	blockService := service.NewBlockService(multiClient.RPC())
	txService := service.NewTxService(multiClient.RPC())

	contractManager := service.NewContractManager(
		multiClient, signer, contractStore, txHistoryStore,
		string(cfg.Network), cfg.NetworkConfig.ChainID,
	)

	if err := contractManager.Initialize(cfg.ERC20Contract); err != nil {
		log.Fatalf("❌ 合约管理器初始化失败: %v", err)
	}

	log.Println("✅ 服务组件初始化完成")
	if signer == nil {
		log.Println("⚠️  未配置签名钱包，将以只读模式运行（交易发送、合约写调用、代币操作不可用）")
	} else {
		log.Println("🔐 签名钱包已配置，全功能模式运行")
	}

	currentContract := contractManager.GetCurrentContract()
	if currentContract != nil {
		log.Printf("📋 当前合约地址: %s", currentContract.Address)
		contractManager.StartListening()
	} else {
		log.Println("⚠️  未配置合约地址，事件监听未启动")
	}

	log.Println("🌐 启动 HTTP API 服务器 (端口: 8080)...")
	handlers := api.NewHandlers(blockService, txService, txHistoryStore)
	txHandlers := api.NewTxHandlers(contractManager)
	contractHandlers := api.NewContractHandlers(contractManager)

	// 前端静态文件服务
	staticHandler := createStaticHandler()
	server := api.NewServer(handlers, txHandlers, contractHandlers, ":8080", staticHandler)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP 服务器错误: %v", err)
		}
	}()
	log.Println("✅ HTTP API 服务器启动完成")
	log.Println("=============================================")
	log.Println("  🚀 服务已就绪，等待请求...")
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
	contractManager.Stop()
	cancel()

	log.Println("🔄 关闭数据库...")
	txHistoryStore.Close()
	contractStore.Close()
	log.Println("✅ 数据库已关闭")

	log.Println("=============================================")
	log.Println("  ✅ 服务已完全关闭，再见！")
	log.Println("=============================================")
}

// createStaticHandler 创建前端静态文件处理器，支持 SPA fallback
func createStaticHandler() http.Handler {
	distPath := "web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Printf("⚠️  前端静态文件目录不存在: %s", distPath)
		return nil
	}

	// 获取绝对路径用于路径穿越校验
	absDistPath, err := filepath.Abs(distPath)
	if err != nil {
		log.Printf("⚠️  无法获取静态文件目录绝对路径: %v", err)
		return nil
	}
	absDistPath = filepath.Clean(absDistPath) + string(filepath.Separator)

	fs := http.FileServer(http.Dir(distPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 去掉 /manage 前缀
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/manage")
		if r.URL.Path == "" || r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}

		// 路径穿越防护：清理后验证仍在 dist 目录内
		cleanedPath := filepath.Clean(r.URL.Path)
		fullPath := filepath.Join(absDistPath, cleanedPath)
		if !strings.HasPrefix(fullPath, absDistPath) {
			log.Printf("⚠️  [Static] 拒绝路径穿越访问: %s", r.URL.Path)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// SPA fallback: 文件不存在时返回 index.html
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			r.URL.Path = "/index.html"
		} else {
			r.URL.Path = "/" + cleanedPath
		}

		fs.ServeHTTP(w, r)
	})
}
