package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

const Version = "0.4"

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

func init() {
	ini.PrettyEqual = true
	ini.PrettyFormat = false
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
		Action:       login,
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

func printWithExplain(v string, e string) {
	escapedV := strings.Replace(v, ":", "\\:", -1)
	escapedV = strings.Replace(escapedV, " ", "\\ ", -1)
	if e == "" {
		fmt.Println(escapedV)
	} else {
		fmt.Printf("%s:%s\n", escapedV, e)
	}
}

// isSixDigit checks given string is a syntax valid mfa code
func isSixDigit(code string) bool {
	reg := regexp.MustCompile(`^\d{6}$`)
	return reg.MatchString(code)
}

// getLastArgument get last not --generate-bash-completion argument for n's level sub-command
func getLastArgument(level int) string {
	args := os.Args[1+level:]
	l := len(args)
	if l == 0 {
		return ""
	}
	if l == 1 {
		if args[0] == "--generate-bash-completion" {
			return ""
		} else {
			return args[0]
		}
	}
	if args[l-1] == "--generate-bash-completion" {
		return args[l-2]
	}
	return args[l-1]
}
