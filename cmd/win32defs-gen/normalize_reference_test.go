package main

import (
	"fmt"
	"math/big"
	"strings"
)

// The arbitrary-precision implementation is retained as an independent
// reference for boundary and fuzz tests of the allocation-reduced parser.
func normalizeInteger(value *big.Int, sourceKind string, spec packageSpec) (string, error) {
	if !isNumericKind(sourceKind) || spec.Width == 0 || spec.Width > 64 {
		return "", fmt.Errorf("invalid integer kind or target width: %s/%d", sourceKind, spec.Width)
	}

	sourceBits := sourceWidth(sourceKind)
	sourceSigned := strings.HasPrefix(sourceKind, "int")
	if !integerFits(value, sourceBits, sourceSigned) {
		return "", fmt.Errorf("value %s does not fit source %s", value, sourceKind)
	}

	normalized := new(big.Int).Set(value)

	if spec.Signed {
		if spec.Name == "hresult" {
			if !integerFits(normalized, 32, normalized.Sign() < 0) {
				return "", fmt.Errorf("value %s does not fit HRESULT", value)
			}

			normalized = reinterpretInteger(normalized, sourceWidth(sourceKind), 32, true)
		}

		if !integerFits(normalized, spec.Width, true) {
			return "", fmt.Errorf("value %s does not fit %s", value, spec.Underlying)
		}

		return normalized.String(), nil
	}

	if normalized.Sign() < 0 {
		normalized.Add(normalized, new(big.Int).Lsh(big.NewInt(1), sourceBits))
	}

	if normalized.Sign() < 0 || normalized.BitLen() > int(spec.Width) {
		return "", fmt.Errorf("value %s does not fit %s", value, spec.Underlying)
	}

	width := int((spec.Width + 3) / 4)

	return fmt.Sprintf("0x%0*X", width, normalized), nil
}

func integerFits(value *big.Int, width uint, signed bool) bool {
	if signed {
		limit := new(big.Int).Lsh(big.NewInt(1), width-1)
		return value.Cmp(new(big.Int).Neg(limit)) >= 0 && value.Cmp(limit) < 0
	}

	return value.Sign() >= 0 && value.BitLen() <= int(width)
}

func reinterpretInteger(value *big.Int, fromWidth uint, toWidth uint, signed bool) *big.Int {
	result := new(big.Int).Set(value)
	if result.Sign() < 0 {
		modulus := new(big.Int).Lsh(big.NewInt(1), fromWidth)

		result.Add(result, modulus)
	}

	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), toWidth), big.NewInt(1))
	result.And(result, mask)

	if signed && result.Bit(int(toWidth-1)) == 1 {
		modulus := new(big.Int).Lsh(big.NewInt(1), toWidth)

		result.Sub(result, modulus)
	}

	return result
}
