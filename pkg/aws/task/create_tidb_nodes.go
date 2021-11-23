// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/luyomo/tisample/pkg/executor"
	"github.com/luyomo/tisample/pkg/aws/spec"
	"go.uber.org/zap"
)

type CreateTiDBNodes struct {
	user           string
	host           string
	awsTopoConfigs *spec.AwsTopoConfigs
	clusterName    string
	clusterType    string
}

// Execute implements the Task interface
func (c *CreateTiDBNodes) Execute(ctx context.Context) error {
	local, err := executor.New(executor.SSHTypeNone, false, executor.SSHConfig{Host: "127.0.0.1", User: c.user})
	if err != nil {
		return nil
	}

	if c.awsTopoConfigs.TiDB.Count == 0 {
		zap.L().Debug("There is no TiDB nodes to be configured")
		return nil
	}

	// 3. Fetch the count of instance from the instance
	command := fmt.Sprintf("aws ec2 describe-instances --filters \"Name=tag-key,Values=Name\" \"Name=tag-value,Values=%s\" \"Name=tag-key,Values=Type\" \"Name=tag-value,Values=%s\" \"Name=tag-key,Values=Component\" \"Name=tag-value,Values=tidb\" \"Name=instance-state-code,Values=0,16,32,64,80\"", c.clusterName, c.clusterType)
	zap.L().Debug("Command", zap.String("describe-instances", command))
	stdout, _, err := local.Execute(ctx, command, false)
	if err != nil {
		return nil
	}

	var reservations Reservations
	if err = json.Unmarshal(stdout, &reservations); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("describe-instances", string(stdout)))
		return nil
	}

	existsNodes := 0
	for _, reservation := range reservations.Reservations {
		existsNodes = existsNodes + len(reservation.Instances)
	}

	for _idx := 0; _idx < c.awsTopoConfigs.TiDB.Count-existsNodes; _idx++ {
		command := fmt.Sprintf("aws ec2 run-instances --count 1 --image-id %s --instance-type %s --key-name %s --security-group-ids %s --subnet-id %s --region %s  --tag-specifications \"ResourceType=instance,Tags=[{Key=Name,Value=%s},{Key=Type,Value=%s},{Key=Component,Value=tidb}]\"", c.awsTopoConfigs.General.ImageId, c.awsTopoConfigs.TiDB.InstanceType, c.awsTopoConfigs.General.KeyName, clusterInfo.privateSecurityGroupId, clusterInfo.privateSubnets[_idx%len(clusterInfo.privateSubnets)], c.awsTopoConfigs.General.Region, c.clusterName, c.clusterType)
		zap.L().Debug("Command", zap.String("run-instances", command))
		stdout, _, err = local.Execute(ctx, command, false)
		if err != nil {
			return nil
		}
	}
	return nil
}

// Rollback implements the Task interface
func (c *CreateTiDBNodes) Rollback(ctx context.Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *CreateTiDBNodes) String() string {
	return fmt.Sprintf("Echo: host=%s ", c.host)
}
