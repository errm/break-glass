package main

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"gopkg.in/ini.v1"
)

func TestUpdateCredentials(t *testing.T) {
	context := context.Background()
	client := MockSTSClient{
		t:                t,
		expectedARN:      "arn:aws:iam::012345678901:role/Admin",
		expectedMFA:      "arn:aws:iam::012345678901:mfa/iphone",
		expectedToken:    "123456",
		expectedDuration: 3600,
	}

	configSource := strings.NewReader(`[test]
aws_role_arn =  arn:aws:iam::012345678901:role/Admin
aws_mfa_device = arn:aws:iam::012345678901:mfa/iphone
duration = 3600
`)
	config, err := ini.Load(configSource)
	if err != nil {
		t.Fatal(err)
	}

	credentials := ini.Empty()
	console := strings.NewReader("123456")

	err = updateCredentials(client, context, config, credentials, console)
	if err != nil {
		t.Fatal(err)
	}

	if !credentials.HasSection("test") {
		t.Error("Expected a test section in credentials file")
	}

	section, err := credentials.GetSection("test")
	if err != nil {
		t.Fatal(err)
	}

	accessKeyId := section.Key("aws_access_key_id").String()
	if accessKeyId != "ACCESS_KEY_ID" {
		t.Errorf("expected aws_access_key_id to equal ACCESS_KEY_ID but got %s", accessKeyId)
	}

	secretKey := section.Key("aws_secret_access_key").String()
	if secretKey != "SECRET_KEY" {
		t.Errorf("expected aws_secret_access_key to equal SECRET_KEY but got %s", secretKey)
	}

	securityToken := section.Key("aws_security_token").String()
	if securityToken != "SESSION_TOKEN" {
		t.Errorf("expected aws_security_token to equal SESSION_TOKEN but got %s", securityToken)
	}
}

type MockSTSClient struct {
	t                *testing.T
	expectedARN      string
	expectedToken    string
	expectedMFA      string
	expectedDuration int32
}

func (c MockSTSClient) AssumeRole(ctx context.Context,
	params *sts.AssumeRoleInput,
	optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {

	if *params.RoleArn != c.expectedARN {
		c.t.Errorf("Expected role ARN to be %s got %s", c.expectedARN, *params.RoleArn)
	}

	if *params.SerialNumber != c.expectedMFA {
		c.t.Errorf("Expected MFA SerialNumber to be %s got %s", c.expectedMFA, *params.SerialNumber)
	}

	if *params.TokenCode != c.expectedToken {
		c.t.Errorf("Expected MFA Token to be %s got %s", c.expectedToken, *params.TokenCode)
	}

	if *params.DurationSeconds != c.expectedDuration {
		c.t.Errorf("Expected duration to be %d got %d", c.expectedDuration, *params.DurationSeconds)
	}

	credentials := types.Credentials{
		AccessKeyId:     aws.String("ACCESS_KEY_ID"),
		SecretAccessKey: aws.String("SECRET_KEY"),
		SessionToken:    aws.String("SESSION_TOKEN"),
	}

	output := &sts.AssumeRoleOutput{
		Credentials: &credentials,
	}

	return output, nil
}
