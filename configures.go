package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	. "github.com/deckarep/golang-set"
	"gopkg.in/ini.v1"
)

const excludeConfigPostfix = "_no_mfa"
const configFile_ = "config"
const credentialsFile_ = "credentials"

const debugAwsFolderPath = "./test_resource/"

var awsFoldPath string
var debugging bool

type CredentialData struct {
	AccessKey    string `ini:"aws_access_key_id,omitempty"`
	SecretKey    string `ini:"aws_secret_access_key,omitempty"`
	SessionToken string `ini:"aws_session_token,omitempty"`
}

type ConfigData struct {
	Region          string `ini:"region,omitempty"`
	Output          string `ini:"output,omitempty"`
	SerialNumber    string `ini:"mfa_serial,omitempty"`
	DurationSeconds int64  `ini:"duration,omitempty"`
	OriginProfile   string `ini:"source_profile,omitempty"`
	AssumeRoleArn   string `ini:"role_arn,omitempty"`
}

var NoProfileError = errors.New("profile not found")

type ConfigDataWithCode struct {
	ConfigData
	Profile string `ini:"-"`
	Code    string `ini:"-"`
}

type Config struct {
	Conf *ini.File
	Cred *ini.File
}

func NewConfig() (c *Config) {
	c = &Config{}
	cfg, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, filepath.Join(awsFoldPath, configFile_))
	if err != nil {
		fmt.Printf("Fail to read file, %v", err)
		os.Exit(1)
	}
	c.Conf = cfg

	cred, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, filepath.Join(awsFoldPath, credentialsFile_))
	if err != nil {
		fmt.Printf("Fail to read file, %v", err)
		os.Exit(1)
	}
	c.Cred = cred
	return c
}

func (c *Config) loadProfileData(profile string) (*ConfigDataWithCode, error) {
	section, err := c.loadSection(profile, c.Conf)
	if err != nil {
		return nil, err
	}
	var configData ConfigData
	err = section.MapTo(&configData)
	return &ConfigDataWithCode{
		ConfigData: configData,
		Profile:    profile,
	}, err
}

// loadSection loads profile file in order
// <profile> is the name of profile.
// "<profile>_no_mfa" => "profile <profile>_no_mfa" => "<profile>" => "profile <profile>"
// return the first one found, if none exists, return NoProfileError
func (c *Config) loadSection(profile string, from *ini.File) (*ini.Section, error) {
	list := []string{
		fmt.Sprintf("%s_no_mfa", profile),
		fmt.Sprintf("profile %s_no_mfa", profile),
		fmt.Sprintf("%s", profile),
		fmt.Sprintf("profile %s", profile),
	}
	for i := 0; i < len(list); i++ {
		section, err := from.GetSection(list[i])
		if err == nil {
			return section, nil
		}
	}
	return nil, NoProfileError
}

// loadCredential read credentialData from profile with fallback
// <profile> is the name of profile.
// "<profile>_no_mfa" => "profile <profile>_no_mfa" => "<profile>" => "profile <profile>"
// return the first one found, if none exists, return NoProfileError
func (c *Config) loadCredential(profile string) (*CredentialData, error) {
	section, err := c.loadSection(profile, c.Cred)
	if err != nil {
		return nil, err
	}
	var cred CredentialData
	err = section.MapTo(&cred)
	return &cred, err
}

// readSpecificCredential read specific credentialData without fallback policy
func (c *Config) readSpecificCredential(profile string) (*CredentialData, error) {
	section, err := c.Cred.GetSection(profile)
	if err != nil {
		return nil, err
	}
	var cred CredentialData
	err = section.MapTo(&cred)
	return &cred, err
}

// loadMFAProfile read the specific mfa profile without fallback
func (c *Config) loadMFAProfile(profile string) ProfileWithCode {
	var cfgData ConfigData
	var credData CredentialData
	_ = c.Conf.Section(profile).MapTo(&cfgData)
	_ = c.Cred.Section(profile).MapTo(&credData)
	ps := &ProfileWithCode{
		ProfileName:     profile,
		DurationSeconds: nil,
		SerialNumber:    nil,
	}
	return *ps
}

func (c *Config) saveConfig(conf *ConfigDataWithCode, configFile string) {
	var confData ConfigData
	confData = conf.ConfigData
	err := c.Conf.Section(conf.Profile).ReflectFrom(&confData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_ = c.Conf.SaveTo(filepath.Join(awsFoldPath, configFile))
}

// saveCredential
func (c *Config) saveCredential(cred *CredentialData, profile string) {
	err := c.Cred.Section(profile).ReflectFrom(&cred)
	if err != nil {
		fmt.Printf("Failed when save credential")
		os.Exit(1)
	}
	c.saveToFiles()
}

func (c *Config) backupProfile(profile string, credFile string) {
	var credData CredentialData
	_ = c.Cred.Section(profile).MapTo(&credData)
	if credData.SessionToken == "" {
		_ = c.Cred.Section(fmt.Sprintf("%s%s", profile, excludeConfigPostfix)).ReflectFrom(&credData)
		_ = c.Cred.SaveTo(filepath.Join(awsFoldPath, credFile))
	}
}

func (c *Config) listAWSProfiles() Set {
	if awsFoldPath == "" {
		return nil
	}
	profiles := NewSet()
	var sectionList = c.Conf.SectionStrings()
	sectionList = append(sectionList, c.Cred.SectionStrings()...)
	for _, profile := range sectionList {
		if strings.HasSuffix(profile, excludeConfigPostfix) || profile == ini.DefaultSection {
			continue
		}
		profiles.Add(profile)
	}
	return profiles
}

func setAWSFolderDefault() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	awsFoldPath = filepath.Join(dir, ".aws")
}

func (c Config) saveToFiles() {
	originalPath := awsFoldPath
	if debugging {
		awsFoldPath = "./test_resource/output/"
	}
	_ = c.Conf.SaveTo(filepath.Join(awsFoldPath, configFile_))
	_ = c.Cred.SaveTo(filepath.Join(awsFoldPath, credentialsFile_))
	if debugging {
		awsFoldPath = originalPath
	}
}
