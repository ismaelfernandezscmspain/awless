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

// DO NOT EDIT
// This file was automatically generated with go generate
package awsspec

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/wallix/awless/logger"
)

type CreateInstance struct {
	_             string `action:"create" entity:"instance" awsAPI:"ec2" awsCall:"RunInstances" awsInput:"ec2.RunInstancesInput" awsOutput:"ec2.Reservation" awsDryRun:""`
	logger        *logger.Logger
	api           ec2iface.EC2API
	Image         *string   `awsName:"ImageId" awsType:"awsstr" templateName:"image" required:""`
	Count         *int64    `awsName:"MaxCount" awsType:"awsint64" templateName:"count" required:""`
	Count         *int64    `awsName:"MinCount" awsType:"awsint64" templateName:"count" required:""`
	Type          *string   `awsName:"InstanceType" awsType:"awsstr" templateName:"type" required:""`
	Subnet        *string   `awsName:"SubnetId" awsType:"awsstr" templateName:"subnet" required:""`
	Name          *struct{} `awsName:"Name" templateName:"name" required:""`
	Keypair       *string   `awsName:"KeyName" awsType:"awsstr" templateName:"keypair"`
	Ip            *string   `awsName:"PrivateIpAddress" awsType:"awsstr" templateName:"ip"`
	Userdata      *string   `awsName:"UserData" awsType:"awsfiletobase64" templateName:"userdata"`
	Securitygroup *[]string `awsName:"SecurityGroupIds" awsType:"awsstringslice" templateName:"securitygroup"`
	Lock          *bool     `awsName:"DisableApiTermination" awsType:"awsbool" templateName:"lock"`
	Role          *string   `awsName:"IamInstanceProfile.Name" awsType:"awsstr" templateName:"role"`
}

func (cmd *CreateInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *CreateInstance) ExtractResult(i interface{}) string {
	return awssdk.StringValue(i.(*ec2.Reservation).Instances[0].InstanceId)
}

type UpdateInstance struct {
	_      string `action:"update" entity:"instance" awsAPI:"ec2" awsCall:"ModifyInstanceAttribute" awsInput:"ec2.ModifyInstanceAttributeInput" awsOutput:"ec2.ModifyInstanceAttributeOutput" awsDryRun:""`
	logger *logger.Logger
	api    ec2iface.EC2API
	Id     *string `awsName:"InstanceId" awsType:"awsstr" templateName:"id" required:""`
	Type   *string `awsName:"InstanceType.Value" awsType:"awsstr" templateName:"type"`
	Lock   *bool   `awsName:"DisableApiTermination" awsType:"awsboolattribute" templateName:"lock"`
}

func (cmd *UpdateInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

type DeleteInstance struct {
	_      string `action:"delete" entity:"instance" awsAPI:"ec2" awsCall:"TerminateInstances" awsInput:"ec2.TerminateInstancesInput" awsOutput:"ec2.TerminateInstancesOutput" awsDryRun:""`
	logger *logger.Logger
	api    ec2iface.EC2API
	Id     *[]string `awsName:"InstanceIds" awsType:"awsstringslice" templateName:"id" required:""`
}

func (cmd *DeleteInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

type StartInstance struct {
	_      string `action:"start" entity:"instance" awsAPI:"ec2" awsCall:"StartInstances" awsInput:"ec2.StartInstancesInput" awsOutput:"ec2.StartInstancesOutput" awsDryRun:""`
	logger *logger.Logger
	api    ec2iface.EC2API
	Id     *[]string `awsName:"InstanceIds" awsType:"awsstringslice" templateName:"id" required:""`
}

func (cmd *StartInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *StartInstance) ExtractResult(i interface{}) string {
	return awssdk.StringValue(i.(*ec2.StartInstancesOutput).StartingInstances[0].InstanceId)
}

type StopInstance struct {
	_      string `action:"stop" entity:"instance" awsAPI:"ec2" awsCall:"StopInstances" awsInput:"ec2.StopInstancesInput" awsOutput:"ec2.StopInstancesOutput" awsDryRun:""`
	logger *logger.Logger
	api    ec2iface.EC2API
	Id     *[]string `awsName:"InstanceIds" awsType:"awsstringslice" templateName:"id" required:""`
}

func (cmd *StopInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *StopInstance) ExtractResult(i interface{}) string {
	return awssdk.StringValue(i.(*ec2.StopInstancesOutput).StoppingInstances[0].InstanceId)
}

type CheckInstance struct {
	_       string `action:"check" entity:"instance" awsAPI:"ec2"`
	logger  *logger.Logger
	api     ec2iface.EC2API
	Id      *struct{} `templateName:"id" required:""`
	State   *struct{} `templateName:"state" required:""`
	Timeout *struct{} `templateName:"timeout" required:""`
}

type AttachInstance struct {
	_           string `action:"attach" entity:"instance" awsAPI:"elbv2" awsCall:"RegisterTargets" awsInput:"elbv2.RegisterTargetsInput" awsOutput:"elbv2.RegisterTargetsOutput"`
	logger      *logger.Logger
	api         elbv2iface.ELBV2API
	Targetgroup *string   `awsName:"TargetGroupArn" awsType:"awsstr" templateName:"targetgroup" required:""`
	Id          *struct{} `awsName:"Targets[0]Id" awsType:"awsslicestruct" templateName:"id" required:""`
	Port        *int64    `awsName:"Targets[0]Port" awsType:"awsslicestructint64" templateName:"port"`
}

func (cmd *AttachInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

type DetachInstance struct {
	_           string `action:"detach" entity:"instance" awsAPI:"elbv2" awsCall:"DeregisterTargets" awsInput:"elbv2.DeregisterTargetsInput" awsOutput:"elbv2.DeregisterTargetsOutput"`
	logger      *logger.Logger
	api         elbv2iface.ELBV2API
	Targetgroup *string   `awsName:"TargetGroupArn" awsType:"awsstr" templateName:"targetgroup" required:""`
	Id          *struct{} `awsName:"Targets[0]Id" awsType:"awsslicestruct" templateName:"id" required:""`
}

func (cmd *DetachInstance) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}