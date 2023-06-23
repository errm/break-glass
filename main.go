package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"gopkg.in/ini.v1"
)

func main() {

	profiles := flag.String("profiles", "all", "comma separated profiles to get credentials for")
	flag.Parse()

	context := context.Background()
	cfg, err := config.LoadDefaultConfig(context)
	check(err)

	client := sts.NewFromConfig(cfg)

	homedir, err := os.UserHomeDir()
	check(err)

	credentialsPath := homedir + "/.aws/credentials"
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		file, err := os.Create(credentialsPath)
		file.Chmod(0600)
		check(err)
		file.Close()
	}

	credentials, err := ini.Load(credentialsPath)
	check(err)

	config, err := ini.Load(homedir + "/.aws/break-glass")
	check(err)

	sections := []*ini.Section{}

	if *profiles == "all" {
		sections = config.Sections()
	} else {
		for _, profile := range strings.Split(*profiles, ",") {
			section, err := config.SectionsByName(profile)
			check(err)
			sections = append(sections, section...)
		}
	}

	check(updateCredentials(client, context, sections, credentials, os.Stdin))
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

func updateCredentials(client STSAssumeRoleAPI, context context.Context, sections []*ini.Section, credentials *ini.File, console io.Reader) error {
	for _, section := range sections {
		if section.Name() == "DEFAULT" {
			continue
		}
		arn := section.Key("aws_role_arn").String()
		duration := int32(section.Key("duration").MustInt(900))
		sessionName := "break-glass-session"
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
