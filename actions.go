package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/urfave/cli/v2"
)

func configAction(c *cli.Context) error {
	fmt.Printf("%s", c.Args().Slice())
	return nil
}

func configBashComplete(c *cli.Context) {
	for _, c := range c.App.Commands {
		fmt.Println(c.Name)
	}
}

func isSixDigit(code string) bool {
	reg := regexp.MustCompile(`^\d{6}$`)
	return reg.MatchString(code)
}

// login process the input, and handler to mfa's or role's login function
// the input profile must exists in config file unless it is not set
func login(c *cli.Context) error {
	profile := c.String(Profile)
	code := c.Args().Get(0)
	if !isSixDigit(code) {
		return fmt.Errorf("input code must be 6 digit, got '%s'", code)
	}

	config := NewConfig()
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

	if confData.OriginProfile != "" {
		return loginForRole(config, profile, code)
	} else {
		return loginForMFA(config, profile, code)
	}
}

func getLastArgument(n int) string {
	args := os.Args[1+n:]
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

func loginBashComplete(c *cli.Context) {
	last := getLastArgument(0)
	if last == "" {
		fmt.Println("config")
		fmt.Println("-p")
		fmt.Println("--help")
		fmt.Println("--version")
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
			fmt.Println("p")
		} else if last == "--" {
			fmt.Println("profile")
		} else {
			fmt.Println("-p")
		}
	}
	if !flagSet.Contains("default") {
		if last == "-" {
			fmt.Println("d")
		} else if last == "--" {
			fmt.Println("default")
		} else {
			fmt.Println("-d")
		}
	}
}
