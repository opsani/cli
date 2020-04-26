// Copyright 2020 Opsani
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type servoImageCommand struct {
	*BaseCommand
}

// NewServoImageCommand returns a new instance of the servo image command
func NewServoImageCommand(baseCmd *BaseCommand) *cobra.Command {
	servoImageCommand := servoImageCommand{BaseCommand: baseCmd}

	servoImageCobra := &cobra.Command{
		Use:   "image",
		Short: "Manage Servo Images",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	servoImageCobra.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Servo images",
		Args:    cobra.NoArgs,
		RunE:    servoImageCommand.RunList,
	})
	servoImageCobra.AddCommand(&cobra.Command{
		Use:   "search",
		Short: "Search for Servo images",
		Args:  cobra.ExactArgs(1),
		RunE:  servoImageCommand.RunSearch,
	})
	servoImageCobra.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Get info about a Servo image",
		Args:  cobra.ExactArgs(1),
		RunE:  servoImageCommand.RunInfo,
	})
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull a Servo image with Docker",
		Args:  cobra.ExactArgs(1),
		RunE:  servoImageCommand.RunPull,
	}
	pullCmd.Flags().StringP(hostArg, "H", "", "Docket host to connect to (overriding DOCKER_HOST)")
	servoImageCobra.AddCommand(pullCmd)

	return servoImageCobra
}

func (cmd *servoImageCommand) RunList(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	data := [][]string{}
	for _, image := range images {
		for _, repoTag := range image.RepoTags {
			if strings.HasPrefix(repoTag, "opsani/") {
				data = append(data, []string{
					repoTag, image.ID,
				})
			}
		}
	}

	table.AppendBulk(data)
	table.Render()

	return nil
}

func (cmd *servoImageCommand) RunSearch(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoImageCommand) RunInfo(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoImageCommand) RunPull(c *cobra.Command, args []string) error {
	dockerHost, err := c.Flags().GetString(hostArg)
	if err != nil {
		return err
	}

	di, err := NewDockerInterface(dockerHost)
	if err != nil {
		return err
	}

	return di.PullImageWithProgressReporting(context.Background(), args[0])
}
