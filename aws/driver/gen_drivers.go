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
package awsdriver

import (
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling/applicationautoscalingiface"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/wallix/awless/logger"
	"github.com/wallix/awless/template/driver"
)

type Ec2Driver struct {
	dryRun bool
	logger *logger.Logger
	ec2iface.EC2API
}

func (d *Ec2Driver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *Ec2Driver) SetLogger(l *logger.Logger) { d.logger = l }
func NewEc2Driver(api ec2iface.EC2API) driver.Driver {
	return &Ec2Driver{false, logger.DiscardLogger, api}
}

func (d *Ec2Driver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *Ec2Driver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type Elbv2Driver struct {
	dryRun bool
	logger *logger.Logger
	elbv2iface.ELBV2API
}

func (d *Elbv2Driver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *Elbv2Driver) SetLogger(l *logger.Logger) { d.logger = l }
func NewElbv2Driver(api elbv2iface.ELBV2API) driver.Driver {
	return &Elbv2Driver{false, logger.DiscardLogger, api}
}

func (d *Elbv2Driver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *Elbv2Driver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type AutoscalingDriver struct {
	dryRun bool
	logger *logger.Logger
	autoscalingiface.AutoScalingAPI
}

func (d *AutoscalingDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *AutoscalingDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewAutoscalingDriver(api autoscalingiface.AutoScalingAPI) driver.Driver {
	return &AutoscalingDriver{false, logger.DiscardLogger, api}
}

func (d *AutoscalingDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *AutoscalingDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type RdsDriver struct {
	dryRun bool
	logger *logger.Logger
	rdsiface.RDSAPI
}

func (d *RdsDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *RdsDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewRdsDriver(api rdsiface.RDSAPI) driver.Driver {
	return &RdsDriver{false, logger.DiscardLogger, api}
}

func (d *RdsDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *RdsDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type EcrDriver struct {
	dryRun bool
	logger *logger.Logger
	ecriface.ECRAPI
}

func (d *EcrDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *EcrDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewEcrDriver(api ecriface.ECRAPI) driver.Driver {
	return &EcrDriver{false, logger.DiscardLogger, api}
}

func (d *EcrDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *EcrDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type EcsDriver struct {
	dryRun bool
	logger *logger.Logger
	ecsiface.ECSAPI
}

func (d *EcsDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *EcsDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewEcsDriver(api ecsiface.ECSAPI) driver.Driver {
	return &EcsDriver{false, logger.DiscardLogger, api}
}

func (d *EcsDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *EcsDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type AcmDriver struct {
	dryRun bool
	logger *logger.Logger
	acmiface.ACMAPI
}

func (d *AcmDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *AcmDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewAcmDriver(api acmiface.ACMAPI) driver.Driver {
	return &AcmDriver{false, logger.DiscardLogger, api}
}

func (d *AcmDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *AcmDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type StsDriver struct {
	dryRun bool
	logger *logger.Logger
	stsiface.STSAPI
}

func (d *StsDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *StsDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewStsDriver(api stsiface.STSAPI) driver.Driver {
	return &StsDriver{false, logger.DiscardLogger, api}
}

func (d *StsDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *StsDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type IamDriver struct {
	dryRun bool
	logger *logger.Logger
	iamiface.IAMAPI
}

func (d *IamDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *IamDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewIamDriver(api iamiface.IAMAPI) driver.Driver {
	return &IamDriver{false, logger.DiscardLogger, api}
}

func (d *IamDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *IamDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type S3Driver struct {
	dryRun bool
	logger *logger.Logger
	s3iface.S3API
}

func (d *S3Driver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *S3Driver) SetLogger(l *logger.Logger) { d.logger = l }
func NewS3Driver(api s3iface.S3API) driver.Driver {
	return &S3Driver{false, logger.DiscardLogger, api}
}

func (d *S3Driver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *S3Driver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type SnsDriver struct {
	dryRun bool
	logger *logger.Logger
	snsiface.SNSAPI
}

func (d *SnsDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *SnsDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewSnsDriver(api snsiface.SNSAPI) driver.Driver {
	return &SnsDriver{false, logger.DiscardLogger, api}
}

func (d *SnsDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *SnsDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type SqsDriver struct {
	dryRun bool
	logger *logger.Logger
	sqsiface.SQSAPI
}

func (d *SqsDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *SqsDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewSqsDriver(api sqsiface.SQSAPI) driver.Driver {
	return &SqsDriver{false, logger.DiscardLogger, api}
}

func (d *SqsDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *SqsDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type Route53Driver struct {
	dryRun bool
	logger *logger.Logger
	route53iface.Route53API
}

func (d *Route53Driver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *Route53Driver) SetLogger(l *logger.Logger) { d.logger = l }
func NewRoute53Driver(api route53iface.Route53API) driver.Driver {
	return &Route53Driver{false, logger.DiscardLogger, api}
}

func (d *Route53Driver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *Route53Driver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type LambdaDriver struct {
	dryRun bool
	logger *logger.Logger
	lambdaiface.LambdaAPI
}

func (d *LambdaDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *LambdaDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewLambdaDriver(api lambdaiface.LambdaAPI) driver.Driver {
	return &LambdaDriver{false, logger.DiscardLogger, api}
}

func (d *LambdaDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *LambdaDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type CloudwatchDriver struct {
	dryRun bool
	logger *logger.Logger
	cloudwatchiface.CloudWatchAPI
}

func (d *CloudwatchDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *CloudwatchDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewCloudwatchDriver(api cloudwatchiface.CloudWatchAPI) driver.Driver {
	return &CloudwatchDriver{false, logger.DiscardLogger, api}
}

func (d *CloudwatchDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *CloudwatchDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type CloudfrontDriver struct {
	dryRun bool
	logger *logger.Logger
	cloudfrontiface.CloudFrontAPI
}

func (d *CloudfrontDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *CloudfrontDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewCloudfrontDriver(api cloudfrontiface.CloudFrontAPI) driver.Driver {
	return &CloudfrontDriver{false, logger.DiscardLogger, api}
}

func (d *CloudfrontDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *CloudfrontDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type CloudformationDriver struct {
	dryRun bool
	logger *logger.Logger
	cloudformationiface.CloudFormationAPI
}

func (d *CloudformationDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *CloudformationDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewCloudformationDriver(api cloudformationiface.CloudFormationAPI) driver.Driver {
	return &CloudformationDriver{false, logger.DiscardLogger, api}
}

func (d *CloudformationDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *CloudformationDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}

type ApplicationautoscalingDriver struct {
	dryRun bool
	logger *logger.Logger
	applicationautoscalingiface.ApplicationAutoScalingAPI
}

func (d *ApplicationautoscalingDriver) SetDryRun(dry bool)         { d.dryRun = dry }
func (d *ApplicationautoscalingDriver) SetLogger(l *logger.Logger) { d.logger = l }
func NewApplicationautoscalingDriver(api applicationautoscalingiface.ApplicationAutoScalingAPI) driver.Driver {
	return &ApplicationautoscalingDriver{false, logger.DiscardLogger, api}
}

func (d *ApplicationautoscalingDriver) LookupIface(lookups ...string) (interface{}, error) {
	return nil, nil
}

func (d *ApplicationautoscalingDriver) Lookup(lookups ...string) (driverFn driver.DriverFn, err error) {
	return nil, driver.ErrDriverFnNotFound
}
