package api

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// corsAllowedOrigins 允许跨域访问的前端域名白名单，逗号分隔。
// 可通过环境变量 CORS_ALLOWED_ORIGINS 配置，默认可本地开发地址。
func corsAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:8080,http://127.0.0.1:8080,http://localhost:3000"
	}
	return strings.Split(raw, ",")
}

// corsMiddleware 添加 CORS 头，仅反射允许的 Origin，不使用通配符。
func corsMiddleware(next http.Handler) http.Handler {
	allowed := corsAllowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// 同源请求，不设置 CORS 头也合法
			next.ServeHTTP(w, r)
			return
		}

		originAllowed := false
		for _, o := range allowed {
			if strings.EqualFold(strings.TrimSpace(o), origin) {
				originAllowed = true
				break
			}
		}

		if !originAllowed {
			http.Error(w, `{"error":"origin not allowed"}`, http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 统一请求日志：方法、路径、耗时、状态码
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("📥 [HTTP] %s %s → %d (%v)", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// recoveryMiddleware 捕获 panic 并记录堆栈，返回 500
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("❌ [HTTP] PANIC recovered: %v\n%s", rec, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RequireSigner 返回签名器未配置的统一错误响应
func RequireSigner(w http.ResponseWriter, operation string) {
	log.Printf("⚠️  [API] 签名钱包未配置，拒绝操作: %s", operation)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(`{"error": "signer not configured, operation unavailable: ` + operation + `"}`))
}
