package main

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli/v2"
)

const (
	TextGenerateConfig = "generate new config item"
	TextProfile        = "login use profile"
	TextShowHelp       = "display help document"
	TextShowVersion    = "show version"
	TextSetToDefault   = "set current profile to default"

	TextConfigMFA  = "generate config using mfa"
	TextConfigRole = "generate config for role using mfa"

	TextConfMFAProfile = "the profile name use to login with mfa, notice it will move profile to <name>_no_mfa and generate a new profile using mfa"

	TextSerialNumber = "mfa serial number for your account"
	TextDuration     = "session duration in seconds, default is 12 hours, notice if you use mfa from a session, the duration will be 1 hour max"
	TextRoleArn      = "arn of the role to assume"

	TextConfRoleSourceProfile = "origin profile name to perform assume role"
	TextConfRoleProfile       = "new profile name using role"
)

// loginBashComplete, bash complete for `aws-login`
func loginBashComplete(c *cli.Context) {
	last := getLastArgument(0)
	if last == "" {
		printWithExplain("config", TextGenerateConfig)
		printWithExplain("--profile", TextProfile)
		printWithExplain("--help", TextShowHelp)
		printWithExplain("--version", TextShowVersion)
		return
	}

	if last == "-p" || last == "--profile" {
		for p := range NewConfig().listAWSProfiles().Iter() {
			fmt.Println(strings.ReplaceAll(p.(string), " ", "\\ "))
		}
		return
	}

	flagSet := mapset.NewSet()
	for _, f := range c.FlagNames() {
		flagSet.Add(f)
	}
	if !flagSet.Contains("profile") {
		if last == "-" {
			printWithExplain("p", TextProfile)
		} else if last == "--" {
			printWithExplain("profile", TextProfile)
		} else {
			printWithExplain("-p", TextProfile)
		}
	}
	if !flagSet.Contains("default") {
		if last == "-" {
			printWithExplain("d", TextSetToDefault)
		} else if last == "--" {
			printWithExplain("default", TextSetToDefault)
		} else {
			printWithExplain("-d", TextSetToDefault)
		}
	}
}

// configBashComplete, bash complete for `aws-login config`
func configBashComplete(_ *cli.Context) {
	printWithExplain(MFA, TextConfigMFA)
	printWithExplain(Role, TextConfigRole)
}

// configMFABashComplete, bash complete for `aws-login config mfa`
func configMFABashComplete(c *cli.Context) {
	last := getLastArgument(2)
	if last == "-p" || last == "--profile" {
		for p := range NewConfig().listAWSProfiles().Iter() {
			printWithExplain(p.(string), "")
		}
		return
	}

	flagSet := mapset.NewSet()
	for _, f := range c.FlagNames() {
		flagSet.Add(f)
	}

	// if current position is serial number and profile is given
	// try to return mfa string by fetch information for aws
	//

	if last == "-n" || last == fmt.Sprintf("--%s", SerialNumber) {
		if flagSet.Contains("profile") {
			p := c.String("profile")
			// todo: possible session timeout if config same profile name twice
			printWithExplain(getMFAString(p), "mfa string or prefix for given profile")
		}
		return
	}

	if !flagSet.Contains("profile") {
		if last == "-" {
			printWithExplain("p", TextConfMFAProfile)
		} else if last == "--" {
			printWithExplain("profile", TextConfMFAProfile)
		} else {
			printWithExplain("-p", TextConfMFAProfile)
		}
	}
	if !flagSet.Contains(SerialNumber) {
		if last == "-" {
			printWithExplain("n", TextSerialNumber)
		} else if last == "--" {
			printWithExplain(SerialNumber, TextSerialNumber)
		} else {
			printWithExplain("-n", TextSerialNumber)
		}
	}
	if !flagSet.Contains(Duration) {
		if last == "-" {
			printWithExplain("t", TextDuration)
		} else if last == "--" {
			printWithExplain(Duration, TextDuration)
		} else {
			printWithExplain("-t", TextDuration)
		}
	}
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
			// todo: possible session timeout if config same profile name twice
			printWithExplain(getMFAString(p), "mfa string or prefix for given profile")
		}
		return
	}

	if last == "-r" || last == fmt.Sprintf("--%s", RoleArn) {
		return
	}

	if !flagSet.Contains(SourceProfile) {
		if last == "-" {
			printWithExplain("s", TextConfRoleSourceProfile)
		} else if last == "--" {
			printWithExplain(SourceProfile, TextConfRoleSourceProfile)
		} else {
			printWithExplain("-s", TextConfRoleSourceProfile)
		}
	}

	if !flagSet.Contains(Profile) {
		if last == "-" {
			printWithExplain("p", TextConfRoleProfile)
		} else if last == "--" {
			printWithExplain("profile", TextConfRoleProfile)
		} else {
			printWithExplain("-p", TextConfRoleProfile)
		}
	}

	if !flagSet.Contains(RoleArn) {
		if last == "-" {
			printWithExplain("r", TextRoleArn)
		} else if last == "--" {
			printWithExplain(RoleArn, TextRoleArn)
		} else {
			printWithExplain("-r", TextRoleArn)
		}
	}

	if !flagSet.Contains(SerialNumber) {
		if last == "-" {
			printWithExplain("n", TextSerialNumber)
		} else if last == "--" {
			printWithExplain(SerialNumber, TextSerialNumber)
		} else {
			printWithExplain("-n", TextSerialNumber)
		}
	}
	if !flagSet.Contains(Duration) {
		if last == "-" {
			printWithExplain("t", TextDuration)
		} else if last == "--" {
			printWithExplain(Duration, TextDuration)
		} else {
			printWithExplain("-t", TextDuration)
		}
	}

}
