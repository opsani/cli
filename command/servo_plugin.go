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
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/github"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type servoPluginCommand struct {
	*BaseCommand
}

// NewServoPluginCommand returns a new instance of the servo image command
func NewServoPluginCommand(baseCmd *BaseCommand) *cobra.Command {
	servoPluginCommand := servoPluginCommand{BaseCommand: baseCmd}

	servoPluginCobra := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Servo plugins",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	servoPluginCobra.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Servo plugins",
		Args:    cobra.NoArgs,
		RunE:    servoPluginCommand.RunList,
	})
	servoPluginCobra.AddCommand(&cobra.Command{
		Use:   "search",
		Short: "Search for Servo plugins",
		Args:  cobra.ExactArgs(1),
		RunE:  servoPluginCommand.RunSearch,
	})
	servoPluginCobra.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Get info about a Servo plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  servoPluginCommand.RunInfo,
	})
	servoPluginCobra.AddCommand(&cobra.Command{
		Use:   "clone",
		Short: "Clone a Servo plugin with Git",
		Args:  cobra.ExactArgs(1),
		RunE:  servoPluginCommand.RunClone,
	})
	servoPluginCobra.AddCommand(&cobra.Command{
		Use:   "fork",
		Short: "Fork a Servo plugin on GitHub",
		Args:  cobra.ExactArgs(1),
		RunE:  servoPluginCommand.RunFork,
	})
	servoPluginCobra.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new Servo plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  servoPluginCommand.RunCreate,
	})

	return servoPluginCobra
}

func (cmd *servoPluginCommand) RunList(_ *cobra.Command, args []string) error {
	client := github.NewClient(nil)

	ctx := context.Background()
	opt := new(github.RepositoryListByOrgOptions)
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, "opsani", opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			// Skip non-servo repos
			if !strings.HasPrefix(*repo.Name, "servo-") {
				continue
			}
			allRepos = append(allRepos, repo)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Build a table outputting all the servo plugins
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
	headers := []string{"NAME", "DESCRIPTION", "UPDATED", "URL"}
	for _, repo := range allRepos {
		row := []string{
			repo.GetName(),
			repo.GetDescription(),
			humanize.Time(repo.GetUpdatedAt().Time),
			fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\\n", repo.GetHTMLURL(), repo.GetHTMLURL()),
		}
		data = append(data, row)
	}
	table.SetHeader(headers)
	table.AppendBulk(data)
	table.Render()

	return nil
}

func (cmd *servoPluginCommand) RunSearch(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoPluginCommand) RunInfo(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoPluginCommand) RunClone(c *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoPluginCommand) RunFork(c *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoPluginCommand) RunCreate(c *cobra.Command, args []string) error {
	return nil
}
