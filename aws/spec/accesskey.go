/* Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsspec

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/wallix/awless/aws/driver"
	"github.com/wallix/awless/logger"
)

type CreateAccesskey struct {
	_        string `action:"create" entity:"accesskey" awsAPI:"iam" awsCall:"CreateAccessKey" awsInput:"iam.CreateAccessKeyInput" awsOutput:"iam.CreateAccessKeyOutput"`
	logger   *logger.Logger
	api      iamiface.IAMAPI
	User     *string `awsName:"UserName" awsType:"awsstr" templateName:"user" required:""`
	NoPrompt *bool   `templateName:"no-prompt"`
}

func (cmd *CreateAccesskey) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *CreateAccesskey) AfterRun(ctx map[string]interface{}, output interface{}) error {
	accessKey := output.(*iam.CreateAccessKeyOutput).AccessKey
	cmd.logger.Infof("Access key created. Here are the crendentials for user %s:", aws.StringValue(accessKey.UserName))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, strings.Repeat("*", 64))
	fmt.Fprintf(os.Stderr, "aws_access_key_id = %s\n", aws.StringValue(accessKey.AccessKeyId))
	fmt.Fprintf(os.Stderr, "aws_secret_access_key = %s\n", aws.StringValue(accessKey.SecretAccessKey))
	fmt.Fprintln(os.Stderr, strings.Repeat("*", 64))
	fmt.Fprintln(os.Stderr)
	cmd.logger.Warning("This is your only opportunity to view the secret access keys.")
	cmd.logger.Warning("Save the user's new access key ID and secret access key in a safe and secure place.")
	cmd.logger.Warning("You will not have access to the secret keys again after this step.\n")

	if !BoolValue(cmd.NoPrompt) {
		var yesorno string
		fmt.Printf("Do you want to save these access keys in %s? (y/n) ", awsdriver.AWSCredFilepath)
		fmt.Scanln(&yesorno)
		if y := strings.TrimSpace(strings.ToLower(yesorno)); y == "y" || y == "yes" {
			var profile string
			fmt.Print("Entry profile name (will default to AWS 'default') ? ")
			fmt.Scanln(&profile)
			profile = strings.TrimSpace(profile)
			if profile == "" {
				profile = "default"
			}
			creds := awsdriver.NewCredsPrompter(profile)
			creds.Val.AccessKeyID = aws.StringValue(accessKey.AccessKeyId)
			creds.Val.SecretAccessKey = aws.StringValue(accessKey.SecretAccessKey)
			created, err := creds.Store()
			if err != nil {
				logger.Errorf("cannot store access keys: %s", err)
			} else {
				if created {
					fmt.Printf("\n\u2713 %s created", awsdriver.AWSCredFilepath)
				}
				fmt.Printf("\n\u2713 Credentials for profile '%s' stored successfully in %s\n\n", creds.Profile, awsdriver.AWSCredFilepath)
			}
		}
	}

	return nil
}

func (cmd *CreateAccesskey) ExtractResult(i interface{}) string {
	return StringValue(i.(*iam.CreateAccessKeyOutput).AccessKey.AccessKeyId)
}

type DeleteAccesskey struct {
	_      string `action:"delete" entity:"accesskey" awsAPI:"iam" awsCall:"DeleteAccessKey" awsInput:"iam.DeleteAccessKeyInput" awsOutput:"iam.DeleteAccessKeyOutput"`
	logger *logger.Logger
	api    iamiface.IAMAPI
	Id     *string `awsName:"AccessKeyId" awsType:"awsstr" templateName:"id" required:""`
	User   *string `awsName:"UserName" awsType:"awsstr" templateName:"user"`
}

func (cmd *DeleteAccesskey) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}
