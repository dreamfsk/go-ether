package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/meu/go-ether/api"
)

func TestCorsMiddleware_NoOrigin(t *testing.T) {
	handler := api.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/block/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("No Origin: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("No Origin request should not have Access-Control-Allow-Origin header")
	}
}

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://allowed.com,https://other.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	handler := api.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/block/1", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Allowed origin: expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("Allowed origin should reflect Origin header, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://allowed.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	handler := api.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/tx/send", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Disallowed origin: expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "origin not allowed") {
		t.Errorf("Disallowed origin: body should contain 'origin not allowed', got %s", rec.Body.String())
	}
}

func TestCorsMiddleware_OPTIONS(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://allowed.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	handler := api.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/tx/send", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS: expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("OPTIONS should set Access-Control-Allow-Origin, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsAllowedOrigins_Default(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := api.CorsAllowedOrigins()
	if len(origins) == 0 {
		t.Error("default origins should not be empty")
	}
	found := false
	for _, o := range origins {
		if strings.TrimSpace(o) == "http://localhost:5173" {
			found = true
			break
		}
	}
	if !found {
		t.Error("default origins should include localhost:5173")
	}
}

func TestCorsAllowedOrigins_Custom(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://other.io")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := api.CorsAllowedOrigins()
	if len(origins) != 2 {
		t.Fatalf("custom origins: expected 2, got %d: %v", len(origins), origins)
	}
	if strings.TrimSpace(origins[0]) != "https://example.com" {
		t.Errorf("first origin = %s, want https://example.com", strings.TrimSpace(origins[0]))
	}
}

func TestLoggingMiddleware_PassesStatusCode(t *testing.T) {
	handler := api.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest("GET", "/api/teapot", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected 418, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	handler := api.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/api/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Recovery: expected 500, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_NormalRequest(t *testing.T) {
	handler := api.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/ok", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Normal request: expected 200, got %d", rec.Code)
	}
}

func TestRequireSigner(t *testing.T) {
	rec := httptest.NewRecorder()
	api.RequireSigner(rec, "ETH 交易发送")

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("RequireSigner: expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "signer not configured") {
		t.Errorf("RequireSigner: body should mention signer, got %s", rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("RequireSigner: Content-Type should be application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestResponseWriter_CapturesStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	rw := api.NewResponseWriter(w)

	if rw.StatusCode != http.StatusOK {
		t.Errorf("default StatusCode should be 200, got %d", rw.StatusCode)
	}

	rw.WriteHeader(http.StatusNotFound)
	if rw.StatusCode != http.StatusNotFound {
		t.Errorf("after WriteHeader: expected 404, got %d", rw.StatusCode)
	}
}
