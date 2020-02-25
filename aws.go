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

// getMFAString get maximum mfa string or its prefix
func getMFAString(profile string) string {
	accountId := getAccountId(profile)
	userId := getUserId(profile)
	res := "arn:aws:iam::"
	if accountId == "" {
		return res
	}
	res += accountId + ":mfa/"
	if userId == "" {
		return res
	}
	return res + userId
}

func getUserId(profile string) string {
	svc := getSession(profile)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return ""
	}
	fmt.Println(result)
	return *result.UserId
}
