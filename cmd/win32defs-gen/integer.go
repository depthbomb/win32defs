package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func normalizeIntegerText(text string, kind string, spec packageSpec) (string, error) {
	if !isNumericKind(kind) || spec.Width == 0 || spec.Width > 64 {
		return "", fmt.Errorf("invalid integer kind or target width: %s/%d", kind, spec.Width)
	}

	width := sourceWidth(kind)
	var value uint64
	var signedValue int64
	negative := false
	if strings.HasPrefix(kind, "int") {
		parsed, err := strconv.ParseInt(text, 10, int(width))
		if err != nil {
			return "", fmt.Errorf("value %q does not fit source %s: %w", text, kind, err)
		}

		signedValue = parsed
		negative = parsed < 0
		value = uint64(parsed)
		if negative && width < 64 {
			value &= (uint64(1) << width) - 1
		}
	} else {
		parsed, err := strconv.ParseUint(strings.TrimPrefix(text, "+"), 10, int(width))
		if err != nil {
			return "", fmt.Errorf("value %q does not fit source %s: %w", text, kind, err)
		}

		value = parsed
		signedValue = int64(parsed)
	}

	if spec.Signed {
		if spec.Name == "hresult" {
			if negative && signedValue < math.MinInt32 || !negative && value > math.MaxUint32 {
				return "", fmt.Errorf("value %q does not fit HRESULT", text)
			}

			return strconv.FormatInt(int64(int32(value)), 10), nil
		}

		if !negative && value > math.MaxInt64 {
			return "", fmt.Errorf("value %q does not fit %s", text, spec.Underlying)
		}

		if spec.Width < 64 {
			limit := int64(1) << (spec.Width - 1)
			if signedValue < -limit || signedValue >= limit {
				return "", fmt.Errorf("value %q does not fit %s", text, spec.Underlying)
			}
		}

		return strconv.FormatInt(signedValue, 10), nil
	}

	if spec.Width < 64 && value>>spec.Width != 0 {
		return "", fmt.Errorf("value %q does not fit %s", text, spec.Underlying)
	}

	var buffer [18]byte
	digits := int((spec.Width + 3) / 4)
	buffer[0], buffer[1] = '0', 'x'
	for index := digits + 1; index >= 2; index-- {
		buffer[index] = "0123456789ABCDEF"[value&15]
		value >>= 4
	}

	return string(buffer[:digits+2]), nil
}
