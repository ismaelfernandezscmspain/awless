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
	"time"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/wallix/awless/logger"
)

type CreateDatabase struct {
	_                 string `action:"create" entity:"database" awsAPI:"rds" awsCall:"CreateDBInstance" awsInput:"rds.CreateDBInstanceInput" awsOutput:"rds.CreateDBInstanceOutput"`
	logger            *logger.Logger
	api               rdsiface.RDSAPI
	Type              *string   `awsName:"DBInstanceClass" awsType:"awsstr" templateName:"type" required:""`
	Id                *string   `awsName:"DBInstanceIdentifier" awsType:"awsstr" templateName:"id" required:""`
	Engine            *string   `awsName:"Engine" awsType:"awsstr" templateName:"engine" required:""`
	Password          *string   `awsName:"MasterUserPassword" awsType:"awsstr" templateName:"password" required:""`
	Username          *string   `awsName:"MasterUsername" awsType:"awsstr" templateName:"username" required:""`
	Size              *int64    `awsName:"AllocatedStorage" awsType:"awsint64" templateName:"size" required:""`
	Autoupgrade       *bool     `awsName:"AutoMinorVersionUpgrade" awsType:"awsbool" templateName:"autoupgrade"`
	Availabilityzone  *string   `awsName:"AvailabilityZone" awsType:"awsstr" templateName:"availabilityzone"`
	Backupretention   *int64    `awsName:"BackupRetentionPeriod" awsType:"awsint64" templateName:"backupretention"`
	Cluster           *string   `awsName:"DBClusterIdentifier" awsType:"awsstr" templateName:"cluster"`
	Dbname            *string   `awsName:"DBName" awsType:"awsstr" templateName:"dbname"`
	Parametergroup    *string   `awsName:"DBParameterGroupName" awsType:"awsstr" templateName:"parametergroup"`
	Dbsecuritygroups  []*string `awsName:"DBSecurityGroups" awsType:"awsstringslice" templateName:"dbsecuritygroups"`
	Subnetgroup       *string   `awsName:"DBSubnetGroupName" awsType:"awsstr" templateName:"subnetgroup"`
	Domain            *string   `awsName:"Domain" awsType:"awsstr" templateName:"domain"`
	Iamrole           *string   `awsName:"DomainIAMRoleName" awsType:"awsstr" templateName:"iamrole"`
	Version           *string   `awsName:"EngineVersion" awsType:"awsstr" templateName:"version"`
	Iops              *int64    `awsName:"Iops" awsType:"awsint64" templateName:"iops"`
	License           *string   `awsName:"LicenseModel" awsType:"awsstr" templateName:"license"`
	Multiaz           *bool     `awsName:"MultiAZ" awsType:"awsbool" templateName:"multiaz"`
	Optiongroup       *string   `awsName:"OptionGroupName" awsType:"awsstr" templateName:"optiongroup"`
	Port              *int64    `awsName:"Port" awsType:"awsint64" templateName:"port"`
	Backupwindow      *string   `awsName:"PreferredBackupWindow" awsType:"awsstr" templateName:"backupwindow"`
	Maintenancewindow *string   `awsName:"PreferredMaintenanceWindow" awsType:"awsstr" templateName:"maintenancewindow"`
	Public            *bool     `awsName:"PubliclyAccessible" awsType:"awsbool" templateName:"public"`
	Encrypted         *bool     `awsName:"StorageEncrypted" awsType:"awsbool" templateName:"encrypted"`
	Storagetype       *string   `awsName:"StorageType" awsType:"awsstr" templateName:"storagetype"`
	Timezone          *string   `awsName:"Timezone" awsType:"awsstr" templateName:"timezone"`
	Vpcsecuritygroups []*string `awsName:"VpcSecurityGroupIds" awsType:"awsstringslice" templateName:"vpcsecuritygroups"`
}

func (cmd *CreateDatabase) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *CreateDatabase) ExtractResult(i interface{}) string {
	return awssdk.StringValue(i.(*rds.CreateDBInstanceOutput).DBInstance.DBInstanceIdentifier)
}

type DeleteDatabase struct {
	_            string `action:"delete" entity:"database" awsAPI:"rds" awsCall:"DeleteDBInstance" awsInput:"rds.DeleteDBInstanceInput" awsOutput:"rds.DeleteDBInstanceOutput"`
	logger       *logger.Logger
	api          rdsiface.RDSAPI
	Id           *string `awsName:"DBInstanceIdentifier" awsType:"awsstr" templateName:"id" required:""`
	SkipSnapshot *bool   `awsName:"SkipFinalSnapshot" awsType:"awsbool" templateName:"skip-snapshot"`
	Snapshot     *string `awsName:"FinalDBSnapshotIdentifier" awsType:"awsstr" templateName:"snapshot"`
}

func (cmd *DeleteDatabase) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

type CheckDatabase struct {
	_       string `action:"check" entity:"database" awsAPI:"rds"`
	logger  *logger.Logger
	api     rdsiface.RDSAPI
	Id      *string `templateName:"id" required:""`
	State   *string `templateName:"state" required:""`
	Timeout *int64  `templateName:"timeout" required:""`
}

func (cmd *CheckDatabase) ValidateParams(params []string) ([]string, error) {
	return validateParams(cmd, params)
}

func (cmd *CheckDatabase) ValidateState() error {
	return NewEnumValidator("available",
		"backing-up",
		"creating",
		"deleting",
		"failed",
		"maintenance",
		"modifying",
		"rebooting",
		"renaming",
		"resetting-master-credentials",
		"restore-error",
		"storage-full",
		"upgrading",
		notFoundState).Validate(cmd.State)
}

func (cmd *CheckDatabase) ManualRun(ctx map[string]interface{}) (interface{}, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: cmd.Id,
	}

	c := &checker{
		description: fmt.Sprintf("database %s", StringValue(cmd.Id)),
		timeout:     time.Duration(Int64AsIntValue(cmd.Timeout)) * time.Second,
		frequency:   5 * time.Second,
		fetchFunc: func() (string, error) {
			output, err := cmd.api.DescribeDBInstances(input)
			if err != nil {
				if awserr, ok := err.(awserr.Error); ok {
					if awserr.Code() == "DatabaseNotFound" {
						return notFoundState, nil
					}
				} else {
					return "", err
				}
			} else {
				if res := output.DBInstances; len(res) > 0 {
					for _, dbinst := range res {
						if StringValue(dbinst.DBInstanceIdentifier) == StringValue(cmd.Id) {
							return StringValue(dbinst.DBInstanceStatus), nil
						}
					}
				}
			}
			return notFoundState, nil
		},
		expect: StringValue(cmd.State),
		logger: cmd.logger,
	}
	return nil, c.check()
}
