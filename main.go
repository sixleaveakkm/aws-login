package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

const (
	Duration      = "duration"
	Profile       = "profile"
	SerialNumber  = "serial-number"
	MFA           = "mfa"
	Role          = "role"
	SourceProfile = "source-profile"
	RoleArn       = "role-arn"
	NoMFA         = "no-mfa"
	// DefaultDurationSeconds 12 hours
	DefaultDurationSeconds = 43200
)

var logger *log.Logger

func init() {
	ini.PrettyEqual = true
	ini.PrettyFormat = false
	setAWSFolderDefault()
	logger = log.New(os.Stdout, "Logger: ", log.Ltime)
}

func main() {
	executor(os.Args)
}

func executor(args []string) {
	app := &cli.App{
		Name:                 "aws-login",
		Usage:                "login your aws cli",
		Version:              "0.1",
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
		Action:       login,
		BashComplete: loginBashComplete,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "config MFA or role method, starts CUI if parameter not enough",
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
