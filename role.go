package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli/v2"
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
		// startRoleCUI(configData)
		os.Exit(1)
	} else {
		configData.SerialNumber = serial
		config.backupProfile(profile, credentialsFile_)
		config.saveConfig(configData, configFile_)
	}
	return nil
}

func configRoleBashComplete(c *cli.Context) {
	last := getLastArgument(2)
	if last == "-s" || last == "--source-profile" {
		for p := range NewConfig().listAWSProfiles().Iter() {
			fmt.Println(strings.ReplaceAll(p.(string), " ", "\\ "))
		}
		return
	}

	flagSet := mapset.NewSet()
	for _, f := range c.FlagNames() {
		flagSet.Add(f)
	}

	if last == "-n" || last == fmt.Sprintf("--%s", SerialNumber) {
		if flagSet.Contains(SourceProfile) {
			p := c.String(SourceProfile)
			mfaPrefix := getMFAPrefix(p)
			if mfaPrefix != "" {
				userId := getUserId(p)

				if userId != "" {
					fmt.Println(mfaPrefix + getUserId(p))
				}
				fmt.Println(mfaPrefix)
			} else {
				fmt.Println("arn:aws:iam::")
			}
		}
	}

	if !flagSet.Contains(SourceProfile) {
		if last == "-" {
			fmt.Println("s")
		} else if last == "--" {
			fmt.Println(SourceProfile)
		} else {
			fmt.Println("-s")
		}
	}

	if !flagSet.Contains(Profile) {
		if last == "-" {
			fmt.Println("p")
		} else if last == "--" {
			fmt.Println(Profile)
		} else {
			fmt.Println("-p")
		}
	}

	if !flagSet.Contains(RoleArn) {
		if last == "-" {
			fmt.Println("r")
		} else if last == "--" {
			fmt.Println(RoleArn)
		} else {
			fmt.Println("-r")
		}
	}

	if !flagSet.Contains(SerialNumber) {
		if last == "-" {
			fmt.Println("n")
		} else if last == "--" {
			fmt.Println(SerialNumber)
		} else {
			fmt.Println("-d")
		}
	}
	if !flagSet.Contains(Duration) {
		if last == "-" {
			fmt.Println("t")
		} else if last == "--" {
			fmt.Println(Duration)
		} else {
			fmt.Println("-t")
		}
	}

}

func getRoleSession(input *ConfigDataWithCode) (*CredentialData, error) {
	if input.OriginProfile == "" {
		return nil, fmt.Errorf("'origin_profile' is not present in %s", input.Profile)
	}
	svc := getSession(input.OriginProfile)
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

func loginForRole(config *Config, profile string, code string) error {
	// section <profile> must exists
	confSection, err := config.Conf.GetSection(profile)
	if err != nil {
		return NoProfileError
	}
	var confData ConfigData
	_ = confSection.MapTo(&confData)

	roleProfile := fmt.Sprintf("%s%s", confData.OriginProfile, excludeConfigPostfix)
	_, err = config.Cred.GetSection(roleProfile)
	if err != nil {
		roleProfile = fmt.Sprintf("profile %s%s", profile, excludeConfigPostfix)
		_, err = config.Cred.GetSection(roleProfile)
		if err != nil {
			return NoProfileError
		}
	}

	confData.OriginProfile = roleProfile
	input := &ConfigDataWithCode{
		ConfigData: confData,
		Profile:    profile,
		Code:       code,
	}

	var out *CredentialData
	out, err = getRoleSession(input)
	if err != nil {
		return fmt.Errorf("failed get mfa, %v\n", err)
	}
	config.saveCredential(out, profile)
	return nil
}
