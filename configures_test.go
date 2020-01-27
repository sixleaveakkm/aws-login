package main

import (
	"errors"
	"path/filepath"
	"testing"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

var cfg *Config

func Test_listProfiles(t *testing.T) {
	profiles := cfg.listAWSProfiles()
	assert.Equal(t, mapset.NewSet("dummy", "default"), profiles)
}

func setAWSFolderTest() {
	awsFoldPath = debugAwsFolderPath
}

func TestConfig_loadProfile(t *testing.T) {
	cfg.loadMFAProfile("dummy")

	_ = cfg.Conf.SaveTo(filepath.Join(awsFoldPath, "test_config"))
}

func TestConfig_loadProfileData_notExist(t *testing.T) {
	_, err := cfg.loadProfileData("not_exist")
	assert.Error(t, err)
	assert.Equal(t, errors.Is(err, NoProfileError), true)
}

func TestConfig_saveConfig(t *testing.T) {
	conf := &ConfigDataWithCode{
		ConfigData: ConfigData{
			SerialNumber:    "arn",
			DurationSeconds: 43200,
		},
		Profile: "test",
		Code:    "111111",
	}
	cfg.saveConfig(conf, "test_config")
}

func TestNewConfig_backupProfile(t *testing.T) {
	cfg.backupProfile("dummy", "test_credentials")
}
