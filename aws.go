package main

import (
	"time"

	aws_ "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	MFAPrefix = "arn:aws:iam::"
)

type GetMFASessionInput struct {
	// Profile name without mfa
	Profile         string
	SerialNumber    string
	DurationSeconds int64
	Code            string
}

type GetAssumeRoleRoleInput struct {
	// SourceProfile name of original profile name
	SourceProfile   string
	AssumeRoleArn   string
	SerialNumber    string
	DurationSeconds int64
	Code            string
}

type AWS interface {
	// GetMFAString get mfa string with 1.5 seconds timeout.
	// GetMFAString is only used for completion.
	//
	// It get mfa information by calling aws api of this session,
	// which requires permission `iam:ListMFADevices` to
	// at least Resource `arn:aws:iam::*:user/${aws:username}` (your own user)
	// with timeout 1.5 second, if timeout, it returns fix prefix "arn:aws:iam::"
	GetMFAString(profile string) string

	GetMFASession(input *GetMFASessionInput) (*SessionCredential, error)
	GetAssumeRoleSession(input *GetAssumeRoleRoleInput) (*SessionCredential, error)
}

type AWSImpl struct {
}

func (s AWSImpl) GetMFAString(profile string) string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{Profile: profile}))
	str := make(chan string, 1)
	go func() {
		si := iam.New(sess)
		res, err := si.ListMFADevices(&iam.ListMFADevicesInput{})
		if err != nil || len(res.MFADevices) == 0 {
			str <- MFAPrefix
		} else if sn := res.MFADevices[0].SerialNumber; sn != nil {
			str <- *sn
		}
	}()

	select {
	case res := <-str:
		return res
	case <-time.After(1500 * time.Millisecond):
		return MFAPrefix
	}
}

func (s AWSImpl) GetMFASession(input *GetMFASessionInput) (*SessionCredential, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{Profile: input.Profile}))
	svc := sts.New(sess)

	output, err := svc.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws_.Int64(input.DurationSeconds),
		SerialNumber:    aws_.String(input.SerialNumber),
		TokenCode:       aws_.String(input.Code),
	})
	if err != nil {
		return nil, err
	}
	return &SessionCredential{
		AccessKey:    *output.Credentials.AccessKeyId,
		SecretKey:    *output.Credentials.SecretAccessKey,
		SessionToken: *output.Credentials.SessionToken,
	}, nil
}

func (s AWSImpl) GetAssumeRoleSession(input *GetAssumeRoleRoleInput) (*SessionCredential, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{Profile: input.SourceProfile}))
	svc := sts.New(sess)

	assumeRoleInput := &sts.AssumeRoleInput{
		DurationSeconds: aws_.Int64(input.DurationSeconds),
		SerialNumber:    aws_.String(input.SerialNumber),
		RoleArn:         &input.AssumeRoleArn,
		// Give a dummy session name
		RoleSessionName: aws_.String("cli"),
	}
	if input.SerialNumber != "" {
		assumeRoleInput.SerialNumber = &input.SerialNumber
		assumeRoleInput.TokenCode = &input.Code
	}
	output, err := svc.AssumeRole(assumeRoleInput)
	if err != nil {
		return nil, err
	}
	return &SessionCredential{
		AccessKey:    *output.Credentials.AccessKeyId,
		SecretKey:    *output.Credentials.SecretAccessKey,
		SessionToken: *output.Credentials.SessionToken,
	}, nil
}
