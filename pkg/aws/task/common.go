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
	//	"encoding/json"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/luyomo/tisample/embed"
	"github.com/luyomo/tisample/pkg/ctxt"
	"github.com/luyomo/tisample/pkg/executor"
	"go.uber.org/zap"
	//	"github.com/luyomo/tisample/pkg/executor"
	//	"strings"
)

type Vpc struct {
	CidrBlock string `json:"CidrBlock"`
	State     string `json:"State"`
	VpcId     string `json:"VpcId"`
	OwnerId   string `json:"OwnerId"`
	Tags      []Tag  `json:"Tags"`
}

type Vpcs struct {
	Vpcs []Vpc `json:"Vpcs"`
}

type ClusterInfo struct {
	cidr                   string
	region                 string
	keyName                string
	keyFile                string
	instanceType           string
	imageId                string
	vpcInfo                Vpc
	privateRouteTableId    string
	publicRouteTableId     string
	privateSecurityGroupId string
	publicSecurityGroupId  string
	privateSubnets         []string
	publicSubnet           string
	pcxTidb2Aurora         string
	excludedAZ             []string
	includedAZ             []string
	enableNAT              string
}

type DBInfo struct {
	DBHost     string
	DBPort     int64
	DBUser     string
	DBPassword string
}

type DMClusterInfo struct {
	Name       string `json:"name"`
	User       string `json:"user"`
	DMVersion  string `json:"version"`
	Path       string `json:"path"`
	PrivateKey string `json:"private_key"`
}

type DMClustersInfo struct {
	Clusters []DMClusterInfo `json:"clusters"`
}

func (v Vpc) String() string {
	return fmt.Sprintf("Cidr: %s, State: %s, VpcId: %s, OwnerId: %s", v.CidrBlock, v.State, v.VpcId, v.OwnerId)
}

func (c ClusterInfo) String() string {
	return fmt.Sprintf("vpcInfo:[%s], privateRouteTableId:%s, publicRouteTableId:%s, privateSecurityGroupId:%s, publicSecurityGroupId:%s, privateSubnets:%s, publicSubnet:%s, pcxTidb2Aurora:%s", c.vpcInfo.String(), c.privateRouteTableId, c.publicRouteTableId, c.privateSecurityGroupId, c.publicSecurityGroupId, strings.Join(c.privateSubnets, ","), c.publicSubnet, c.pcxTidb2Aurora)
}

type Route struct {
	DestinationCidrBlock string `json:"DestinationCidrBlock"`
	TransitGatewayId     string `json:"TransitGatewayId"`
	GatewayId            string `json:"GatewayId"`
	Origin               string `json:"Origin"`
	State                string `json:"State"`
}

type RouteTable struct {
	RouteTableId string  `json:"RouteTableId"`
	Tags         []Tag   `json:"Tags"`
	Routes       []Route `json:"Routes"`
}

type ResultRouteTable struct {
	TheRouteTable RouteTable `json:"RouteTable"`
}

type RouteTables struct {
	RouteTables []RouteTable `json:"RouteTables"`
}

func (r RouteTable) String() string {
	return fmt.Sprintf("RouteTableId:%s", r.RouteTableId)
}

func (r ResultRouteTable) String() string {
	return fmt.Sprintf("RetRouteTable:%s", r.String())
}

func (r RouteTables) String() string {
	var res []string
	for _, route := range r.RouteTables {
		res = append(res, route.String())
	}
	return fmt.Sprintf("RouteTables:%s", strings.Join(res, ","))
}

type IpRanges struct {
	CidrIp string `json:"CidrIp"`
}

type IpPermissions struct {
	FromPort   int        `json:"FromPort"`
	IpProtocol string     `json:"IpProtocol"`
	IpRanges   []IpRanges `json:"IpRanges"`
	ToPort     int        `json:"ToPort"`
}

type SecurityGroups struct {
	SecurityGroups []SecurityGroup `json:"SecurityGroups"`
}

type SecurityGroup struct {
	GroupId       string          `json:"GroupId"`
	GroupName     string          `json:"GroupName"`
	IpPermissions []IpPermissions `json:"IpPermissions"`
	Tags          []Tag           `json:"Tags"`
}

func (s SecurityGroup) String() string {
	return fmt.Sprintf(s.GroupId)
}

func (i SecurityGroups) String() string {
	var res []string
	for _, sg := range i.SecurityGroups {
		res = append(res, sg.String())
	}
	return strings.Join(res, ",")
}

type Attachment struct {
	State string `json:"State"`
	VpcId string `json:"VpcId"`
}

type InternetGateway struct {
	InternetGatewayId string       `json:"InternetGatewayId"`
	Attachments       []Attachment `json:"Attachments"`
}

type InternetGateways struct {
	InternetGateways []InternetGateway `json:"InternetGateways"`
}

type NewInternetGateway struct {
	InternetGateway InternetGateway `json:"InternetGateway"`
}

func (i InternetGateway) String() string {
	return fmt.Sprintf("InternetGatewayId: %s", i.InternetGatewayId)
}

func (i InternetGateways) String() string {
	var res []string
	for _, gw := range i.InternetGateways {
		res = append(res, gw.String())
	}
	return strings.Join(res, ",")
}

func (i NewInternetGateway) String() string {
	return i.InternetGateway.String()
}

type VPCStatus struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type VpcPeer struct {
	VpcPeeringConnectionId string    `json:"VpcPeeringConnectionId"`
	VpcStatus              VPCStatus `json:"Status"`
}

type VpcConnection struct {
	VpcPeeringConnection VpcPeer `json:"VpcPeeringConnection"`
}

type VpcConnections struct {
	VpcPeeringConnections []VpcPeer `json:"VpcPeeringConnections"`
}

type ResourceTag struct {
	clusterName    string
	clusterType    string
	subClusterType string
	port           []int
}

type CreateVpcPeering struct {
	user      string
	host      string
	sourceVPC ResourceTag
	targetVPC ResourceTag
}

type DMTaskDetail struct {
	Result  bool   `json:"result"`
	Msg     string `json:"msg"`
	Sources []struct {
		Result       bool   `json:"result"`
		Msg          string `json:"msg"`
		SourceStatus struct {
			Source      string `json:"source"`
			Worker      string `json:"worker"`
			Result      string `json:"result"`
			RelayStatus string `json:"relayStatus"`
		} `json:"sourceStatus"`
		SubTaskStatus []struct {
			Name                string `json:"name"`
			Stage               string `json:"stage"`
			Unit                string `json:"unit"`
			Result              string `json:"result"`
			UnresolvedDDLLockID string `json:"unresolvedDDLLockID"`
			Sync                struct {
				TotalEvents         string   `json:"totalEvents"`
				TotalTps            string   `json:"totalTps"`
				RecentTps           string   `json:"recentTps"`
				MasterBinlog        string   `json:"masterBinlog"`
				MasterBinlogGtid    string   `json:"masterBinlogGtid"`
				SyncerBinlog        string   `json:"syncerBinlog"`
				SyncerBinlogGtid    string   `json:"syncerBinlogGtid"`
				BlockingDDLs        []string `json:"blockingDDLs"`
				UnresolvedGroups    []string `json:"unresolvedGroups"`
				Synced              bool     `json:"synced"`
				BinlogType          string   `json:"binlogType"`
				SecondsBehindMaster string   `json:"secondsBehindMaster"`
				BlockDDLOwner       string   `json:"blockDDLOwner"`
				ConflictMsg         string   `json:"conflictMsg"`
			} `json:"sync"`
		} `json:"subTaskStatus"`
	} `json:"sources"`
}

type DisplayDMCluster struct {
	ClusterMeta struct {
		ClusterType    string `json:"cluster_type"`
		ClusterName    string `json:"cluster_name"`
		ClusterVersion string `json:"cluster_version"`
		DeployUser     string `json:"deploy_user"`
		SshType        string `json:"ssh_type"`
		TlsEnabled     bool   `json:"tls_enabled"`
	} `json:"cluster_meta"`
	Instances []struct {
		ID            string `json:"id"`
		Role          string `json:"role"`
		Host          string `json:"host"`
		Ports         string `json:"ports"`
		OsArch        string `json:"os_arch"`
		Status        string `json:"status"`
		Since         string `json:"since"`
		DataDir       string `json:"data_dir"`
		DeployDir     string `json:"deploy_dir"`
		ComponentName string `json:"ComponentName"`
		Port          int    `json:"Port"`
	} `json:"instances"`
}

func contains(s *[]map[string]string, str string) bool {
	for _, v := range *s {
		if v["Cluster"] == str {
			return true
		}
	}

	return false
}

func SearchVPCName(executor *ctxt.Executor, ctx context.Context, clusterKeyWord string) (*[]map[string]string, error) {
	stdout, _, err := (*executor).Execute(ctx, fmt.Sprintf("aws ec2 describe-vpcs --filters Name=tag:Cluster,Values=%s ", clusterKeyWord), false)
	if err != nil {
		return nil, err
	}
	var retValue []map[string]string

	var vpcs Vpcs
	if err := json.Unmarshal(stdout, &vpcs); err != nil {
		zap.L().Debug("The error to parse the string ", zap.Error(err))
		return nil, err
	}
	for _, vpc := range vpcs.Vpcs {
		entry := make(map[string]string)
		for _, tag := range vpc.Tags {
			if tag.Key == "Name" {
				entry["Name"] = tag.Value
			}
			if tag.Key == "Type" {
				entry["Type"] = tag.Value
			}
		}
		if !contains(&retValue, entry["Cluster"]) {
			retValue = append(retValue, entry)
		}
	}

	return &retValue, nil

}

func getVPCInfos(executor ctxt.Executor, ctx context.Context, vpc ResourceTag) (*Vpcs, error) {
	stdout, _, err := executor.Execute(ctx, fmt.Sprintf("aws ec2 describe-vpcs --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" ", vpc.clusterName, vpc.clusterType), false)
	if err != nil {
		return nil, err
	}

	var vpcs Vpcs
	if err := json.Unmarshal(stdout, &vpcs); err != nil {
		zap.L().Debug("The error to parse the string ", zap.Error(err))
		return nil, err
	}
	return &vpcs, nil
}

func getVPCInfo(executor ctxt.Executor, ctx context.Context, vpc ResourceTag) (*Vpc, error) {
	stdout, _, err := executor.Execute(ctx, fmt.Sprintf("aws ec2 describe-vpcs --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=tag:Type,Values=%s\"", vpc.clusterName, vpc.clusterType, vpc.subClusterType), false)
	if err != nil {
		return nil, err
	}

	var vpcs Vpcs
	if err := json.Unmarshal(stdout, &vpcs); err != nil {
		zap.L().Debug("The error to parse the string ", zap.Error(err))
		return nil, err
	}
	if len(vpcs.Vpcs) > 1 {
		return nil, errors.New("Multiple VPC found")
	}

	if len(vpcs.Vpcs) == 0 {
		return nil, errors.New("No VPC found")
	}
	return &(vpcs.Vpcs[0]), nil
}

func getNetworks(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, subClusterType, scope string) (*[]Subnet, error) {
	//command := fmt.Sprintf("aws ec2 describe-subnets --filters \"Name=tag-key,Values=Name\" \"Name=tag-value,Values=%s\" \"Name=tag-key,Values=Cluster\" \"Name=tag-value,Values=%s\" \"Name=tag-key,Values=Type\" \"Name=tag-value,Values=%s\" \"Name=tag-key,Values=Scope\" \"Name=tag-value,Values=%s\"", clusterName, clusterType, subClusterType, scope)
	command := fmt.Sprintf("aws ec2 describe-subnets --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=tag:Type,Values=%s\"", clusterName, clusterType, subClusterType)
	zap.L().Debug("Command", zap.String("describe-subnets", command))
	stdout, _, err := executor.Execute(ctx, command, false)
	if err != nil {
		return nil, err
	}

	var subnets Subnets
	if err = json.Unmarshal(stdout, &subnets); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("subnets", string(stdout)))
		return nil, err
	}
	return &subnets.Subnets, nil
}

func getNetworksString(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, subClusterType, scope string) (string, error) {
	subnets, err := getNetworks(executor, ctx, clusterName, clusterType, subClusterType, scope)
	if err != nil {
		return "", err
	}
	if subnets == nil {
		return "", errors.New("No subnets found")
	}
	var arrSubnets []string
	for _, subnet := range *subnets {
		arrSubnets = append(arrSubnets, "\""+subnet.SubnetId+"\"")
	}
	return "[" + strings.Join(arrSubnets, ",") + "]", nil

}

func getTransitGateway(executor ctxt.Executor, ctx context.Context, clusterName, clusterType string) (*TransitGateway, error) {
	command := fmt.Sprintf("aws ec2 describe-transit-gateways --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=state,Values=available,modifying,pending\"", clusterName, clusterType)
	stdout, _, err := executor.Execute(ctx, command, false)
	if err != nil {
		return nil, err
	} else {
		var transitGateways TransitGateways
		if err = json.Unmarshal(stdout, &transitGateways); err != nil {
			return nil, err
		}
		for _, transitGateway := range transitGateways.TransitGateways {
			return &transitGateway, nil
		}
	}
	return nil, nil
}

func getRouteTable(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, subClusterType string) (*RouteTable, error) {
	stdout, _, err := executor.Execute(ctx, fmt.Sprintf("aws ec2 describe-route-tables --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=tag:Type,Values=%s\" ", clusterName, clusterType, subClusterType), false)
	if err != nil {
		return nil, err
	}

	var routeTables RouteTables
	if err = json.Unmarshal(stdout, &routeTables); err != nil {
		zap.L().Error("Failed to parse the route table", zap.String("describe-route-table", string(stdout)))
		return nil, err
	}

	zap.L().Debug("Print the route tables", zap.String("routeTables", routeTables.String()))
	if len(routeTables.RouteTables) == 0 {
		return nil, errors.New("No route table found")
	}
	if len(routeTables.RouteTables) > 1 {
		return nil, errors.New("Multiple route tables found")
	}
	return &routeTables.RouteTables[0], nil
}

func getRouteTableByVPC(executor ctxt.Executor, ctx context.Context, clusterName, vpcID string) (*RouteTable, error) {
	stdout, _, err := executor.Execute(ctx, fmt.Sprintf("aws ec2 describe-route-tables --filters \"Name=tag:Name,Values=%s\" \"Name=vpc-id,Values=%s\" ", clusterName, vpcID), false)
	if err != nil {
		return nil, err
	}

	var routeTables RouteTables
	if err = json.Unmarshal(stdout, &routeTables); err != nil {
		zap.L().Error("Failed to parse the route table", zap.String("describe-route-table", string(stdout)))
		return nil, err
	}

	zap.L().Debug("Print the route tables", zap.String("routeTables", routeTables.String()))
	if len(routeTables.RouteTables) == 0 {
		return nil, errors.New("No route table found")
	}
	if len(routeTables.RouteTables) > 1 {
		return nil, errors.New("Multiple route tables found")
	}
	return &routeTables.RouteTables[0], nil
}

func getWorkstation(executor ctxt.Executor, ctx context.Context, clusterName, clusterType string) (*EC2, error) {
	command := fmt.Sprintf("aws ec2 describe-instances --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=tag:Type,Values=%s\" \"Name=instance-state-code,Values=16\"", clusterName, clusterType, "workstation")
	zap.L().Debug("Command", zap.String("describe-instance", command))
	stdout, _, err := executor.Execute(ctx, command, false)
	if err != nil {
		return nil, err
	}

	var reservations Reservations
	if err = json.Unmarshal(stdout, &reservations); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("describe-instances", string(stdout)))
		return nil, err
	}

	var theInstance EC2
	cntInstance := 0
	for _, reservation := range reservations.Reservations {
		for _, instance := range reservation.Instances {
			cntInstance++
			theInstance = instance
		}
	}

	if cntInstance > 1 {
		return nil, errors.New("Multiple workstation nodes")
	}
	if cntInstance == 0 {
		return nil, errors.New("No workstation node")
	}

	return &theInstance, nil
}

func GetWSExecutor(texecutor ctxt.Executor, ctx context.Context, clusterName, clusterType, user, keyFile string) (*ctxt.Executor, error) {
	workstation, err := getWorkstation(texecutor, ctx, clusterName, clusterType)
	if err != nil {
		return nil, err
	}

	wsexecutor, err := executor.New(executor.SSHTypeSystem, false, executor.SSHConfig{Host: workstation.PublicIpAddress, User: user, KeyFile: keyFile}, []string{})
	if err != nil {
		return nil, err
	}
	//lsb_release --id
	return &wsexecutor, nil
}

func containString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func installPKGs(wsexecutor *ctxt.Executor, ctx context.Context, packages []string) error {
	stdout, _, err := (*wsexecutor).Execute(ctx, "lsb_release --id", true)
	if err != nil {
		return err
	}
	osVersion := strings.Split(string(stdout), ":")

	for _, pkg := range packages {
		if containString([]string{"Debian"}, strings.TrimSpace(osVersion[1])) {
			if _, _, err := (*wsexecutor).Execute(ctx, fmt.Sprintf("apt-get install -y %s", pkg), true); err != nil {
				return err
			}
		} else {
			if _, _, err := (*wsexecutor).Execute(ctx, fmt.Sprintf("yum install -y %s", pkg), true); err != nil {
				return err
			}
		}

	}
	return nil

}

func getTiDBClusterInfo(wsexecutor *ctxt.Executor, ctx context.Context, clusterName string) (*TiDBClusterDetail, error) {

	stdout, _, err := (*wsexecutor).Execute(ctx, fmt.Sprintf(`/home/admin/.tiup/bin/tiup cluster display %s --format json `, clusterName), false)
	if err != nil {
		return nil, err
	}

	var tidbClusterDetail TiDBClusterDetail
	if err = json.Unmarshal(stdout, &tidbClusterDetail); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("tidb cluster list", string(stdout)))
		return nil, err
	}

	return &tidbClusterDetail, nil
}

func getDMClusterInfo(wsexecutor *ctxt.Executor, ctx context.Context, clusterName string) (*DMClusterInfo, error) {

	stdout, _, err := (*wsexecutor).Execute(ctx, fmt.Sprintf(`/home/admin/.tiup/bin/tiup dm list --format json`), false)
	if err != nil {
		return nil, err
	}

	var dmClustersInfo DMClustersInfo
	if err = json.Unmarshal(stdout, &dmClustersInfo); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("tidb cluster list", string(stdout)))
		return nil, err
	}

	for _, clusterInfo := range dmClustersInfo.Clusters {
		if clusterInfo.Name == clusterName {
			return &clusterInfo, nil
		}
	}

	return nil, nil
}

func getEC2Nodes(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, componentName string) (*[]EC2, error) {
	var reservations Reservations
	command := fmt.Sprintf("aws ec2 describe-instances --filters \"Name=tag:Name,Values=%s\" \"Name=tag:Cluster,Values=%s\" \"Name=tag:Component,Values=%s\" \"Name=instance-state-code,Values=0,16,32,64,80\"", clusterName, clusterType, componentName)
	zap.L().Debug("Command", zap.String("describe-instance", command))
	stdout, _, err := executor.Execute(ctx, command, false)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(stdout, &reservations); err != nil {
		zap.L().Debug("Json unmarshal", zap.String("describe-instances", string(stdout)))
		return nil, err
	}

	var theEC2s []EC2
	for _, reservation := range reservations.Reservations {
		for _, instance := range reservation.Instances {
			theEC2s = append(theEC2s, instance)
		}
	}

	return &theEC2s, nil

}

// func deploy(executor ctxt.Executor, ctx context.Context, host string, port int) error {
// 	deployFreetds(executor, ctx, "REPLICA", host, port)
// 	return nil
// }

func deployFreetds(executor ctxt.Executor, ctx context.Context, name, host string, port int) error {

	if err := installPKGs(&executor, ctx, []string{"freetds-bin"}); err != nil {
		return err
	}

	fdFile, err := os.Create(fmt.Sprintf("/tmp/%s", "freetds.conf"))
	if err != nil {
		return err
	}
	defer fdFile.Close()

	fp := path.Join("templates", "config", fmt.Sprintf("%s.tpl", "freetds.conf"))
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return err
	}

	tmpl, err := template.New("test").Parse(string(tpl))
	if err != nil {
		return err
	}

	var tplData TplSQLServer
	tplData.Name = name
	tplData.Host = host
	tplData.Port = port
	if err := tmpl.Execute(fdFile, tplData); err != nil {
		return err
	}

	err = executor.Transfer(ctx, fmt.Sprintf("/tmp/%s", "freetds.conf"), "/tmp/freetds.conf", false, 0)
	if err != nil {
		return err
	}

	command := fmt.Sprintf(`mv /tmp/freetds.conf /etc/freetds/`)
	_, _, err = executor.Execute(ctx, command, true)
	if err != nil {
		return err
	}

	return nil
}

/************************** The function for the [][]string sort **************/
type byComponentNameZone [][]string

func (items byComponentNameZone) Len() int      { return len(items) }
func (items byComponentNameZone) Swap(i, j int) { items[i], items[j] = items[j], items[i] }
func (items byComponentNameZone) Less(i, j int) bool {
	if items[i][0] < items[j][0] {
		return true
	}
	if items[i][0] == items[j][0] && items[i][1] < items[j][1] {
		return true
	}
	return false
}

type byComponentName [][]string

func (items byComponentName) Len() int      { return len(items) }
func (items byComponentName) Swap(i, j int) { items[i], items[j] = items[j], items[i] }
func (items byComponentName) Less(i, j int) bool {
	if items[i][0] < items[j][0] {
		return true
	}
	return false
}

type TargetGroups struct {
	TargetGroups []TargetGroup `json:"TargetGroups"`
}

type TargetGroup struct {
	TargetGroupArn  string `json:"TargetGroupArn"`
	TargetGroupName string `json:"TargetGroupName"`
	Protocol        string `json:"Protocol"`
	Port            int    `json:"Port"`
	VpcId           string `json:"VpcId"`
	TargetType      string `json:"TargetType"`
}

type TagDescription struct {
	Tags []Tag `json:"Tags"`
}

type TagDescriptions struct {
	TagDescriptions []TagDescription `json:"TagDescriptions"`
}

func ExistsELBResource(executor ctxt.Executor, ctx context.Context, clusterType, subClusterType, clusterName, resourceName string) bool {
	command := fmt.Sprintf("aws elbv2 describe-tags --resource-arns %s ", resourceName)
	stdout, _, err := executor.Execute(ctx, command, false)
	if err != nil {
		return false
	}

	var tagDescriptions TagDescriptions
	if err = json.Unmarshal(stdout, &tagDescriptions); err != nil {
		return false
	}
	matchedCnt := 0
	for _, tagDescription := range tagDescriptions.TagDescriptions {
		for _, tag := range tagDescription.Tags {
			if tag.Key == "Cluster" && tag.Value == clusterType {
				matchedCnt++
			}
			if tag.Key == "Type" && tag.Value == subClusterType {
				matchedCnt++
			}
			if tag.Key == "Name" && tag.Value == clusterName {
				matchedCnt++
			}
			if matchedCnt == 3 {
				return true
			}
		}
	}
	return false
}

func getTargetGroup(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, subClusterType string) (*TargetGroup, error) {
	command := fmt.Sprintf("aws elbv2 describe-target-groups --name \"%s\"", clusterName)
	stdout, stderr, err := executor.Execute(ctx, command, false)
	if err != nil {
		if strings.Contains(string(stderr), "One or more target groups not found") {
			return nil, errors.New("No target group found")
		} else {
			return nil, err
		}
	}
	var targetGroups TargetGroups
	if err = json.Unmarshal(stdout, &targetGroups); err != nil {
		return nil, err
	}

	for _, targetGroup := range targetGroups.TargetGroups {
		if existsResource := ExistsELBResource(executor, ctx, clusterType, subClusterType, clusterName, targetGroup.TargetGroupArn); existsResource == true {
			return &targetGroup, nil
		}
	}
	return nil, errors.New("No target group found")
}

type LoadBalancer struct {
	LoadBalancerArn  string `json:"LoadBalancerArn"`
	DNSName          string `json:"DNSName"`
	LoadBalancerName string `json:"LoadBalancerName"`
	Scheme           string `json:"Scheme"`
	VpcId            string `json:"VpcId"`
	State            struct {
		Code string `json:"Code"`
	} `json:"State"`
	Type string `json:"Type"`
}

type LoadBalancers struct {
	LoadBalancers []LoadBalancer `json:"LoadBalancers"`
}

func getNLB(executor ctxt.Executor, ctx context.Context, clusterName, clusterType, subClusterType string) (*LoadBalancer, error) {
	command := fmt.Sprintf("aws elbv2 describe-load-balancers --name \"%s\"", clusterName)
	stdout, stderr, err := executor.Execute(ctx, command, false)
	if err != nil {
		if strings.Contains(string(stderr), fmt.Sprintf("Load balancers '[%s]' not found", clusterName)) {
			return nil, errors.New("No NLB found")
		} else {
			return nil, err
		}
	}
	var loadBalancers LoadBalancers
	if err = json.Unmarshal(stdout, &loadBalancers); err != nil {
		return nil, err
	}

	for _, loadBalancer := range loadBalancers.LoadBalancers {
		if existsResource := ExistsELBResource(executor, ctx, clusterType, subClusterType, clusterName, loadBalancer.LoadBalancerArn); existsResource == true {
			return &loadBalancer, nil
		}
	}
	return nil, errors.New("No NLB found")
}

func installWebSSH2(wexecutor *ctxt.Executor, ctx context.Context) error {

	if err := installPKGs(wexecutor, ctx, []string{"nodejs", "npm", "cmake"}); err != nil {
		return err
	}

	commands := []string{"[ -d /opt/webssh2 ] || git clone https://github.com/billchurch/webssh2.git /opt/webssh2", "npm install /opt/webssh2/app"}

	for _, command := range commands {
		_, _, err := (*wexecutor).Execute(ctx, command, true)
		if err != nil {
			return err
		}
	}

	err := (*wexecutor).Transfer(ctx, "embed/templates/systemd/webssh2.service", "/tmp/", false, 0)
	if err != nil {
		return err
	}

	_, _, err = (*wexecutor).Execute(ctx, "mv /tmp/webssh2.service /etc/systemd/system/", true)
	if err != nil {
		return err
	}

	return nil
}

func containsInArray(s []string, searchterm string) bool {

	if len(s) == 0 {
		return false
	}
	i := sort.SearchStrings(s, searchterm)

	return i < len(s) && s[i] == searchterm
}

type DBConnectInfo struct {
	DBHost     string `yaml:"Host"`
	DBPort     int    `yaml:"Port"`
	DBUser     string `yaml:"User"`
	DBPassword string `yaml:"Password"`
}

func ReadTiDBConntionInfo(workstation *ctxt.Executor, fileName string) (*DBConnectInfo, error) {

	// 02. Get the TiDB connection info
	// if err := (*workstation).Transfer(context.Background(), fmt.Sprintf("/opt/tidb-db-info.yml"), "/tmp/tidb-db-info.yml", true, 1024); err != nil {
	if err := (*workstation).Transfer(context.Background(), fmt.Sprintf("/opt/%s", fileName), fmt.Sprintf("/tmp/%s", fileName), true, 1024); err != nil {
		return nil, err
	}

	dbConnectInfo := DBConnectInfo{}

	yfile, err := ioutil.ReadFile(fmt.Sprintf("/tmp/%s", fileName))
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(yfile, &dbConnectInfo); err != nil {
		return nil, err
	}

	return &dbConnectInfo, nil
}
func TransferToWorkstation(workstation *ctxt.Executor, sourceFile, destFile, mode string, params interface{}) error {

	ctx := context.Background()

	err := (*workstation).TransferTemplate(ctx, sourceFile, fmt.Sprintf("/tmp/%s", "test.file"), mode, params, true, 0)
	if err != nil {
		return err
	}

	if _, _, err := (*workstation).Execute(ctx, fmt.Sprintf("mv /tmp/%s %s", "test.file", destFile), true); err != nil {
		return err
	}

	return nil

}

func ParseRangeData(inputData string) (*[]int, error) {

	var varRet []int

	numberPattern := regexp.MustCompile(`^\d+$`)
	rangePattern := regexp.MustCompile(`^(\d+)-(\d+)/(\d+)$`)

	match := numberPattern.MatchString(inputData)
	if match == false {
		dataRange := rangePattern.FindStringSubmatch(inputData)

		if dataRange == nil {
			return nil, errors.New("Not match user num pattern")
		}
		num, _ := strconv.Atoi(dataRange[1])
		endNum, _ := strconv.Atoi(dataRange[2])
		interval, _ := strconv.Atoi(dataRange[3])

		for {
			if num > endNum {
				break
			}

			varRet = append(varRet, num)
			num += interval
		}

	} else {
		intNum, err := strconv.Atoi(inputData)
		if err != nil {
			return nil, err
		}
		varRet = append(varRet, intNum)
	}

	return &varRet, nil
}
