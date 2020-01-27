package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type ProfileWithCode struct {
	ProfileName     string
	DurationSeconds *int64
	SerialNumber    *string
	TokenCode       *string
}

// getSession get sts session with provided profile read from .aws folder
func getSession(profile string) *sts.STS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: profile,
	}))
	return sts.New(sess)
}

func getAccountId(profile string) string {
	svc := getSession(profile)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return ""
	}
	fmt.Println(result)
	return *result.Account
}
