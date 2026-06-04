package api

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer       *http.Server
	handlers         *Handlers
	txHandlers       *TxHandlers
	contractHandlers *ContractHandlers
}

func NewServer(handlers *Handlers, txHandlers *TxHandlers, contractHandlers *ContractHandlers, addr string, staticHandler http.Handler) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/block/", handlers.GetBlock)
	mux.HandleFunc("/api/tx/", handlers.GetTransaction)
	mux.HandleFunc("/api/events", handlers.GetEvents)
	mux.HandleFunc("/api/tx/send", txHandlers.SendTransaction)
	mux.HandleFunc("/api/tx/history", txHandlers.GetTxHistory)
	mux.HandleFunc("/api/tx/detail", txHandlers.GetTxByHash)
	mux.HandleFunc("/api/contract/list", contractHandlers.ContractList)
	mux.HandleFunc("/api/contract/switch", contractHandlers.ContractSwitch)
	mux.HandleFunc("/api/contract/current", contractHandlers.ContractCurrent)
	mux.HandleFunc("/api/contract/view", contractHandlers.ContractView)
	mux.HandleFunc("/api/contract/call", contractHandlers.ContractCall)
	mux.HandleFunc("/api/token/info", contractHandlers.TokenInfo)
	mux.HandleFunc("/api/token/balance", contractHandlers.TokenBalance)
	mux.HandleFunc("/api/token/transfer", contractHandlers.TokenTransfer)
	mux.HandleFunc("/api/token/mint", contractHandlers.TokenMint)
	mux.HandleFunc("/api/token/deploy", contractHandlers.TokenDeploy)
	mux.HandleFunc("/api/config", contractHandlers.Config)

	if staticHandler != nil {
		mux.Handle("/manage/", staticHandler)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: RecoveryMiddleware(LoggingMiddleware(CorsMiddleware(mux))),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		handlers:         handlers,
		txHandlers:       txHandlers,
		contractHandlers: contractHandlers,
	}
}

func (s *Server) Start() error {
	log.Printf("🌐 [HTTP] 服务器启动中，监听地址: %s", s.httpServer.Addr)
	log.Println("🗺️  [HTTP] 已注册路由:")
	log.Println("   - GET /api/block/{id}")
	log.Println("   - GET /api/tx/{hash}")
	log.Println("   - GET /api/events")
	log.Println("   - POST /api/tx/send")
	log.Println("   - GET /api/tx/history")
	log.Println("   - GET /api/tx/detail")
	log.Println("   - GET /api/contract/list")
	log.Println("   - POST /api/contract/switch")
	log.Println("   - GET /api/contract/current")
	log.Println("   - POST /api/contract/view")
	log.Println("   - POST /api/contract/call")
	log.Println("   - GET /api/token/info")
	log.Println("   - GET /api/token/balance")
	log.Println("   - POST /api/token/transfer")
	log.Println("   - POST /api/token/mint")
	log.Println("   - POST /api/token/deploy")
	log.Println("   - GET /api/config")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("🔄 [HTTP] 正在优雅关闭服务器...")
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Printf("⚠️  [HTTP] 服务器关闭警告: %v", err)
	} else {
		log.Println("✅ [HTTP] 服务器已成功关闭")
	}
	return err
}
