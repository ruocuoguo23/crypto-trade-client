package amount

import (
	"math/big"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		decimals int
		want     string
		wantErr  bool
	}{
		{name: "whole", value: "2", decimals: 8, want: "200000000"},
		{name: "fraction", value: "0.00000001", decimals: 8, want: "1"},
		{name: "pads fraction", value: "1.5", decimals: 6, want: "1500000"},
		{name: "zero", value: "0", decimals: 8, wantErr: true},
		{name: "negative", value: "-1", decimals: 8, wantErr: true},
		{name: "exponent", value: "1e3", decimals: 8, wantErr: true},
		{name: "too precise", value: "0.000000001", decimals: 8, wantErr: true},
		{name: "missing whole", value: ".1", decimals: 8, wantErr: true},
		{name: "missing fraction", value: "1.", decimals: 8, wantErr: true},
		{name: "spaces", value: " 1", decimals: 8, wantErr: true},
		{name: "letters", value: "one", decimals: 8, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.value, test.decimals)
			if test.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) succeeded, want error", test.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q): %v", test.value, err)
			}
			if got.String() != test.want {
				t.Fatalf("Parse(%q) = %s, want %s", test.value, got, test.want)
			}
		})
	}
}

func TestIntegerRanges(t *testing.T) {
	tooLarge := new(big.Int).Add(new(big.Int).SetUint64(^uint64(0)), big.NewInt(1))
	if _, err := Uint64(tooLarge); err == nil {
		t.Fatal("Uint64 accepted an overflowing value")
	}
	if _, err := Int64(new(big.Int).SetUint64(^uint64(0))); err == nil {
		t.Fatal("Int64 accepted an overflowing value")
	}
}
