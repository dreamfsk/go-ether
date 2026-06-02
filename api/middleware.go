package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// RequireSigner 校验签名服务是否可用。
// 先决条件：调用方已判断对应 service 为 nil。
// 写入 HTTP 503 + JSON 错误响应，提示用户配置 SENDER_PRIVATE_KEY。
func RequireSigner(w http.ResponseWriter, name string) {
	msg := fmt.Sprintf("%s失败：未配置签名操作无法执行", name)
	log.Printf("❌ [API] %s", msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}
