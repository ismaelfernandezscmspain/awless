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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/wallix/awless/aws/driver"
	"github.com/wallix/awless/logger"
)

type CreatePolicy struct {
	_           string `action:"create" entity:"policy" awsAPI:"iam" awsCall:"CreatePolicy" awsInput:"iam.CreatePolicyInput" awsOutput:"iam.CreatePolicyOutput"`
	logger      *logger.Logger
	api         iamiface.IAMAPI
	Name        *string   `awsName:"PolicyName" awsType:"awsstr" templateName:"name" required:""`
	Effect      *string   `templateName:"effect" required:""`
	Action      []*string `templateName:"action" required:""`
	Resource    []*string `templateName:"resource" required:""`
	Description *string   `awsName:"Description" awsType:"awsstr" templateName:"description"`
	Document    *string   `awsName:"PolicyDocument" awsType:"awsstr"`
	Conditions  []*string `templateName:"conditions"`
}

func (cmd *CreatePolicy) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *CreatePolicy) BeforeRun(ctx map[string]interface{}) error {
	stat, err := buildStatementFromParams(cmd.Effect, cmd.Resource, cmd.Action, cmd.Conditions)
	if err != nil {
		return err
	}
	policy := &policyBody{
		Version:   "2012-10-17",
		Statement: []*policyStatement{stat},
	}

	b, err := json.MarshalIndent(policy, "", " ")
	if err != nil {
		return fmt.Errorf("cannot marshal policy document: %s", err)
	}
	cmd.Document = String(string(b))
	cmd.logger.ExtraVerbosef("policy document json:\n%s\n", string(b))
	return nil
}

func (cmd *CreatePolicy) ExtractResult(i interface{}) string {
	return StringValue(i.(*iam.CreatePolicyOutput).Policy.Arn)
}

type UpdatePolicy struct {
	_              string `action:"update" entity:"policy" awsAPI:"iam" awsCall:"CreatePolicyVersion" awsInput:"iam.CreatePolicyVersionInput" awsOutput:"iam.CreatePolicyVersionOutput"`
	logger         *logger.Logger
	api            iamiface.IAMAPI
	Arn            *string   `awsName:"PolicyArn" awsType:"awsstr" templateName:"arn" required:""`
	Effect         *string   `templateName:"effect" required:""`
	Action         []*string `templateName:"action" required:""`
	Resource       []*string `templateName:"resource" required:""`
	Conditions     []*string `templateName:"conditions"`
	Document       *string   `awsName:"PolicyDocument" awsType:"awsstr"`
	DefaultVersion *bool     `awsName:"SetAsDefault" awsType:"awsbool"`
}

func (cmd *UpdatePolicy) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *UpdatePolicy) BeforeRun(ctx map[string]interface{}) error {
	document, err := cmd.getPolicyLastVersionDocument(cmd.Arn)
	if err != nil {
		return err
	}
	var defaultPolicyDocument *struct {
		Version    string             `json:",omitempty"`
		ID         string             `json:"Id,omitempty"`
		Statements []*json.RawMessage `json:"Statement,omitempty"`
	}

	if err = json.Unmarshal([]byte(document), &defaultPolicyDocument); err != nil {
		return err
	}
	stat, err := buildStatementFromParams(cmd.Effect, cmd.Resource, cmd.Action, cmd.Conditions)
	if err != nil {
		return err
	}

	var newStatement json.RawMessage
	if newStatement, err = json.Marshal(stat); err != nil {
		return err
	}
	defaultPolicyDocument.Statements = append(defaultPolicyDocument.Statements, &newStatement)

	b, err := json.MarshalIndent(defaultPolicyDocument, "", " ")
	if err != nil {
		return fmt.Errorf("cannot marshal policy document: %s", err)
	}
	cmd.Document = String(string(b))
	cmd.DefaultVersion = aws.Bool(true)
	cmd.logger.ExtraVerbosef("policy document json:\n%s\n", string(b))
	return nil
}

func (cmd *UpdatePolicy) getPolicyLastVersionDocument(arn *string) (string, error) {
	listVersionsInput := &iam.ListPolicyVersionsInput{
		PolicyArn: arn,
	}
	listVersionsOut, err := cmd.api.ListPolicyVersions(listVersionsInput)
	if err != nil {
		return "", err
	}
	var defaultVersion *iam.PolicyVersion
	for _, version := range listVersionsOut.Versions {
		if aws.BoolValue(version.IsDefaultVersion) {
			policyDetailInput := &iam.GetPolicyVersionInput{
				VersionId: version.VersionId,
				PolicyArn: arn,
			}
			var policyDetailOutput *iam.GetPolicyVersionOutput
			if policyDetailOutput, err = cmd.api.GetPolicyVersion(policyDetailInput); err != nil {
				return "", err
			}
			defaultVersion = policyDetailOutput.PolicyVersion
		}
	}
	if defaultVersion == nil {
		return "", fmt.Errorf("update policy: can not find default version for policy with arn '%s'", StringValue(arn))
	}
	document, err := url.QueryUnescape(aws.StringValue(defaultVersion.Document))
	if err != nil {
		return "", fmt.Errorf("decoding policy document: %s", err)
	}
	return document, nil
}

type DeletePolicy struct {
	_           string `action:"delete" entity:"policy" awsAPI:"iam"  awsCall:"DeletePolicy" awsInput:"iam.DeletePolicyInput" awsOutput:"iam.DeletePolicyOutput"`
	logger      *logger.Logger
	api         iamiface.IAMAPI
	Arn         *string `awsName:"PolicyArn" awsType:"awsstr" templateName:"arn" required:""`
	AllVersions *bool   `templateName:"all-versions"`
}

func (cmd *DeletePolicy) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *DeletePolicy) BeforeRun(ctx map[string]interface{}) error {
	if BoolValue(cmd.AllVersions) {
		list, err := cmd.api.ListPolicyVersions(&iam.ListPolicyVersionsInput{PolicyArn: cmd.Arn})
		if err != nil {
			return fmt.Errorf("list all policy versions: %s", err)
		}
		for _, v := range list.Versions {
			if !aws.BoolValue(v.IsDefaultVersion) {
				cmd.logger.Verbosef("deleting version '%s' of policy '%s'", aws.StringValue(v.VersionId), StringValue(cmd.Arn))
				if _, err := cmd.api.DeletePolicyVersion(&iam.DeletePolicyVersionInput{PolicyArn: cmd.Arn, VersionId: v.VersionId}); err != nil {
					return fmt.Errorf("delete version %s: %s", aws.StringValue(v.VersionId), err)
				}
			}
		}
	}
	return nil
}

type AttachPolicy struct {
	_      string `action:"attach" entity:"policy" awsAPI:"iam"`
	logger *logger.Logger
	api    iamiface.IAMAPI
	Arn    *string `awsName:"PolicyArn" awsType:"awsstr" templateName:"arn"`
	User   *string `awsName:"UserName" awsType:"awsstr" templateName:"user"`
	Group  *string `awsName:"GroupName" awsType:"awsstr" templateName:"group"`
	Role   *string `awsName:"RoleName" awsType:"awsstr" templateName:"role"`
}

func (cmd *AttachPolicy) ValidateParams(params []string) ([]string, error) {
	return paramRule{
		tree: allOf(oneOfE(node("user"), node("role"), node("group")), oneOf(node("arn"), allOf(node("access"), node("service")))),
	}.verify(params)
}

func (cmd *AttachPolicy) ConvertParams() ([]string, func(values map[string]interface{}) (map[string]interface{}, error)) {
	return []string{"access", "service"},
		func(values map[string]interface{}) (map[string]interface{}, error) {
			service, hasService := values["service"].(string)
			access, hasAccess := values["access"].(string)

			if hasService && hasAccess {
				pol, err := awsdriver.LookupAWSPolicy(service, access)
				if err != nil {
					return values, err
				}
				return map[string]interface{}{"arn": pol.Arn}, nil
			} else {
				return nil, nil
			}
		}
}

func (cmd *AttachPolicy) ManualValidateCommand(params map[string]interface{}, refs []string) (errs []error) {
	if cmd.User == nil && cmd.Group == nil && cmd.Role == nil {
		errs = append(errs, fmt.Errorf("missing required field 'user', 'group' or 'role'"))
	}
	if cmd.Arn == nil {
		errs = append(errs, fmt.Errorf("missing required field 'arn'"))
	}
	return
}

func (cmd *AttachPolicy) ManualRun(ctx map[string]interface{}) (interface{}, error) {
	start := time.Now()
	switch {
	case cmd.User != nil:
		input := &iam.AttachUserPolicyInput{}
		input.PolicyArn = cmd.Arn
		input.UserName = cmd.User
		_, err := cmd.api.AttachUserPolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.AttachUserPolicy call took %s", time.Since(start))
	case cmd.Group != nil:
		input := &iam.AttachGroupPolicyInput{}
		input.PolicyArn = cmd.Arn
		input.GroupName = cmd.Group
		_, err := cmd.api.AttachGroupPolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.AttachGroupPolicy call took %s", time.Since(start))
	case cmd.Role != nil:
		input := &iam.AttachRolePolicyInput{}
		input.PolicyArn = cmd.Arn
		input.RoleName = cmd.Role
		_, err := cmd.api.AttachRolePolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.AttachRolePolicy call took %s", time.Since(start))
	default:
		return nil, errors.New("missing one of 'user, group, role' param")
	}
	return "", nil
}

type DetachPolicy struct {
	_      string `action:"detach" entity:"policy" awsAPI:"iam"`
	logger *logger.Logger
	api    iamiface.IAMAPI
	Arn    *string `awsName:"PolicyArn" awsType:"awsstr" templateName:"arn"`
	User   *string `awsName:"UserName" awsType:"awsstr" templateName:"user"`
	Group  *string `awsName:"GroupName" awsType:"awsstr" templateName:"group"`
	Role   *string `awsName:"RoleName" awsType:"awsstr" templateName:"role"`
}

func (cmd *DetachPolicy) ValidateParams(params []string) ([]string, error) {
	return paramRule{
		tree: allOf(oneOfE(node("user"), node("role"), node("group")), oneOf(node("arn"), allOf(node("access"), node("service")))),
	}.verify(params)
}

func (cmd *DetachPolicy) ConvertParams() ([]string, func(values map[string]interface{}) (map[string]interface{}, error)) {
	return []string{"access", "service"},
		func(values map[string]interface{}) (map[string]interface{}, error) {
			service, hasService := values["service"].(string)
			access, hasAccess := values["access"].(string)

			if hasService && hasAccess {
				pol, err := awsdriver.LookupAWSPolicy(service, access)
				if err != nil {
					return values, err
				}
				return map[string]interface{}{"arn": pol.Arn}, nil
			} else {
				return nil, nil
			}
		}
}

func (cmd *DetachPolicy) ManualValidateCommand(params map[string]interface{}, refs []string) (errs []error) {
	if cmd.User == nil && cmd.Group == nil && cmd.Role == nil {
		errs = append(errs, fmt.Errorf("missing required field 'user', 'group' or 'role'"))
	}
	if cmd.Arn == nil {
		errs = append(errs, fmt.Errorf("missing required field 'arn'"))
	}
	return
}

func (cmd *DetachPolicy) ManualRun(ctx map[string]interface{}) (interface{}, error) {
	start := time.Now()
	switch {
	case cmd.User != nil:
		input := &iam.DetachUserPolicyInput{}
		input.PolicyArn = cmd.Arn
		input.UserName = cmd.User
		_, err := cmd.api.DetachUserPolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.DetachUserPolicy call took %s", time.Since(start))
	case cmd.Group != nil:
		input := &iam.DetachGroupPolicyInput{}
		input.PolicyArn = cmd.Arn
		input.GroupName = cmd.Group
		_, err := cmd.api.DetachGroupPolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.DetachGroupPolicy call took %s", time.Since(start))
	case cmd.Role != nil:
		input := &iam.DetachRolePolicyInput{}
		input.PolicyArn = cmd.Arn
		input.RoleName = cmd.Role
		_, err := cmd.api.DetachRolePolicy(input)
		if err != nil {
			return nil, err
		}
		cmd.logger.ExtraVerbosef("ec2.DetachRolePolicy call took %s", time.Since(start))
	default:
		return nil, errors.New("missing one of 'user, group, role' param")
	}
	return "", nil
}

type policyBody struct {
	Version   string
	Statement []*policyStatement
}

type policyStatement struct {
	Effect     string           `json:",omitempty"`
	Actions    []string         `json:"Action,omitempty"`
	Resources  []string         `json:"Resource,omitempty"`
	Principal  *principal       `json:",omitempty"`
	Conditions policyConditions `json:"Condition,omitempty"`
}

type principal struct {
	AWS     interface{} `json:",omitempty"`
	Service interface{} `json:",omitempty"`
}

type policyCondition struct {
	Type  string
	Key   string
	Value string
}

func buildStatementFromParams(effect *string, resource, action, condition []*string) (*policyStatement, error) {
	stat := &policyStatement{Effect: strings.Title(StringValue(effect))}
	if resource != nil {
		res := castStringSlice(resource)
		if len(res) == 1 && res[0] == "all" {
			res[0] = "*"
		}
		stat.Resources = res
	}

	if action != nil {
		stat.Actions = castStringSlice(action)
	}
	if condition != nil {
		condStr := castStringSlice(condition)
		for _, str := range condStr {
			cond, err := parseCondition(str)
			if err != nil {
				return stat, err
			}
			stat.Conditions = append(stat.Conditions, cond)
		}
	}
	return stat, nil
}

type policyConditions []*policyCondition

func (c *policyConditions) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("\"\""), nil
	}
	var buff bytes.Buffer
	buff.WriteRune('{')
	for i, cond := range *c {
		buff.WriteString(fmt.Sprintf("\"%s\":{\"%s\":\"%s\"}", cond.Type, cond.Key, cond.Value))
		if i < len(*c)-1 {
			buff.WriteRune(',')
		}
	}
	buff.WriteRune('}')
	return buff.Bytes(), nil
}

var conditionRegex = regexp.MustCompile("^([a-zA-Z0-9:_\\-\\[\\]\\*]+)(==|!=|=~|!~|<=|>=|<|>)(.*)$")

func parseCondition(condition string) (*policyCondition, error) {
	matches := conditionRegex.FindStringSubmatch(condition)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid condition '%s'", condition)
	}
	key, operator, value := matches[1], matches[2], matches[3]
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
		value = value[1 : len(value)-1]
	}
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
		value = value[1 : len(value)-1]
	}

	if strings.ToLower(value) == "null" {
		switch operator {
		case "==":
			return &policyCondition{Type: "Null", Key: key, Value: "true"}, nil
		case "!=":
			return &policyCondition{Type: "Null", Key: key, Value: "false"}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for null value '%s', expected either '==' or '!='", operator, value)
		}
	} else if strings.HasPrefix(value, "arn:") {
		switch operator {
		case "==":
			return &policyCondition{Type: "ArnEquals", Key: key, Value: value}, nil
		case "!=":
			return &policyCondition{Type: "ArnNotEquals", Key: key, Value: value}, nil
		case "=~":
			return &policyCondition{Type: "ArnLike", Key: key, Value: value}, nil
		case "!~":
			return &policyCondition{Type: "ArnNotLike", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for arn value '%s', expected either '==', '!=', '=~' or '!~'", operator, value)
		}
	} else if _, _, cidrErr := net.ParseCIDR(value); cidrErr == nil || net.ParseIP(value) != nil {
		switch operator {
		case "==":
			return &policyCondition{Type: "IpAddress", Key: key, Value: value}, nil
		case "!=":
			return &policyCondition{Type: "NotIpAddress", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for IP value '%s', expected either '==' or '!='", operator, value)
		}
	} else if _, err := time.Parse("2006-01-02T15:04:05Z", value); err == nil {
		switch operator {
		case "==":
			return &policyCondition{Type: "DateEquals", Key: key, Value: value}, nil
		case "!=":
			return &policyCondition{Type: "DateNotEquals", Key: key, Value: value}, nil
		case "<":
			return &policyCondition{Type: "DateLessThan", Key: key, Value: value}, nil
		case "<=":
			return &policyCondition{Type: "DateLessThanEquals", Key: key, Value: value}, nil
		case ">":
			return &policyCondition{Type: "DateGreaterThan", Key: key, Value: value}, nil
		case ">=":
			return &policyCondition{Type: "DateGreaterThanEquals", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for date value '%s', expected either '==', '!=', '>', '>=', '<' or '<='", operator, value)
		}
	} else if _, err := strconv.Atoi(value); err == nil {
		switch operator {
		case "==":
			return &policyCondition{Type: "NumericEquals", Key: key, Value: value}, nil
		case "!=":
			return &policyCondition{Type: "NumericNotEquals", Key: key, Value: value}, nil
		case "<":
			return &policyCondition{Type: "NumericLessThan", Key: key, Value: value}, nil
		case "<=":
			return &policyCondition{Type: "NumericLessThanEquals", Key: key, Value: value}, nil
		case ">":
			return &policyCondition{Type: "NumericGreaterThan", Key: key, Value: value}, nil
		case ">=":
			return &policyCondition{Type: "NumericGreaterThanEquals", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for int value '%s', expected either '==', '!=', '>', '>=', '<' or '<='", operator, value)
		}
	} else if b, err := strconv.ParseBool(value); err == nil {
		switch operator {
		case "==":
			return &policyCondition{Type: "Bool", Key: key, Value: fmt.Sprint(b)}, nil
		case "!=":
			return &policyCondition{Type: "Bool", Key: key, Value: fmt.Sprint(!b)}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for bool value '%s', expected either '==' or '!='", operator, value)
		}
	} else if _, err := base64.StdEncoding.DecodeString(value); value != "" && err == nil {
		switch operator {
		case "==":
			return &policyCondition{Type: "BinaryEquals", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for binary value '%s', expected '=='", operator, value)
		}
	} else {
		switch operator {
		case "==":
			return &policyCondition{Type: "StringEquals", Key: key, Value: value}, nil
		case "!=":
			return &policyCondition{Type: "StringNotEquals", Key: key, Value: value}, nil
		case "=~":
			return &policyCondition{Type: "StringLike", Key: key, Value: value}, nil
		case "!~":
			return &policyCondition{Type: "StringNotLike", Key: key, Value: value}, nil
		default:
			return nil, fmt.Errorf("invalid operator '%s' for string value '%s', expected either '==', '!=', '=~' or '!~'", operator, value)
		}
	}
}
