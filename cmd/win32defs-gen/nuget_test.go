package main

import "testing"

func TestCompareNumericVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		left  string
		right string
		want  int
	}{
		{left: "71.0.20-preview", right: "70.0.11-preview", want: 1},
		{left: "10.3.16-preview", right: "10.2.185-preview", want: 1},
		{left: "1.2.0", right: "1.2", want: 0},
	}

	for _, test := range tests {
		t.Run(test.left+"_"+test.right, func(t *testing.T) {
			t.Parallel()

			got := compareNumericVersions(parseNumericVersion(test.left), parseNumericVersion(test.right))
			if test.want == 0 && got != 0 {
				t.Fatalf("compare = %d, want 0", got)
			}

			if test.want < 0 && got >= 0 {
				t.Fatalf("compare = %d, want negative", got)
			}

			if test.want > 0 && got <= 0 {
				t.Fatalf("compare = %d, want positive", got)
			}
		})
	}
}
