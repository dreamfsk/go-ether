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
	log.Printf("HTTP server listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
