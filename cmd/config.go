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

package cmd

import (
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
)

// ReduceRunEFuncsO reduces a list of Cobra run functions that return an error into a single aggregate run function
func ReduceRunEFuncsO(runFuncs ...RunEFunc) func(cmd *Command, args []string) error {
	return func(cmd *Command, args []string) error {
		for _, runFunc := range runFuncs {
			if err := runFunc(cmd.Command, args); err != nil {
				return err
			}
		}
		return nil
	}
}

var configCmd = NewCommandWithCobraCommand(&cobra.Command{
	Use:   "config",
	Short: "Manages client configuration",
	Args:  cobra.NoArgs,
}, func(cmd *Command) {
	cmd.RunE = RunConfig
	cmd.PersistentPreRunE = ReduceRunEFuncsO(InitConfigRunE, RequireConfigFileFlagToExistRunE, RequireInitRunE)
})

// RunConfig displays Opsani CLI config info
func RunConfig(cmd *Command, args []string) error {
	cmd.Println("Using config from:", opsani.ConfigFile)
	return cmd.PrettyPrintJSONObject(opsani.GetAllSettings())
}

func init() {
	rootCmd.AddCommand(configCmd.Command)
}
