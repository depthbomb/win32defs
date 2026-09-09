package shell

import (
	"slices"
	"testing"
)

func TestGeneratedShellConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value Value
		want  Value
	}{
		{
			name:  "NIM_ADD",
			value: NIM_ADD,
			want:  0,
		},
		{
			name:  "NIF_ICON",
			value: NIF_ICON,
			want:  2,
		},
		{
			name:  "FOS_PICKFOLDERS",
			value: FOS_PICKFOLDERS,
			want:  0x20,
		},
		{
			name:  "FOS_SUPPORTSTREAMABLEITEMS",
			value: FOS_SUPPORTSTREAMABLEITEMS,
			want:  0x80000000,
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
