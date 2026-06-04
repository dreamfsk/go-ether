package config

import (
	"os"
)

type WalletConfig struct {
	// Keystore 方式配置
	KeystorePath string // keystore 文件路径或目录路径
	KeystorePassword string // keystore 密码
	
	// 环境变量方式配置（兼容旧版本）
	PrivateKeyEnv string
}

// LoadWalletConfig 加载钱包配置
func LoadWalletConfig() *WalletConfig {
	return &WalletConfig{
		KeystorePath:    os.Getenv("KEYSTORE_PATH"),
		KeystorePassword: os.Getenv("KEYSTORE_PASSWORD"),
		PrivateKeyEnv:   "SENDER_PRIVATE_KEY",
	}
}
