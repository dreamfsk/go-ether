package api

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	handlers   *Handlers
}

func NewServer(handlers *Handlers, addr string) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/block/", handlers.GetBlock)
	mux.HandleFunc("/api/tx/", handlers.GetTransaction)
	mux.HandleFunc("/api/events", handlers.GetEvents)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		handlers: handlers,
	}
}

func (s *Server) Start() error {
	log.Printf("🌐 [HTTP] 服务器启动中，监听地址: %s", s.httpServer.Addr)
	log.Println("🗺️  [HTTP] 已注册路由:")
	log.Println("   - GET /api/block/{id}")
	log.Println("   - GET /api/tx/{hash}")
	log.Println("   - GET /api/events")
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
