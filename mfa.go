package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

var MFACommand = &cli.Command{
	Name:         MFA,
	Usage:        "config mfa method",
	Action:       configMFAAction,
	BashComplete: configMFABashComplete,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    SerialNumber,
			Aliases: []string{"n"},
			Usage:   "set mfa device arn display on console",
		},
		&cli.StringFlag{
			Name:    Profile,
			Aliases: []string{"p"},
			Usage:   "profile name to use mfa",
		},
		&cli.Int64Flag{
			Name:    Duration,
			Aliases: []string{"t"},
			Usage:   "mfa duration in seconds, default is 43200(12hours)",
			Value:   DefaultDurationSeconds,
		},
	},
}

// configMFAAction is action function for `aws-login config mfa`
func configMFAAction(c *cli.Context) error {
	config := NewConfig(awsFoldPath)
	profile := c.String("profile")
	if !config.listPossibleProfiles().Contains(ShortSectionName(profile)) {
		return errors.New("input profile is not valid")
	}

	configData, err := config.loadConfig(profile)
	if err != nil {
		// startMFACUI(configData)
		return fmt.Errorf("failed to load profile %s", profile)
	}

	inputDuration := c.Int64(Duration)
	if inputDuration < 900 {
		return fmt.Errorf("duration has minimum value 900 (30 minutes)")
	}
	if inputDuration > 129600 {
		return fmt.Errorf("duration has maximum value 12900 (36 hours)")
	}
	configData.DurationSeconds = inputDuration

	serial := c.String(SerialNumber)
	if serial == "" {
		return fmt.Errorf("serial-number cannot be blank")
	}

	// SerialNumber exists, old mfa profile already set. over write
	if configData.SerialNumber != "" {
		configData.SerialNumber = serial
		config.saveConfig(configData, profile, configFile_)
		return nil
	}
	// SerialNumber doesn't exist, backup credential to "_no_mfa" and save
	// if original profile contains "profile " prefix, no_mfa profile will also has this prefix.
	configData.SerialNumber = serial
	config.backupNoMFACredential(profile, credentialsFile_)
	config.saveConfig(configData, profile, configFile_)
	return nil
}

// func startMFACUI(configData *ConfigDataWithCode) {
// 	fmt.Println("start mfa cui")
// }

// loginForMFA
// <prof> must exists in config, <prof_no_mfa> or <profile prof_no_mfa> exists in credential
func loginForMFA(config *Config, profile string, code string, toDefault bool) error {
	profileNoMFA := fmt.Sprintf("%s%s", profile, excludeConfigPostfix)
	_, err := config.Cred.GetSection(profileNoMFA)
	if err != nil {
		profileNoMFA = fmt.Sprintf("profile %s%s", profile, excludeConfigPostfix)
		_, err = config.Cred.GetSection(profileNoMFA)
		if err != nil {
			return NoProfileError
		}
	}

	// section <profile> must exists
	confSection := config.Conf.Section("profile " + profile)
	var confData ConfigData
	_ = confSection.MapTo(&confData)

	out, err := aws.GetMFASession(&GetMFASessionInput{
		Profile:         profileNoMFA,
		SerialNumber:    confData.SerialNumber,
		DurationSeconds: confData.DurationSeconds,
		Code:            code,
	})
	if err != nil {
		return fmt.Errorf("failed get mfa, %v\n", err)
	}
	cred := &SessionCredential{
		AccessKey:    out.AccessKey,
		SecretKey:    out.SecretKey,
		SessionToken: out.SessionToken,
	}
	config.saveCredential(cred, profile, credentialsFile_)

	if toDefault {
		config.saveConfig(&confData, "default", configFile_)
		config.saveCredential(cred, "default", credentialsFile_)
	}
	return nil
}
