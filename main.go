package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

const Version = "0.5"

const (
	Duration           = "duration"
	Profile            = "profile"
	SerialNumber       = "serial-number"
	SerialNumberInFile = "mfa_serial"
	MFA                = "mfa"
	Role               = "role"
	SourceProfile      = "source-profile"
	RoleArn            = "role-arn"
	NoMFA              = "no-mfa"
	// DefaultDurationSeconds 12 hours
	DefaultDurationSeconds = 43200
)

var (
	aws AWS
)

func init() {
	ini.PrettyEqual = true
	ini.PrettyFormat = false

	aws = AWSImpl{}
	setAWSFolderDefault()
}

func main() {
	executor(os.Args)
}

func executor(args []string) {
	app := &cli.App{
		Name:                 "aws-login",
		Usage:                "login your aws cli",
		Version:              Version,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				Aliases: []string{"p"},
				Usage:   "mfa or role profile name to login, default is \"default\"",
				Value:   "default",
			},
			&cli.BoolFlag{
				Name:    "default",
				Aliases: []string{"d"},
				Usage:   "profile set as default",
				Value:   false,
			},
		},
		Action:       loginAction,
		BashComplete: loginBashComplete,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "config MFA or role method",
				Subcommands: []*cli.Command{
					MFACommand,
					RoleCommand,
				},
				Action:       configAction,
				BashComplete: configBashComplete,
			},
		},
	}
	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

// configAction for `aws-login config` which only contains two sub-commands
func configAction(c *cli.Context) error {
	return cli.ShowAppHelp(c)
}

// isSixDigit checks given string is a syntax valid mfa code
func isSixDigit(code string) bool {
	reg := regexp.MustCompile(`^\d{6}$`)
	return reg.MatchString(code)
}

// login process the input, and handler to mfa's or role's login function
// the input profile is checked previously
func loginAction(c *cli.Context) error {
	profile := c.String(Profile)
	code := c.Args().Get(0)
	if !isSixDigit(code) {
		return fmt.Errorf("input code must be 6 digit, got '%s'", code)
	}

	config := NewConfig(awsFoldPath)
	confSection, err := config.Conf.GetSection(profile)
	if err != nil {
		confSection, err = config.Conf.GetSection(fmt.Sprintf("profile %s", profile))
		if err != nil {
			scriptName := os.Args[0]
			return fmt.Errorf("%q %w\nYou could try:\n\t%s config <mfa|role> ...\n to create config", profile, NoProfileError, scriptName)
		}
	}
	var confData ConfigData
	_ = confSection.MapTo(&confData)

	setToDefault := c.Bool("default")
	if confData.SourceProfile != "" {
		return loginForRole(config, profile, code, setToDefault)
	} else {
		return loginForMFA(config, profile, code, setToDefault)
	}
}
