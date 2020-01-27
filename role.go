package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/urfave/cli/v2"
)

var RoleCommand = &cli.Command{
	Name:         Role,
	Usage:        "config role method, starts CUI if parameter not enough",
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

func startRoleCUI(configData *ConfigDataWithCode) {
	fmt.Println("start role cui")
}

func configRoleAction(c *cli.Context) error {
	config := NewConfig()
	profile := c.String("profile")

	configData := &ConfigDataWithCode{
		ConfigData: ConfigData{
			SerialNumber:    c.String(SerialNumber),
			DurationSeconds: c.Int64(Duration),
			OriginProfile:   c.String(SourceProfile),
			AssumeRoleArn:   c.String(RoleArn),
		},
		Profile: profile,
		Code:    "",
	}

	// Check original profile, if original profile contains token (one time),
	// maximum duration is 1 hour. start gui to confirm
	originProfile, err := config.readSpecificCredential(configData.OriginProfile)
	if err != nil {
		fmt.Printf("source profile: (%s) doesn't exists", configData.OriginProfile)
		os.Exit(1)
		// startRoleCUI(configData)
		// return nil
	}
	if originProfile.SessionToken != "" {
		fmt.Printf("source profile is a temporary profile with maximum 1 hour")
	}
	if configData.DurationSeconds == 0 || c.Int64(Duration) != DefaultDurationSeconds {
		configData.DurationSeconds = c.Int64(Duration)
	}
	serial := c.String(SerialNumber)
	if serial == "" {
		startRoleCUI(configData)
	} else {
		configData.SerialNumber = serial
		config.backupProfile(profile, credentialsFile_)
		config.saveConfig(configData, configFile_)
	}
	return nil
}

func configRoleBashComplete(c *cli.Context) {

}

func getRoleSession(input *ConfigDataWithCode) (*CredentialData, error) {
	if input.OriginProfile == "" {
		return nil, fmt.Errorf("'origin_profile' is not present in %s", input.Profile)
	}
	svc := getSession(input.Profile)
	assumeRoleInput := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(input.DurationSeconds),
		RoleArn:         &input.AssumeRoleArn,
		RoleSessionName: aws.String("cli"),
	}
	if input.SerialNumber != "" {
		assumeRoleInput.SerialNumber = &input.SerialNumber
		assumeRoleInput.TokenCode = &input.Code
	}
	output, err := svc.AssumeRole(assumeRoleInput)
	if err != nil {
		return nil, err
	}
	return &CredentialData{
		AccessKey:    *output.Credentials.AccessKeyId,
		SecretKey:    *output.Credentials.SecretAccessKey,
		SessionToken: *output.Credentials.SessionToken,
	}, nil
}

func loginForRole() error {
	return nil
}
