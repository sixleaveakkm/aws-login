package main

import (
	"fmt"
	"os"
	"strings"

	a "github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
)

func getProfile(c *cli.Context) string {
	profile := c.String(Profile)
	if profile == "" {
		if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" && !strings.HasSuffix(envProfile, "_no_mfa") {
			profile = envProfile
			fmt.Print("Using AWS_PROFILE environment: ", a.Bold(a.Blue(fmt.Sprintf("%s\n", envProfile))))
		} else {
			profile = "default"
		}
	}
	return profile
}
