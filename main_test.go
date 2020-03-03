package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

func setAWSFolderTest() {
	awsFoldPath = "./test_resource"
}

func TestMain(m *testing.M) {
	ini.PrettyEqual = true
	ini.PrettyFormat = false
	debugging = true
	setAWSFolderTest()
	os.Exit(m.Run())
}

func TestConfigMFA(t *testing.T) {
	// test config mfa success
	args := []string{"aws-login", "config", "mfa", "-p", "user-profile", "-n", "arn"}
	executor(args)
	outConfig := NewConfig(filepath.Join(debugAwsFolderPath, "output"))
	fmt.Println(outConfig.Conf.Section("user-profile").Key("mfa_serial").String())
	assert.Equal(t, "arn", outConfig.Conf.Section("user-profile").Key("mfa_serial").String())
	assert.Equal(t, "43200", outConfig.Conf.Section("user-profile").Key("duration").String())
	assert.Equal(t, "DEFAULT_KEY_ID", outConfig.Cred.Section("user-profile_no_mfa").Key("aws_access_key_id").String())

	// test config mfa override
	args = []string{"aws-login", "config", "mfa", "-p", "profile-exist", "-n", "arn:another", "-t", "30000"}
	executor(args)
	outConfig = NewConfig(filepath.Join(debugAwsFolderPath, "output"))
	assert.Equal(t, "arn:another", outConfig.Conf.Section("profile-exist").Key("mfa_serial").String())
	assert.Equal(t, "30000", outConfig.Conf.Section("profile-exist").Key("duration").String())
}

func TestMFALogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockAWS(ctrl)
	m.EXPECT().GetMFASession(gomock.Any()).Return(&SessionCredential{
		AccessKey:    "MFA_KEY_ID",
		SecretKey:    "MFA_SECRET",
		SessionToken: "MFA_SESSION_TOKEN",
	}, nil)
	aws = m

	args := []string{"aws-login", "-p", "dummy", "-d", "123456"}
	executor(args)
	outConfig := NewConfig(filepath.Join(debugAwsFolderPath, "output"))
	assert.Equal(t, "MFA_KEY_ID",
		outConfig.Cred.Section("dummy").Key("aws_access_key_id").String())
	assert.Equal(t, "MFA_SESSION_TOKEN",
		outConfig.Cred.Section("dummy").Key("aws_session_token").String())

	assert.Equal(t, "MFA_KEY_ID",
		outConfig.Cred.Section("default").Key("aws_access_key_id").String())
	assert.Equal(t, "MFA_SESSION_TOKEN",
		outConfig.Cred.Section("default").Key("aws_session_token").String())
}

func TestConfigRole(t *testing.T) {
	// test config mfa success
	args := []string{"aws-login", "config", "role", "-p", "user-role", "-s", "user-profile", "-n", "arn", "-r", "arn:dummy-role"}
	executor(args)
	outConfig := NewConfig(filepath.Join(debugAwsFolderPath, "output"))
	assert.Equal(t, "arn", outConfig.Conf.Section("user-role").Key("mfa_serial").String())
	assert.Equal(t, "43200", outConfig.Conf.Section("user-role").Key("duration").String())
	assert.Equal(t, "arn:dummy-role", outConfig.Conf.Section("user-role").Key("role_arn").String())
	assert.Equal(t, "arn", outConfig.Conf.Section("user-role").Key("mfa_serial").String())
}

func TestRoleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockAWS(ctrl)
	m.EXPECT().GetAssumeRoleSession(gomock.Any()).Return(&SessionCredential{
		AccessKey:    "MFA_KEY_ID",
		SecretKey:    "MFA_SECRET",
		SessionToken: "MFA_SESSION_TOKEN",
	}, nil)
	aws = m

	args := []string{"aws-login", "-p", "user-role-2", "-d", "123456"}
	executor(args)
	outConfig := NewConfig(filepath.Join(debugAwsFolderPath, "output"))
	assert.Equal(t, "MFA_KEY_ID",
		outConfig.Cred.Section("user-role-2").Key("aws_access_key_id").String())
	assert.Equal(t, "MFA_SESSION_TOKEN",
		outConfig.Cred.Section("user-role-2").Key("aws_session_token").String())

	assert.Equal(t, "MFA_KEY_ID",
		outConfig.Cred.Section("default").Key("aws_access_key_id").String())
	assert.Equal(t, "MFA_SESSION_TOKEN",
		outConfig.Cred.Section("default").Key("aws_session_token").String())
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
