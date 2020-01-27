package main

import "testing"

func Test_isSixDigit(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			"match",
			"111111",
			true,
		},
		{
			"match",
			"123456",
			true,
		},
		{
			"not match",
			"12345",
			false,
		},
		{
			"not match",
			"1234567",
			false,
		},
		{
			"not match",
			"assdfe",
			false,
		},
		{
			"not match",
			"asss",
			false,
		},
		{
			"not match",
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSixDigit(tt.code); got != tt.want {
				t.Errorf("isSixDigit(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
