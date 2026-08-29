// Package propertykey provides generated Windows property keys.
package propertykey

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/depthbomb/win32defs/guid"
)

// Key identifies a property by format identifier and numeric property ID.
type Key struct {
	FormatID   guid.GUID
	PropertyID uint32
}

// Parse converts a "GUID property-id" representation into a key.
func Parse(value string) (Key, error) {
	guidText, propertyText, found := strings.Cut(strings.TrimSpace(value), " ")
	if !found {
		return Key{}, fmt.Errorf("propertykey: missing property ID")
	}

	formatID, err := guid.Parse(guidText)
	if err != nil {
		return Key{}, err
	}

	propertyID, err := strconv.ParseUint(strings.TrimSpace(propertyText), 10, 32)
	if err != nil {
		return Key{}, fmt.Errorf("propertykey: invalid property ID: %w", err)
	}

	return Key{FormatID: formatID, PropertyID: uint32(propertyID)}, nil
}

// String returns a stable "GUID property-id" representation.
func (key Key) String() string {
	return key.FormatID.String() + " " + strconv.FormatUint(uint64(key.PropertyID), 10)
}
