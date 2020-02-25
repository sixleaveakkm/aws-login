package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
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

func configMFAAction(c *cli.Context) error {
	config := NewConfig()
	profile := c.String("profile")

	configData, err := config.loadProfileData(profile)

	if err != nil {
		// startMFACUI(configData)
		os.Exit(1)
		return nil
	}
	if configData.DurationSeconds == 0 || c.Int64(Duration) != DefaultDurationSeconds {
		configData.DurationSeconds = c.Int64(Duration)
	}
	serial := c.String(SerialNumber)
	if serial == "" {
		// startMFACUI(configData)
		os.Exit(1)
	} else {
		configData.SerialNumber = serial
		config.backupProfile(profile, credentialsFile_)
		config.saveConfig(configData, configFile_)
	}
	return nil
}

// func startMFACUI(configData *ConfigDataWithCode) {
// 	fmt.Println("start mfa cui")
// }

func getMFASession(input *ConfigDataWithCode) (*CredentialData, error) {
	svc := getSession(input.Profile)
	output, err := svc.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(input.DurationSeconds),
		SerialNumber:    aws.String(input.SerialNumber),
		TokenCode:       aws.String(input.Code),
	})
	if err != nil {
		return nil, err
	}
	return &CredentialData{
		AccessKey:    *output.Credentials.AccessKeyId,
		SecretKey:    *output.Credentials.SecretAccessKey,
		SessionToken: *output.Credentials.SessionToken,
	}, nil
}

// loginForMFA
// <prof> must exists in config, <prof_no_mfa> or <profile prof_no_mfa> exists in credential
func loginForMFA(config *Config, profile string, code string) error {
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
	confSection := config.Conf.Section(profile)
	var confData ConfigData
	_ = confSection.MapTo(&confData)

	input := &ConfigDataWithCode{
		ConfigData: ConfigData{
			SerialNumber:    confData.SerialNumber,
			DurationSeconds: confData.DurationSeconds,
		},
		Profile: profileNoMFA,
		Code:    code,
	}

	var out *CredentialData
	out, err = getMFASession(input)
	if err != nil {
		return fmt.Errorf("failed get mfa, %v\n", err)
	}
	config.saveCredential(out, profile)
	return nil
}
