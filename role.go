package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

var RoleCommand = &cli.Command{
	Name:         Role,
	Usage:        "config role method",
	Action:       configRoleAction,
	BashComplete: configRoleBashComplete,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    SerialNumber,
			Aliases: []string{"n"},
			Usage:   "set mfa device arn display on console",
		},
		&cli.StringFlag{
			Name:    Profile,
			Aliases: []string{"p"},
			Usage:   "new profile name for assume role",
		},
		&cli.StringFlag{
			Name:    SourceProfile,
			Aliases: []string{"s"},
			Usage:   "source profile used to assume role",
		},
		&cli.StringFlag{
			Name:    RoleArn,
			Aliases: []string{"r"},
			Usage:   "role arn in destination account used to assume role",
		},
		&cli.Int64Flag{
			Name:    Duration,
			Aliases: []string{"t"},
			Usage:   "mfa duration in seconds, default is 43200(12hours)",
			Value:   DefaultDurationSeconds,
		},
		&cli.BoolFlag{
			Name:  NoMFA,
			Usage: "explicitly indicate assume role without mfa, will confirm for mfa if no mfa provided when this flag is present",
		},
	},
}

// func startRoleCUI(configData *ConfigDataWithCode) {
// 	fmt.Println("start role cui")
// }

func configRoleAction(c *cli.Context) error {
	config := NewConfig(awsFoldPath)
	profile := getProfile(c)
	sourceProfile := c.String(SourceProfile)
	if !config.listPossibleProfiles().Contains(ShortSectionName(sourceProfile)) {
		return errors.New("input profile is not valid")
	}

	configData := &ConfigData{
		SerialNumber:    c.String(SerialNumber),
		DurationSeconds: c.Int64(Duration),
		SourceProfile:   c.String(SourceProfile),
		AssumeRoleArn:   c.String(RoleArn),
	}

	if configData.DurationSeconds == 0 || c.Int64(Duration) != DefaultDurationSeconds {
		configData.DurationSeconds = c.Int64(Duration)
	}

	// Check original profile, if original profile contains token (one time),
	// maximum duration is 1 hour. start gui to confirm
	originProfile, err := config.loadCredential(configData.SourceProfile)
	if err != nil {

		fmt.Printf("source profile: (%s) doesn't exists", configData.SourceProfile)
		os.Exit(1)
		// startRoleCUI(configData)
		// return nil
	}
	if originProfile.SessionToken != "" {
		fmt.Printf("source profile is a temporary profile with maximum 1 hour")
		configData.DurationSeconds = 3600
	}

	serial := c.String(SerialNumber)
	if serial == "" {
		// startRoleCUI(configData)
		os.Exit(1)
	} else {
		configData.SerialNumber = serial
		config.saveConfig(configData, profile, configFile_)
	}
	return nil
}

func getRoleSession(input *GetAssumeRoleRoleInput) (*SessionCredential, error) {
	if input.SourceProfile == "" {
		return nil, fmt.Errorf("'origin_profile' is not present in %s", input.SourceProfile)
	}
	return aws.GetAssumeRoleSession(input)
}

func loginForRole(config *Config, profile string, code string, toDefault bool) error {
	// section <profile> must exists
	confSection, err := config.Conf.GetSection(Profile + " " + profile)
	if err != nil {
		return NoProfileError
	}
	var confData ConfigData
	_ = confSection.MapTo(&confData)

	sProfile := confData.SourceProfile

	var cred *ini.Section
	cred, err = config.Cred.GetSection(sProfile)
	if err != nil {
		sProfile = fmt.Sprintf("profile %s%s", profile, excludeConfigPostfix)
		cred, err = config.Cred.GetSection(sProfile)
		if err != nil {
			return NoProfileError
		}
	}
	if _, err = cred.GetKey("aws_session_token"); err == nil {
		// session has token, get no_mfa profile
		sProfile = fmt.Sprintf("%s%s", confData.SourceProfile, excludeConfigPostfix)
		_, err = config.Cred.GetSection(sProfile)
		if err != nil {
			sProfile = fmt.Sprintf("profile %s%s", profile, excludeConfigPostfix)
			_, err = config.Cred.GetSection(sProfile)
			if err != nil {
				return NoProfileError
			}
		}
	}

	confData.SourceProfile = sProfile
	input := &GetAssumeRoleRoleInput{
		SourceProfile:   confData.SourceProfile,
		AssumeRoleArn:   confData.AssumeRoleArn,
		SerialNumber:    confData.SerialNumber,
		DurationSeconds: confData.DurationSeconds,
		Code:            code,
	}

	var out *SessionCredential
	out, err = getRoleSession(input)
	if err != nil {
		return fmt.Errorf("failed get mfa, %v\n", err)
	}
	config.saveCredential(out, profile, credentialsFile_)

	if toDefault {
		config.saveConfig(&confData, "default", configFile_)
		config.saveCredential(out, "default", credentialsFile_)
	}
	return nil
}
