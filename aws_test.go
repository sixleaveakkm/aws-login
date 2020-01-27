package main

import (
	"os"
	"testing"
)

func Test_getAccountId(t *testing.T) {
	type args struct {
		profile string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test com",
			args{profile: "com"},
			os.Getenv("COM_ACCOUNT_ID"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAccountId(tt.args.profile); got != tt.want {
				t.Errorf("getAccountId() = %v, want %v", got, tt.want)
			}
		})
	}
}
