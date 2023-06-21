package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"gopkg.in/ini.v1"
)

func main() {
	context := context.Background()
	cfg, err := config.LoadDefaultConfig(context)
	check(err)

	client := sts.NewFromConfig(cfg)

	homedir, err := os.UserHomeDir()
	check(err)

	config, err := ini.Load(homedir + "/.aws/multipass")
	check(err)

	credentials, err := ini.Load(homedir + "/.aws/credentials")
	check(err)

	check(updateCredentials(client, context, config, credentials, os.Stdin))
	check(credentials.SaveTo(homedir + "/.aws/credentials"))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// This interface is for ease of testing our code
type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func updateCredentials(client STSAssumeRoleAPI, context context.Context, config, credentials *ini.File, console io.Reader) error {
	for _, section := range config.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}
		arn := section.Key("aws_role_arn").String()
		duration := int32(section.Key("duration").MustInt(900))
		sessionName := "multipass-session"
		input := &sts.AssumeRoleInput{
			RoleArn:         &arn,
			RoleSessionName: &sessionName,
			DurationSeconds: &duration,
		}

		if section.HasKey("aws_mfa_device") {
			device := section.Key("aws_mfa_device").String()
			input.SerialNumber = &device
			token := getToken(console, section.Name())
			input.TokenCode = &token
		}
		result, err := client.AssumeRole(context, input)
		if err != nil {
			return err
		}

		section, err = credentials.NewSection(section.Name())
		if err != nil {
			return err
		}
		section.NewKey("aws_access_key_id", *result.Credentials.AccessKeyId)
		section.NewKey("aws_secret_access_key", *result.Credentials.SecretAccessKey)
		section.NewKey("aws_security_token", *result.Credentials.SessionToken)
	}
	return nil
}

func getToken(reader io.Reader, name string) string {
	fmt.Printf("MFA Token for %s -> ", name)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
