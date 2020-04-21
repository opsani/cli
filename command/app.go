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

// NewAppCommand returns a new `opsani app` command instance
func NewAppCommand() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage apps",

		// All commands require an initialized client
		PersistentPreRunE: InitConfigRunE,
	}

	// Initialize our subcommands
	appStartCmd := NewAppStartCommand()
	appStopCmd := NewAppStopCommand()
	appRestartCmd := NewAppRestartCommand()
	appStatusCmd := NewAppStatusCommand()
	appConfigCmd := NewAppConfigCommand()

	// Lifecycle
	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appRestartCmd)
	appCmd.AddCommand(appStatusCmd)

	// Config
	appCmd.AddCommand(appConfigCmd)

	return appCmd
}
