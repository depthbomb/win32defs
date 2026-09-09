package main

import (
	"math/big"
	"testing"
)

func TestIntegerBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		kind    string
		pkg     string
		want    string
		wantErr bool
	}{
		{
			value: "4294967295",
			kind:  "uint32",
			pkg:   "shell",
			want:  "0xFFFFFFFF",
		},
		{
			value:   "4294967296",
			kind:    "uint64",
			pkg:     "shell",
			wantErr: true,
		},
		{
			value:   "65536",
			kind:    "uint32",
			pkg:     "facility",
			wantErr: true,
		},
		{
			value: "-1",
			kind:  "int32",
			pkg:   "shell",
			want:  "0xFFFFFFFF",
		},
		{
			value: "-1",
			kind:  "int32",
			pkg:   "memory",
			want:  "0x00000000FFFFFFFF",
		},
		{
			value: "2147500037",
			kind:  "uint32",
			pkg:   "hresult",
			want:  "-2147467259",
		},
		{
			value:   "4294967296",
			kind:    "uint64",
			pkg:     "hresult",
			wantErr: true,
		},
		{
			value:   "-1",
			kind:    "uint32",
			pkg:     "shell",
			wantErr: true,
		},
		{
			value:   "256",
			kind:    "uint8",
			pkg:     "winmsg",
			wantErr: true,
		},
		{
			value:   "9223372036854775808",
			kind:    "uint64",
			pkg:     "winmsg",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.pkg+"/"+test.kind+"/"+test.value, func(t *testing.T) {
			t.Parallel()

			value, _ := new(big.Int).SetString(test.value, 10)
			got, err := normalizeInteger(value, test.kind, specByName(test.pkg))
			if (err != nil) != test.wantErr || !test.wantErr && got != test.want {
				t.Fatalf("normalize = %q, %v; want %q, error=%v", got, err, test.want, test.wantErr)
			}
		})
	}
}

func TestFloatConstants(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"float32", "float64"} {
		for _, value := range []string{"0.25", "1", "1e-20"} {
			item := metadataConstant{
				Name:  "FLOAT_TEST",
				Kind:  kind,
				Value: value,
			}
			constant, err := normalizeConstant(item, specByName("gdi"))
			if err != nil || constant.Type != kind || constant.Numeric {
				t.Fatalf("float %s/%s = %#v, %v", kind, value, constant, err)
			}

			if _, err := renderConstants(specByName("gdi"), []generatedConstant{constant}, nil, sourceLock{}); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, value := range []string{"NaN", "+Inf", "-Inf", "1e1000"} {
		item := metadataConstant{
			Kind:  "float32",
			Value: value,
		}
		if _, err := normalizeConstant(item, specByName("gdi")); err == nil {
			t.Fatalf("accepted non-finite float %q", value)
		}
	}
}
