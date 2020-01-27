package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/urfave/cli/v2"
)

func configAction(c *cli.Context) error {
	fmt.Printf("%s", c.Args().Slice())
	return nil
}

func configBashComplete(c *cli.Context) {

}

func configMFABashComplete(c *cli.Context) {

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
		return loginForRole()
	} else {
		return loginForMFA(config, profile, code)
	}
}

func loginBashComplete(c *cli.Context) {

}
