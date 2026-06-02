package converter

import (
	"fmt"
	"math/big"
	"strings"
)

const (
	WeiPerEther   = 1e18
	GweiPerEther  = 1e9
	WeiPerGwei    = 1e9
)

// WeiToEther converts wei to ether
func WeiToEther(wei *big.Int) (*big.Float, error) {
	if wei == nil {
		return nil, fmt.Errorf("wei is nil")
	}
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(WeiPerEther))
	return ether, nil
}

// EtherToWei converts ether to wei
func EtherToWei(ether *big.Float) (*big.Int, error) {
	if ether == nil {
		return nil, fmt.Errorf("ether is nil")
	}
	wei := new(big.Float).Mul(ether, big.NewFloat(WeiPerEther))
	result, _ := wei.Int(new(big.Int))
	return result, nil
}

// EtherStrToWei converts ether string to wei
func EtherStrToWei(etherStr string) (*big.Int, error) {
	ether, _, err := big.ParseFloat(etherStr, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid ether amount: %w", err)
	}
	return EtherToWei(ether)
}

// WeiToEtherString converts wei to ether string
func WeiToEtherString(wei *big.Int) (string, error) {
	ether, err := WeiToEther(wei)
	if err != nil {
		return "", err
	}
	return ether.Text('f', 18), nil
}

// WeiToGwei converts wei to gwei
func WeiToGwei(wei *big.Int) (*big.Float, error) {
	if wei == nil {
		return nil, fmt.Errorf("wei is nil")
	}
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(WeiPerGwei))
	return gwei, nil
}

// GweiToWei converts gwei to wei
func GweiToWei(gwei *big.Float) (*big.Int, error) {
	if gwei == nil {
		return nil, fmt.Errorf("gwei is nil")
	}
	wei := new(big.Float).Mul(gwei, big.NewFloat(WeiPerGwei))
	result, _ := wei.Int(new(big.Int))
	return result, nil
}

// TokenAmountToDecimals converts token amount with specified decimals to smallest unit
func TokenAmountToDecimals(amount string, decimals uint8) (*big.Int, error) {
	if amount == "" {
		return nil, fmt.Errorf("amount is empty")
	}

	// 检查是否包含小数点
	parts := strings.SplitN(amount, ".", 2)
	if len(parts) == 1 {
		// 没有小数点，直接转换
		result, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid integer amount")
		}
		// 补零
		return new(big.Int).Mul(result, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)), nil
	}

	// 有小数点
	integerPart := parts[0]
	fractionPart := parts[1]

	// 检查小数位数是否超过 decimals
	if len(fractionPart) > int(decimals) {
		return nil, fmt.Errorf("too many decimal places: %d > %d", len(fractionPart), decimals)
	}

	// 补零到 decimals 位
	paddedFraction := fractionPart + strings.Repeat("0", int(decimals)-len(fractionPart))

	// 合并整数和小数部分
	fullAmount := integerPart + paddedFraction

	result, ok := new(big.Int).SetString(fullAmount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}
	return result, nil
}

// TokenAmountFromDecimals converts smallest unit to token amount with specified decimals
func TokenAmountFromDecimals(amount *big.Int, decimals uint8) (*big.Float, error) {
	if amount == nil {
		return nil, fmt.Errorf("amount is nil")
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	result := new(big.Float).SetInt(amount)
	result.Quo(result, new(big.Float).SetInt(divisor))
	return result, nil
}
