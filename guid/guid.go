// Package guid provides Windows-compatible GUID values and generated interface
// and type identifiers.
package guid

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// GUID is the in-memory field representation of a Windows GUID.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// Parse converts the canonical textual representation of a GUID into a value.
// Braces around the value are accepted.
func Parse(value string) (GUID, error) {
	cleaned := strings.TrimSpace(value)
	cleaned = strings.TrimPrefix(cleaned, "{")
	cleaned = strings.TrimSuffix(cleaned, "}")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	if len(cleaned) != 32 {
		return GUID{}, fmt.Errorf("guid: invalid length %d", len(cleaned))
	}

	bytes := make([]byte, 16)
	if _, err := hex.Decode(bytes, []byte(cleaned)); err != nil {
		return GUID{}, fmt.Errorf("guid: decode: %w", err)
	}

	result := GUID{
		Data1: uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3]),
		Data2: uint16(bytes[4])<<8 | uint16(bytes[5]),
		Data3: uint16(bytes[6])<<8 | uint16(bytes[7]),
	}
	copy(result.Data4[:], bytes[8:])

	return result, nil
}

// MustParse is like Parse but panics when value is invalid.
func MustParse(value string) GUID {
	result, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return result
}

// String returns the canonical lowercase textual representation of value.
func (value GUID) String() string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		value.Data1,
		value.Data2,
		value.Data3,
		value.Data4[0],
		value.Data4[1],
		value.Data4[2],
		value.Data4[3],
		value.Data4[4],
		value.Data4[5],
		value.Data4[6],
		value.Data4[7],
	)
}

// MarshalText implements encoding.TextMarshaler.
func (value GUID) MarshalText() ([]byte, error) {
	return []byte(value.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (value *GUID) UnmarshalText(text []byte) error {
	if value == nil {
		return errors.New("guid: UnmarshalText on nil pointer")
	}

	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*value = parsed

	return nil
}
