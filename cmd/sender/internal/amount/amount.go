package amount

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Parse converts a positive decimal amount into an integer number of base units.
func Parse(value string, decimals int) (*big.Int, error) {
	if decimals < 0 {
		return nil, errors.New("decimals must not be negative")
	}
	if value == "" || strings.TrimSpace(value) != value {
		return nil, errors.New("amount must be a plain decimal number")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return nil, errors.New("amount must be a plain decimal number")
	}
	if !digitsOnly(parts[0]) || (len(parts) == 2 && !digitsOnly(parts[1])) {
		return nil, errors.New("amount must be a plain decimal number")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > decimals {
		return nil, fmt.Errorf("amount supports at most %d decimal places", decimals)
	}

	whole := new(big.Int)
	if _, ok := whole.SetString(parts[0], 10); !ok {
		return nil, errors.New("invalid amount")
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	result := new(big.Int).Mul(whole, scale)

	if fraction != "" {
		fraction += strings.Repeat("0", decimals-len(fraction))
		fractionValue := new(big.Int)
		if _, ok := fractionValue.SetString(fraction, 10); !ok {
			return nil, errors.New("invalid amount")
		}
		result.Add(result, fractionValue)
	}
	if result.Sign() <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	return result, nil
}

func Uint64(value *big.Int) (uint64, error) {
	if value == nil || value.Sign() <= 0 || !value.IsUint64() {
		return 0, errors.New("amount is outside the uint64 range")
	}
	return value.Uint64(), nil
}

func Int64(value *big.Int) (int64, error) {
	if value == nil || value.Sign() <= 0 || !value.IsInt64() {
		return 0, errors.New("amount is outside the int64 range")
	}
	return value.Int64(), nil
}

func digitsOnly(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
