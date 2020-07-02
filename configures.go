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

// configure read configure files
// - aws-login config mfa --profile <>
//   list profiles doesn't end with "_no_mfa"
// - aws-login config role --source-profile <>
//   list profiles doesn't end with "_no_mfa"
// - aws-login --profile <>
//   list profiles with serial_number attached

const excludeConfigPostfix = "_no_mfa"
const configFile_ = "config"
const credentialsFile_ = "credentials"

const debugAwsFolderPath = "./test_resource/"

var awsFoldPath string
var debugging bool

type SessionCredential struct {
	AccessKey    string `ini:"aws_access_key_id,omitempty"`
	SecretKey    string `ini:"aws_secret_access_key,omitempty"`
	SessionToken string `ini:"aws_session_token,omitempty"`
}

type ConfigData struct {
	Region string `ini:"region,omitempty"`
	Output string `ini:"output,omitempty"`

	SerialNumber    string `ini:"mfa_serial,omitempty"`
	DurationSeconds int64  `ini:"duration,omitempty"`
	SourceProfile   string `ini:"source_profile,omitempty"`
	AssumeRoleArn   string `ini:"role_arn,omitempty"`
}

var NoProfileError = errors.New("profile not found")

type Config struct {
	Conf *ini.File
	Cred *ini.File
}

func NewConfig(folder string) (c *Config) {
	c = &Config{}
	cfg, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, filepath.Join(folder, configFile_))
	if err != nil {
		fmt.Printf("Fail to read file, %v", err)
		os.Exit(1)
	}
	c.Conf = cfg

	cred, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
	}, filepath.Join(folder, credentialsFile_))
	if err != nil {
		fmt.Printf("Fail to read file, %v", err)
		os.Exit(1)
	}
	c.Cred = cred

	return c
}

// ShortSectionName get profile name without "profile " prefix.
func ShortSectionName(name string) string {
	s := strings.Split(name, "profile ")
	return s[len(s)-1]
}

// listMFAProfiles list profiles with serial_number attached.
// It is used for `aws-login -p ` completion.
func (c Config) listMFAProfiles() (results map[string]string) {
	results = make(map[string]string)
	confSections := c.Conf.Sections()
	for _, section := range confSections {
		name := ShortSectionName(section.Name())
		_, err := section.GetKey(SerialNumberInFile)
		if err == nil {
			if source, e := section.GetKey(SourceProfile); e == nil {
				// contains "source-profile", is role with mfa
				results[name] = fmt.Sprintf("assume role from '%s'", source.String())
			} else {
				// no "source-profile", is mfa
				results[name] = fmt.Sprintf("login '%s' with mfa", section.Name())
			}
		}
	}
	return results
}

// listPossibleProfiles list possible profiles to config mfa.
// Exclude profiles with suffix "_no_mfa"
func (c *Config) listPossibleProfiles() Set {
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
		profiles.Add(ShortSectionName(profile))
	}
	return profiles
}

// backupNoMFACredential get current credential and save to "_no_mfa".
// no mfa profile matches original profile name
func (c *Config) backupNoMFACredential(profile string, credFile string) {
	var credData SessionCredential
	section, err := c.getNoMFACredential(profile)
	if err != nil {
		return
	}
	_ = section.MapTo(&credData)
	if credData.SessionToken == "" {
		name := fmt.Sprintf("%s%s", profile, excludeConfigPostfix)
		c.saveCredential(&credData, name, credFile)
	}
}

// getNoMFACredential get credential profile is not mfa
func (c *Config) getNoMFACredential(profile string) (*ini.Section, error) {
	list := []string{
		fmt.Sprintf("%s_no_mfa", profile),
		fmt.Sprintf("%s", profile),
	}
	for i := 0; i < len(list); i++ {
		section, err := c.Cred.GetSection(list[i])
		if err == nil {
			return section, nil
		}
	}
	return nil, NoProfileError
}

func (c *Config) loadConfig(profile string) (*ConfigData, error) {
	section, err := c.loadSection(profile, c.Conf)
	if err != nil {
		return nil, err
	}
	var conf ConfigData
	err = section.MapTo(&conf)
	return &conf, err
}

func (c *Config) saveConfig(conf *ConfigData, profile string, configFile string) {
	err := c.Conf.Section(profile).ReflectFrom(&conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	originalPath := awsFoldPath
	if debugging {
		awsFoldPath = "./test_resource/output/"
	}
	_ = c.Conf.SaveTo(filepath.Join(awsFoldPath, configFile))
	if debugging {
		awsFoldPath = originalPath
	}
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
func (c *Config) loadCredential(profile string) (*SessionCredential, error) {
	section, err := c.loadSection(profile, c.Cred)
	if err != nil {
		return nil, err
	}
	var cred SessionCredential
	err = section.MapTo(&cred)
	return &cred, err
}

// saveCredential
func (c *Config) saveCredential(cred *SessionCredential, profile string, credFile string) {
	err := c.Cred.Section(profile).ReflectFrom(&cred)
	if err != nil {
		fmt.Printf("Failed when save credential")
		os.Exit(1)
	}

	originalPath := awsFoldPath
	if debugging {
		awsFoldPath = "./test_resource/output/"
	}
	_ = c.Cred.SaveTo(filepath.Join(awsFoldPath, credFile))
	if debugging {
		awsFoldPath = originalPath
	}
}

// setAWSFolderDefault set aws configure files' default folder
func setAWSFolderDefault() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	awsFoldPath = filepath.Join(dir, ".aws")
}
