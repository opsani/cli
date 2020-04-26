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

type configCommand struct {
	*BaseCommand
}

// NewConfigCommand returns a new instance of the root command for Opsani CLI
func NewConfigCommand(baseCmd *BaseCommand) *cobra.Command {
	cfgCmd := configCommand{BaseCommand: baseCmd}
	return &cobra.Command{
		Use:               "config",
		Short:             "Manages client configuration",
		Args:              cobra.NoArgs,
		RunE:              cfgCmd.Run,
		PersistentPreRunE: ReduceRunEFuncs(baseCmd.InitConfigRunE, baseCmd.RequireConfigFileFlagToExistRunE, baseCmd.RequireInitRunE),
	}
}

// RunConfig displays Opsani CLI config info
func (configCmd *configCommand) Run(_ *cobra.Command, args []string) error {
	configCmd.Println("Using config from:", configCmd.ConfigFile)
	return configCmd.PrettyPrintJSONObject(configCmd.GetAllSettings())
}
