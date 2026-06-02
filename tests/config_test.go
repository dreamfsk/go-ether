package tests

import (
	"math/big"
	"os"
	"testing"

	"github.com/meu/go-ether/config"
)

func TestGetNetworkConfig(t *testing.T) {
	tests := []struct {
		name        string
		network     config.NetworkType
		wantName    config.NetworkType
		wantChainID *big.Int
		wantErr     bool
	}{
		{
			name:        "sepolia network",
			network:     config.NetworkSepolia,
			wantName:    config.NetworkSepolia,
			wantChainID: big.NewInt(11155111),
			wantErr:     false,
		},
		{
			name:        "local network",
			network:     config.NetworkLocal,
			wantName:    config.NetworkLocal,
			wantChainID: big.NewInt(31337),
			wantErr:     false,
		},
		{
			name:        "unsupported network",
			network:     "unknown",
			wantName:    "",
			wantChainID: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.GetNetworkConfig(tt.network)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetNetworkConfig() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetNetworkConfig() unexpected error: %v", err)
				return
			}
			if cfg.Name != tt.wantName {
				t.Errorf("GetNetworkConfig() Name = %v, want %v", cfg.Name, tt.wantName)
			}
			if cfg.ChainID.Cmp(tt.wantChainID) != 0 {
				t.Errorf("GetNetworkConfig() ChainID = %v, want %v", cfg.ChainID, tt.wantChainID)
			}
		})
	}
}

func TestGetDefaultNetwork(t *testing.T) {
	expected := config.NetworkLocal
	result := config.GetDefaultNetwork()
	if result != expected {
		t.Errorf("GetDefaultNetwork() = %v, want %v", result, expected)
	}
}

func TestConfig_GetNodeURL(t *testing.T) {
	tests := []struct {
		name     string
		rpcURL   string
		wsURL    string
		expected string
	}{
		{
			name:     "WS URL takes precedence",
			rpcURL:   "http://localhost:8545",
			wsURL:    "ws://localhost:8545",
			expected: "ws://localhost:8545",
		},
		{
			name:     "fallback to RPC URL",
			rpcURL:   "http://localhost:8545",
			wsURL:    "",
			expected: "http://localhost:8545",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				NetworkConfig: config.NetworkConfig{
					RPCURL: tt.rpcURL,
					WSURL:  tt.wsURL,
				},
			}
			result := cfg.GetNodeURL()
			if result != tt.expected {
				t.Errorf("GetNodeURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	originalNetwork := os.Getenv("NETWORK")
	originalRPC := os.Getenv("ETH_RPC_URL")
	originalWS := os.Getenv("ETH_WS_URL")
	originalContract := os.Getenv("ERC20_CONTRACT")

	// Clean up
	os.Unsetenv("NETWORK")
	os.Unsetenv("ETH_RPC_URL")
	os.Unsetenv("ETH_WS_URL")
	os.Unsetenv("ERC20_CONTRACT")

	// Restore after test
	defer func() {
		os.Setenv("NETWORK", originalNetwork)
		os.Setenv("ETH_RPC_URL", originalRPC)
		os.Setenv("ETH_WS_URL", originalWS)
		os.Setenv("ERC20_CONTRACT", originalContract)
	}()

	// Test default config
	cfg := config.Load()
	if cfg.Network != config.NetworkLocal {
		t.Errorf("Load() default Network = %v, want %v", cfg.Network, config.NetworkLocal)
	}

	// Test with custom network
	os.Setenv("NETWORK", "sepolia")
	cfg = config.Load()
	if cfg.Network != config.NetworkSepolia {
		t.Errorf("Load() Network with env = %v, want %v", cfg.Network, config.NetworkSepolia)
	}

	// Test with custom RPC URL
	customRPC := "http://custom-rpc.example.com"
	os.Setenv("ETH_RPC_URL", customRPC)
	cfg = config.Load()
	if cfg.NetworkConfig.RPCURL != customRPC {
		t.Errorf("Load() RPCURL = %v, want %v", cfg.NetworkConfig.RPCURL, customRPC)
	}
}