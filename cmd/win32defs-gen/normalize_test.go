package main

import (
	"math/big"
	"strconv"
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

			got, err = normalizeIntegerText(test.value, test.kind, specByName(test.pkg))
			if (err != nil) != test.wantErr || !test.wantErr && got != test.want {
				t.Fatalf("fast normalize = %q, %v; want %q, error=%v", got, err, test.want, test.wantErr)
			}
		})
	}
}

func FuzzIntegerNormalization(f *testing.F) {
	f.Add(int64(-1), false, uint8(4), uint8(1))
	f.Add(int64(4294967296), true, uint8(7), uint8(1))
	f.Add(int64(-9223372036854775808), false, uint8(6), uint8(3))
	f.Fuzz(func(t *testing.T, number int64, unsigned bool, kindIndex uint8, packageIndex uint8) {
		kinds := [...]string{"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "char"}
		packages := [...]string{"facility", "shell", "hresult", "winmsg", "memory"}
		text := strconv.FormatInt(number, 10)
		if unsigned {
			text = strconv.FormatUint(uint64(number), 10)
		}

		value, _ := new(big.Int).SetString(text, 10)
		kind := kinds[int(kindIndex)%len(kinds)]
		spec := specByName(packages[int(packageIndex)%len(packages)])
		want, wantErr := normalizeInteger(value, kind, spec)
		got, err := normalizeIntegerText(text, kind, spec)
		if (err != nil) != (wantErr != nil) || err == nil && got != want {
			t.Fatalf("%s/%s/%s: fast = %q, %v; reference = %q, %v", text, kind, spec.Name, got, err, want, wantErr)
		}
	})
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
