package tests

import (
	"math/big"
	"testing"

	"github.com/meu/go-ether/pkg/converter"
)

func TestWeiToEther(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected string
	}{
		{"zero", big.NewInt(0), "0.000000000000000000"},
		{"1 wei", big.NewInt(1), "0.000000000000000001"},
		{"1 gwei", big.NewInt(1e9), "0.000000001000000000"},
		{"0.1 ETH", big.NewInt(1e17), "0.100000000000000000"},
		{"1 ETH", big.NewInt(1e18), "1.000000000000000000"},
		{"100 ETH", new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)), "100.000000000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.WeiToEtherString(tt.wei)
			if err != nil {
				t.Fatalf("WeiToEtherString returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("WeiToEtherString(%v) = %s, want %s", tt.wei, result, tt.expected)
			}
		})
	}
}

func TestEtherStrToWei(t *testing.T) {
	tests := []struct {
		name     string
		etherStr string
		expected *big.Int
		wantErr  bool
	}{
		{"0", "0", big.NewInt(0), false},
		{"0.1", "0.1", big.NewInt(1e17), false},
		{"1", "1", big.NewInt(1e18), false},
		{"100", "100", new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)), false},
		{"1.5", "1.5", big.NewInt(1500000000000000000), false},
		{"invalid", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.EtherStrToWei(tt.etherStr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EtherStrToWei(%q) expected error, got nil", tt.etherStr)
				}
				return
			}
			if err != nil {
				t.Errorf("EtherStrToWei(%q) returned error: %v", tt.etherStr, err)
				return
			}
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("EtherStrToWei(%q) = %v, want %v", tt.etherStr, result, tt.expected)
			}
		})
	}
}

func TestTokenAmountToDecimals(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		decimals uint8
		expected *big.Int
		wantErr  bool
	}{
		{"integer 18 decimals", "1", 18, big.NewInt(1000000000000000000), false},
		{"decimal 18 decimals", "0.1", 18, big.NewInt(100000000000000000), false},
		{"decimal 6 decimals", "1.5", 6, big.NewInt(1500000), false},
		{"zero", "0", 18, big.NewInt(0), false},
		{"too many decimals", "0.1234567", 6, nil, true},
		{"empty", "", 18, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.TokenAmountToDecimals(tt.amount, tt.decimals)
			if tt.wantErr {
				if err == nil {
					t.Errorf("TokenAmountToDecimals(%q, %d) expected error, got nil", tt.amount, tt.decimals)
				}
				return
			}
			if err != nil {
				t.Errorf("TokenAmountToDecimals(%q, %d) returned error: %v", tt.amount, tt.decimals, err)
				return
			}
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("TokenAmountToDecimals(%q, %d) = %v, want %v", tt.amount, tt.decimals, result, tt.expected)
			}
		})
	}
}

func TestWeiToGwei(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected float64
	}{
		{"1 gwei", big.NewInt(1e9), 1.0},
		{"100 gwei", big.NewInt(1e11), 100.0},
		{"0.5 gwei", big.NewInt(5e8), 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.WeiToGwei(tt.wei)
			if err != nil {
				t.Fatalf("WeiToGwei returned error: %v", err)
			}
			floatVal, _ := result.Float64()
			if floatVal != tt.expected {
				t.Errorf("WeiToGwei(%v) = %v, want %v", tt.wei, floatVal, tt.expected)
			}
		})
	}
}