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
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NOTE: Binding vars instead of using flags because the call stack is messy atm
type profileCommand struct {
	*BaseCommand
	verbose bool
	force   bool
}

// NewProfileCommand returns a new instance of the profile command
func NewProfileCommand(baseCmd *BaseCommand) *cobra.Command {
	profileCommand := profileCommand{BaseCommand: baseCmd}

	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage app profiles",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	// Profile registry
	listCmd := &cobra.Command{
		Use:         "list",
		Annotations: map[string]string{"registry": "true"},
		Aliases:     []string{"ls"},
		Short:       "List app profiles",
		Args:        cobra.NoArgs,
		RunE:        profileCommand.RunProfileList,
	}
	listCmd.Flags().BoolVarP(&profileCommand.verbose, "verbose", "v", false, "Display verbose output")
	profileCmd.AddCommand(listCmd)
	addCmd := &cobra.Command{
		Use:                   "add [OPTIONS] [NAME]",
		Long:                  "Add an app profile to the local registry",
		Annotations:           map[string]string{"registry": "true"},
		Short:                 "Add a profile",
		Args:                  cobra.MaximumNArgs(1),
		RunE:                  profileCommand.RunAddProfile,
		DisableFlagsInUseLine: true,
	}
	profileCmd.AddCommand(addCmd)

	removeCmd := &cobra.Command{
		Use:                   "remove [OPTIONS] [NAME]",
		Long:                  "Remove an app profile from the local registry",
		Annotations:           map[string]string{"registry": "true"},
		Aliases:               []string{"rm"},
		Short:                 "Remove a Profile",
		Args:                  cobra.ExactArgs(1),
		RunE:                  profileCommand.RunRemoveProfile,
		DisableFlagsInUseLine: true,
	}
	removeCmd.Flags().BoolVarP(&profileCommand.force, "force", "f", false, "Don't prompt for confirmation")
	profileCmd.AddCommand(removeCmd)

	return profileCmd
}

func (profileCmd *profileCommand) RunAddProfile(c *cobra.Command, args []string) error {
	profile := Profile{
		App:     profileCmd.appFromFlagsOrEnv(),
		Token:   profileCmd.tokenFromFlagsOrEnv(),
		BaseURL: profileCmd.BaseURL(),
	}
	if len(args) > 0 {
		profile.Name = args[0]
	}

	if profile.Name == "" {
		err := profileCmd.AskOne(&survey.Input{
			Message: "Profile name?",
		}, &profile.Name, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	if profile.App == "" {
		err := profileCmd.AskOne(&survey.Input{
			Message: "Opsani app (i.e. domain.com/app)?",
		}, &profile.App, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	if profile.Token == "" {
		err := profileCmd.AskOne(&survey.Input{
			Message: "API Token?",
		}, &profile.Token, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}

	registry := NewProfileRegistry(profileCmd.viperCfg)
	registry.AddProfile(profile)
	return registry.Save()
}

func (profileCmd *profileCommand) RunRemoveProfile(_ *cobra.Command, args []string) error {
	registry := NewProfileRegistry(profileCmd.viperCfg)
	name := args[0]
	profile := registry.ProfileNamed(name)
	if profile == nil {
		return fmt.Errorf("Unable to find profile %q", name)
	}

	confirmed := profileCmd.force
	if !confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Remove profile %q?", profile.Name),
		}
		profileCmd.AskOne(prompt, &confirmed)
	}

	if confirmed {
		registry.RemoveProfile(*profile)
		return registry.Save()
	}

	return nil
}

func (profileCmd *profileCommand) RunProfileList(_ *cobra.Command, args []string) error {
	table := tablewriter.NewWriter(profileCmd.OutOrStdout())
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
	registry := NewProfileRegistry(profileCmd.viperCfg)
	profiles, _ := registry.Profiles()

	if profileCmd.verbose {
		headers := []string{"NAME", "APP", "TOKEN"}
		for _, profile := range profiles {
			row := []string{
				profile.Name,
				profile.App,
				profile.Token,
			}
			data = append(data, row)
		}
		table.SetHeader(headers)
	} else {
		for _, profile := range profiles {
			row := []string{
				profile.Name,
				profile.App,
				profile.Token,
			}
			data = append(data, row)
		}
	}

	table.AppendBulk(data)
	table.Render()
	return nil
}
