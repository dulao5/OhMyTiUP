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

package command

import (
	//"context"
	"fmt"
	"os"
	"path"

	"github.com/luyomo/tisample/pkg/aws/manager"
	"github.com/luyomo/tisample/pkg/aws/spec"
	//	"github.com/luyomo/tisample/pkg/aws/task"
	operator "github.com/luyomo/tisample/pkg/aws/operation"
	"github.com/luyomo/tisample/pkg/set"
	"github.com/luyomo/tisample/pkg/tui"
	"github.com/luyomo/tisample/pkg/utils"
	perrs "github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

func newAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aurora2tidbcloud <sub_command>",
		Short: "Data migration from aurora 2 tidb cloud",
	}

	cmd.AddCommand(
		newAurora2TiDBCloudDeploy(),
		newListAurora2TiDBCloudCmd(),
		newDestroyAurora2TiDBCloudCmd(),
		newVPCPeeringAurora2TiDBCloudCmd(),
		newVPCPeeringAcceptAurora2TiDBCloudCmd(),
		newStartSyncAurora2TiDBCloudCmd(),
		newQuerySyncStatusAurora2TiDBCloudCmd(),
		newStopSyncTaskAurora2TiDBCloudCmd(),
		newAurora2TiDBCloudMeasurementCmd(),
		newAurora2TiDBCloudDataDiffCmd(),
	)
	return cmd
}

func newAurora2TiDBCloudDeploy() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "deploy <cluster-name> <topology.yaml>",
		Short:        "Deploy an aurora to TiDB Cloud migration demo",
		Long:         "Deploy an aurora for demo.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 2)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]
			topoFile := args[1]
			if data, err := os.ReadFile(topoFile); err == nil {
				teleTopology = string(data)
			}
			fmt.Printf("The command here is %v \n", teleCommand)

			return cm.Aurora2TiDBCloudDeploy(clusterName, topoFile, skipConfirm, gOpt)
		},
	}

	return cmd
}

func newListAurora2TiDBCloudCmd() *cobra.Command {
	opt := manager.DeployOptions{
		IdentityFile: path.Join(utils.UserHome(), ".ssh", "id_rsa"),
	}
	cmd := &cobra.Command{
		Use:   "list <cluster-name>",
		Short: "List all clusters or cluster of aurora db",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.ListAurora2TiDBCloudCluster(clusterName, opt)
		},
	}

	cmd.Flags().StringVarP(&opt.User, "user", "u", utils.CurrentUser(), "The user name to login via SSH. The user must has root (or sudo) privilege.")

	//	fmt.Printf("The option is <%#v> \n", opt)
	return cmd
}

func newVPCPeeringAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-vpc-peering <cluster-name>",
		Short: "Show the vpc peering between DM and TiDB Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.ShowVPCPeering(clusterName, "ohmytiup-aurora2tidbcloud", []string{"workstation", "aurora", "dm"})
		},
	}

	return cmd
}

func newVPCPeeringAcceptAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept-vpc-peering <cluster-name> ",
		Short: "Accept the vpc peering between DM and TiDB Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.AcceptVPCPeering(clusterName, "ohmytiup-aurora2tidbcloud", []string{"workstation", "aurora", "dm"})
		},
	}

	return cmd
}

func newStartSyncAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-sync <cluster-name>",
		Short: "Create the DM's source and task",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.StartSyncAurora2TiDBCloudCluster(clusterName, gOpt)
		},
	}

	return cmd
}

func newQuerySyncStatusAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-status <cluster-name>",
		Short: "Create the DM's source and task",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.QuerySyncStatusAurora2TiDBCloudCluster(clusterName, gOpt)
		},
	}

	return cmd
}

func newStopSyncTaskAurora2TiDBCloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-task <cluster-name>",
		Short: "Stop DM's task",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.StopSyncTaskAurora2TiDBCloudCluster(clusterName, gOpt)
		},
	}

	return cmd
}

func newDestroyAurora2TiDBCloudCmd() *cobra.Command {
	destroyOpt := operator.Options{}
	cmd := &cobra.Command{
		Use:   "destroy <cluster-name>",
		Short: "Destroy a specified cluster",
		Long: `Destroy a specified cluster, which will clean the deployment binaries and data.
You can retain some nodes and roles data when destroy cluster, eg:

  $ tiup cluster destroy <cluster-name> --retain-role-data prometheus
  $ tiup cluster destroy <cluster-name> --retain-node-data 172.16.13.11:9000
  $ tiup cluster destroy <cluster-name> --retain-node-data 172.16.13.12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}

			clusterName := args[0]
			clusterReport.ID = scrubClusterName(clusterName)
			teleCommand = append(teleCommand, scrubClusterName(clusterName))

			// Validate the retained roles to prevent unexpected deleting data
			if len(destroyOpt.RetainDataRoles) > 0 {
				validRoles := set.NewStringSet(spec.AllComponentNames()...)
				for _, role := range destroyOpt.RetainDataRoles {
					if !validRoles.Exist(role) {
						return perrs.Errorf("role name `%s` invalid", role)
					}
				}
			}

			return cm.DestroyAurora2TiDBCloudCluster(clusterName, gOpt, destroyOpt, skipConfirm)
		},
	}

	cmd.Flags().StringArrayVar(&destroyOpt.RetainDataNodes, "retain-node-data", nil, "Specify the nodes or hosts whose data will be retained")
	cmd.Flags().StringArrayVar(&destroyOpt.RetainDataRoles, "retain-role-data", nil, "Specify the roles whose data will be retained")
	cmd.Flags().BoolVar(&destroyOpt.Force, "force", false, "Force will ignore remote error while destroy the cluster")

	return cmd
}

func newAurora2TiDBCloudMeasurementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "measure-latency <sub_command>",
		Short: "Run measure latency against tidb",
	}

	cmd.AddCommand(
		newAurora2TiDBCloudMeasurementPrepareCmd(),
		newAurora2TiDBCloudMeasurementRunCmd(),
		newAurora2TiDBCloudMeasurementRunTiDBCloudCmd(),
	)
	return cmd
}

func newAurora2TiDBCloudMeasurementPrepareCmd() *cobra.Command {

	opt := operator.LatencyWhenBatchOptions{
		TiKVMode: "simple",
	}

	cmd := &cobra.Command{
		Use:   "prepare <cluster-name>",
		Short: "Prepare resource for test",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.Aurora2TiDBCloudPrepareCluster(clusterName, opt, gOpt)
		},
	}

	cmd.Flags().IntVar(&opt.SysbenchNumTables, "sysbench-num-tables", 8, "sysbench: --tables")
	cmd.Flags().IntVar(&opt.SysbenchNumRows, "sysbench-num-rows", 10, "sysbench: --table-size")
	cmd.Flags().StringVarP(&opt.SysbenchDBName, "sysbench-db-name", "d", "sbtest", "sysbench: database-name")
	cmd.Flags().StringVarP(&opt.SysbenchPluginName, "sysbench-plugin-name", "p", "oltp_point_select", "sysbench: oltp_point_select")

	cmd.Flags().Int64Var(&opt.SysbenchExecutionTime, "sysbench-execution-time", 600, "sysbench: --execution-time")
	cmd.Flags().IntVar(&opt.SysbenchThread, "sysbench-thread", 4, "sysbench: --thread")
	cmd.Flags().IntVar(&opt.SysbenchReportInterval, "sysbench-report-interval", 10, "sysbench: --report-interval")

	return cmd
}

func newAurora2TiDBCloudMeasurementRunCmd() *cobra.Command {

	opt := operator.LatencyWhenBatchOptions{
		TransInterval: 2,
	}

	cmd := &cobra.Command{
		Use:   "run <cluster-name>",
		Short: "Run the query for latency performance test",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.Aurora2TiDBCloudRunCluster(clusterName, opt, gOpt)
		},
	}

	cmd.Flags().IntVar(&opt.SysbenchNumTables, "sysbench-num-tables", 8, "sysbench: --tables")
	cmd.Flags().IntVar(&opt.SysbenchNumRows, "sysbench-num-rows", 10000, "sysbench: --table-size")
	cmd.Flags().StringVarP(&opt.SysbenchTargetInstance, "sysbench-db-cluster", "t", "Aurora", "sysbench target: TiDBCloud or Aurora")
	cmd.Flags().StringVarP(&opt.SysbenchPluginName, "sysbench-plugin-name", "p", "tidb_oltp_insert_simple", "sysbench: oltp_point_select")
	cmd.Flags().Int64Var(&opt.SysbenchExecutionTime, "sysbench-execution-time", 600, "sysbench: --execution-time")

	return cmd
}

func newAurora2TiDBCloudMeasurementRunTiDBCloudCmd() *cobra.Command {

	opt := operator.LatencyWhenBatchOptions{}

	cmd := &cobra.Command{
		Use:   "run-tidbcloud <cluster-name>",
		Short: "Run the query for latency performance test",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.Aurora2TiDBCloudRunTiDBCloudCluster(clusterName, opt, gOpt)
		},
	}

	cmd.Flags().IntVar(&opt.SysbenchNumTables, "sysbench-num-tables", 8, "sysbench: --tables")
	cmd.Flags().IntVar(&opt.SysbenchNumRows, "sysbench-num-rows", 10000, "sysbench: --table-size")
	cmd.Flags().StringVarP(&opt.SysbenchPluginName, "sysbench-plugin-name", "p", "tidb_oltp_insert_simple", "sysbench: oltp_point_select")
	cmd.Flags().Int64Var(&opt.SysbenchExecutionTime, "sysbench-execution-time", 600, "sysbench: --execution-time")

	return cmd
}

func newAurora2TiDBCloudDataDiffCmd() *cobra.Command {

	opt := operator.LatencyWhenBatchOptions{
		TransInterval: 2,
	}

	cmd := &cobra.Command{
		Use:   "run <cluster-name>",
		Short: "Diff the data between TiDB Cloud and aurora",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.Aurora2TiDBCloudRunCluster(clusterName, opt, gOpt)
		},
	}

	return cmd
}

func newAurora2TiDBCloudExportCmd() *cobra.Command {

	// Buckname/
	opt := operator.LatencyWhenBatchOptions{
		TransInterval: 2,
	}

	cmd := &cobra.Command{
		Use:   "export-aurora <cluster-name>",
		Short: "Export aurora data to S3 ",
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := tui.CheckCommandArgsAndMayPrintHelp(cmd, args, 1)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]

			return cm.Aurora2TiDBCloudRunCluster(clusterName, opt, gOpt)
		},
	}

	return cmd
}
