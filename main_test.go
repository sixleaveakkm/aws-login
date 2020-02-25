package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gopkg.in/ini.v1"
)

func TestMain(m *testing.M) {
	ini.PrettyEqual = true
	ini.PrettyFormat = false
	debugging = true
	setAWSFolderTest()
	cfg = NewConfig()
	os.Exit(m.Run())
}

// In main test, it use real profiles with mfa called from 1password cli
// Require the name of profile used: "TEST_OP_COM"
// and 1pass session: "OP_SESSION_my" both
var OpProfileName = os.Getenv("TEST_OP_COM")

func GetMFA() string {
	mfaCli := exec.Command("op", "get", "totp", OpProfileName)
	var stdErr bytes.Buffer
	mfaCli.Stderr = &stdErr
	mfaBytes, err := mfaCli.Output()
	if err != nil {
		fmt.Printf("%v\n%s\n", err, stdErr.String())
		return ""
	}
	mfa := strings.TrimSpace(string(mfaBytes))
	return mfa
}

func TestWithActualCredential(t *testing.T) {
	if OpProfileName == "" {
		t.Skip("skipping test require 1pass parameters")
	}
	awsFoldPath = "~/.aws"
	t.Run("login mfa", func(t *testing.T) {
		mfa := GetMFA()
		args := []string{"aws-login", "-p", "com", mfa}
		executor(args)
	})

	// wait for 31 seconds to use another mfa code, since one mfa cannot be used twice
	time.Sleep(32 * time.Second)
	t.Run("login role with mfa", func(t *testing.T) {
		mfa := GetMFA()
		args := []string{"aws-login", "-p", "dev-role", mfa}
		executor(args)
	})
	awsFoldPath = debugAwsFolderPath
}

func TestConfigMFA(t *testing.T) {
	args := []string{"aws-login", "config", "mfa", "-p", "dummy", "-n", "arn"}
	executor(args)
}

func TestConfigRoleWithMFA(t *testing.T) {
	awsFoldPath = "./test_resource"
	args := []string{
		"aws-login",
		"config",
		"role",
		"-p", "dummy-with-role",
		"-n", "arn-dummy-mfa",
		"-s", "dummy",
		"-r", "arn-dummy-role",
	}
	executor(args)
}

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
