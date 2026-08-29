//go:build !windows

package message

import (
	"errors"
	"testing"
)

func TestFormatWithOptionsUnsupported(t *testing.T) {
	t.Parallel()

	_, err := FormatWithOptions(5, Options{SearchSystem: true})
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("FormatWithOptions error = %v", err)
	}
}
