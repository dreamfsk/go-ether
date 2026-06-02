package config

type WalletConfig struct {
	PrivateKeyEnv string
}

func LoadWalletConfig() *WalletConfig {
	return &WalletConfig{
		PrivateKeyEnv: "SENDER_PRIVATE_KEY",
	}
}
