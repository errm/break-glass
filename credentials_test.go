package main

import (
	"context"
	"strings"
	"testing"
	"reflect"
	   "encoding/json"


	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"

	"gopkg.in/ini.v1"
)

func TestUpdateCredentials(t *testing.T) {
	context := context.Background()
	client := &MockSTSClient{
		t: t,
		expectations: []STSExpectation{
			STSExpectation{
				params: &sts.AssumeRoleInput{
					RoleArn:         p("arn:aws:iam::012345678901:role/Admin"),
					RoleSessionName: p("break-glass-session"),
					DurationSeconds: p(int32(3600)),
					SerialNumber:    p("arn:aws:iam::012345678901:mfa/iphone"),
					TokenCode:       p("123456"),
				},
				output: &sts.AssumeRoleOutput{
					Credentials: &types.Credentials{
						AccessKeyId:     p("ACCESS_KEY_ID"),
						SecretAccessKey: p("SECRET_KEY"),
						SessionToken:    p("SESSION_TOKEN"),
					},
				},
			},
		},
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

	err = updateCredentials(client, context, config.Sections(), credentials, console)
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
	t            *testing.T
	expectations []STSExpectation
}

type STSExpectation struct {
	params *sts.AssumeRoleInput
	output *sts.AssumeRoleOutput
}

func (c *MockSTSClient) AssumeRole(ctx context.Context,
	params *sts.AssumeRoleInput,
	optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {

	var expected STSExpectation
	expected, c.expectations = c.expectations[0], c.expectations[1:]

	if !reflect.DeepEqual(params, expected.params) {
		c.t.Errorf("Expected params to be %s but got %s", j(*expected.params, c.t), j(*params, c.t))
	}

	return expected.output, nil
}

func p[Value any](v Value) *Value {
    return &v
}

func j(v interface{}, t *testing.T) []byte {
    valueJson, err := json.Marshal(v)
    if err != nil {
      t.Fatal(err)
    }
    return valueJson
}
