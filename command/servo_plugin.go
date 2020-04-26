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
		Short: "Manage Servo Plugins",
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
