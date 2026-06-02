package tests

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meu/go-ether/wallet"
)

// TestNewEnvSigner tests the NewEnvSigner function with various inputs
func TestNewEnvSigner(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("SENDER_PRIVATE_KEY")
	defer os.Setenv("SENDER_PRIVATE_KEY", originalKey)

	tests := []struct {
		name        string
		privateKey  string
		wantErr     bool
	}{
		{
			name:        "valid private key without 0x",
			privateKey:  "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d",
			wantErr:     false,
		},
		{
			name:        "valid private key with 0x",
			privateKey:  "0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d",
			wantErr:     false,
		},
		{
			name:       "empty private key",
			privateKey: "",
			wantErr:    true,
		},
		{
			name:       "invalid private key",
			privateKey: "invalid-key",
			wantErr:    true,
		},
		{
			name:       "short private key",
			privateKey: "abc123",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SENDER_PRIVATE_KEY", tt.privateKey)

			signer, err := wallet.NewEnvSigner()
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEnvSigner() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewEnvSigner() unexpected error: %v", err)
				return
			}

			if signer == nil {
				t.Error("NewEnvSigner() returned nil signer")
			}
		})
	}
}

func TestEnvSigner_Address(t *testing.T) {
	// Use a known valid test private key
	// This corresponds to address 0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1
	os.Setenv("SENDER_PRIVATE_KEY", "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	defer os.Unsetenv("SENDER_PRIVATE_KEY")

	signer, err := wallet.NewEnvSigner()
	if err != nil {
		t.Fatalf("NewEnvSigner() failed: %v", err)
	}

	addr := signer.Address()
	expectedAddr := common.HexToAddress("0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1")

	if addr != expectedAddr {
		t.Errorf("Address() = %v, want %v", addr.Hex(), expectedAddr.Hex())
	}
}