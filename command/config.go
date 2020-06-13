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
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type configCommand struct {
	*BaseCommand
}

// NewConfigCommand returns a new instance of the root command for Opsani CLI
func NewConfigCommand(baseCmd *BaseCommand) *cobra.Command {
	cfgCmd := configCommand{BaseCommand: baseCmd}
	cobraCmd := &cobra.Command{
		Use:               "config",
		Short:             "Manage configuration",
		Annotations:       map[string]string{"other": "true"},
		Args:              cobra.NoArgs,
		RunE:              cfgCmd.Run,
		PersistentPreRunE: ReduceRunEFuncs(baseCmd.InitConfigRunE, baseCmd.RequireConfigFileFlagToExistRunE, baseCmd.RequireInitRunE),
	}

	cobraEditCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return openFileInEditor(baseCmd.viperCfg.ConfigFileUsed(), os.Getenv("EDITOR"))
		},
	}
	cobraCmd.AddCommand(cobraEditCmd)

	return cobraCmd
}

// RunConfig displays Opsani CLI config info
func (configCmd *configCommand) Run(_ *cobra.Command, args []string) error {
	configCmd.Println("Using config from:", configCmd.viperCfg.ConfigFileUsed())

	yaml, err := yaml.Marshal(configCmd.GetAllSettings())
	if err != nil {
		return err
	}

	return configCmd.PrettyPrintYAML(yaml, false)
}
