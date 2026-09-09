package winmsg

import (
	"slices"
	"testing"
)

func TestGeneratedControlConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value Value
		want  Value
	}{
		{
			name:  "CB_ADDSTRING",
			value: CB_ADDSTRING,
			want:  0x143,
		},
		{
			name:  "CB_ERR",
			value: CB_ERR,
			want:  -1,
		},
		{
			name:  "CBS_DROPDOWNLIST",
			value: CBS_DROPDOWNLIST,
			want:  3,
		},
		{
			name:  "BS_AUTOCHECKBOX",
			value: BS_AUTOCHECKBOX,
			want:  3,
		},
		{
			name:  "BM_SETCHECK",
			value: BM_SETCHECK,
			want:  0xF1,
		},
		{
			name:  "BN_CLICKED",
			value: BN_CLICKED,
			want:  0,
		},
		{
			name:  "BST_CHECKED",
			value: BST_CHECKED,
			want:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.value != test.want {
				t.Fatalf("constant = %d, want %d", test.value, test.want)
			}

			value, ok := Parse(test.name)
			if !ok || value != test.want {
				t.Fatalf("Parse(%q) = %d, %v", test.name, value, ok)
			}

			if !slices.Contains(Names(test.want), test.name) {
				t.Fatalf("Names(%d) is missing %q", test.want, test.name)
			}
		})
	}
}

func TestExpandedAliasesPreserveCanonicalNames(t *testing.T) {
	t.Parallel()

	name, ok := Name(1)
	if !ok || name != "MB_OKCANCEL" {
		t.Fatalf("Name(1) = %q, %v", name, ok)
	}
}
