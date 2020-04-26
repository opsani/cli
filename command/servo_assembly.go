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

type servoAssemblyCommand struct {
	*BaseCommand
}

// NewServoAssemblyCommand returns a new instance of the servo image command
func NewServoAssemblyCommand(baseCmd *BaseCommand) *cobra.Command {
	servoAssemblyCommand := servoAssemblyCommand{BaseCommand: baseCmd}

	servoAssemblyCobra := &cobra.Command{
		Use:   "assembly",
		Short: "Manage Servo Assemblies",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Servo assemblies",
		Args:    cobra.NoArgs,
		RunE:    servoAssemblyCommand.RunList,
	})
	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:   "search",
		Short: "Search for Servo assemblies",
		Args:  cobra.ExactArgs(1),
		RunE:  servoAssemblyCommand.RunSearch,
	})
	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Get info about a Servo assembly",
		Args:  cobra.ExactArgs(1),
		RunE:  servoAssemblyCommand.RunInfo,
	})
	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:   "clone",
		Short: "Clone a Servo assembly with Git",
		Args:  cobra.ExactArgs(1),
		RunE:  servoAssemblyCommand.RunClone,
	})
	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:   "fork",
		Short: "Fork a Servo assembly on GitHub",
		Args:  cobra.ExactArgs(1),
		RunE:  servoAssemblyCommand.RunFork,
	})
	servoAssemblyCobra.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new Servo assembly",
		Args:  cobra.ExactArgs(1),
		RunE:  servoAssemblyCommand.RunCreate,
	})

	return servoAssemblyCobra
}

func (cmd *servoAssemblyCommand) RunList(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoAssemblyCommand) RunSearch(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoAssemblyCommand) RunInfo(_ *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoAssemblyCommand) RunClone(c *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoAssemblyCommand) RunFork(c *cobra.Command, args []string) error {
	return nil
}

func (cmd *servoAssemblyCommand) RunCreate(c *cobra.Command, args []string) error {
	return nil
}
