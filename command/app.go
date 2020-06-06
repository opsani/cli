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
	"log"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// NewAppCommand returns a new `opsani app` command instance
func NewAppCommand(baseCmd *BaseCommand) *cobra.Command {
	appCmd := &cobra.Command{
		Use:     "app",
		Aliases: []string{"optimizer"},
		Short:   "Manage apps",

		// All commands require an initialized client
		PersistentPreRunE: baseCmd.InitConfigRunE,
	}

	// Initialize our subcommands
	appStartCmd := NewAppStartCommand(baseCmd)
	appStopCmd := NewAppStopCommand(baseCmd)
	appRestartCmd := NewAppRestartCommand(baseCmd)
	appStatusCmd := NewAppStatusCommand(baseCmd)
	appConfigCmd := NewAppConfigCommand(baseCmd)

	// Lifecycle
	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appRestartCmd)
	appCmd.AddCommand(appStatusCmd)

	// Config
	appCmd.AddCommand(appConfigCmd)

	appCmd.AddCommand(NewAppConsoleCommand(baseCmd))

	return appCmd
}

// NewAppConsoleCommand returns a command that opens the Opsani Console
// in the default browser
func NewAppConsoleCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "console",
		Short: "Open the Opsani console in the default web browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			org, appID := baseCmd.GetAppComponents()
			url := fmt.Sprintf("https://console.opsani.com/accounts/%s/applications/%s", org, appID)
			openURLInDefaultBrowser(url)
			return nil
		},
	}
}

func openURLInDefaultBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
