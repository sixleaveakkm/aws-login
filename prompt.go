package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	a "github.com/logrusorgru/aurora"
)

// promptSixDigitCode, prompt user to enter six digit code and return the code with enter
// If input if incorrect, prompt to re-enter
func promptSixDigitCode() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(a.Bold(a.BrightCyan("MFA code: ")))

	for {
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if isSixDigit(text) {
			return text
		}

		fmt.Printf("%s, %s",
			a.Bold(a.BrightRed("x Invalid Input")),
			a.Bold(a.BrightCyan("MFA code: ")),
		)
	}
}
