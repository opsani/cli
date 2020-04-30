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
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

const confirmedArg = "confirmed"

type initCommand struct {
	*BaseCommand

	confirmed bool
}

// RunInitCommand initializes Opsani CLI config
func (initCmd *initCommand) RunInitCommand(_ *cobra.Command, args []string) error {
	// Handle reinitialization case
	overwrite := false
	if _, err := os.Stat(initCmd.ConfigFile); !os.IsNotExist(err) && !initCmd.confirmed {
		initCmd.Println("Using config from:", initCmd.ConfigFile)
		initCmd.PrettyPrintJSONObject(initCmd.GetAllSettings())

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Existing config found. Overwrite %s?", initCmd.ConfigFile),
		}
		err := initCmd.AskOne(prompt, &overwrite)
		if err != nil {
			return err
		}
		if !overwrite {
			return terminal.InterruptErr
		}
	}
	app := initCmd.App()
	token := initCmd.AccessToken()
	whiteBold := ansi.ColorCode("white+b")

	if overwrite || app == "" {
		err := initCmd.AskOne(&survey.Input{
			Message: "Opsani app (i.e. domain.com/app):",
			Default: app,
		}, &app, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		initCmd.Printf("%si %sApp: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, app, ansi.Reset)
	}

	if overwrite || token == "" {
		err := initCmd.AskOne(&survey.Input{
			Message: "API Token:",
			Default: token,
		}, &token, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		initCmd.Printf("%si %sAPI Token: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, token, ansi.Reset)
	}

	// Confirm that the user wants to write this config
	initCmd.SetApp(app)
	initCmd.SetAccessToken(token)

	initCmd.Printf("\nOpsani config initialized:\n")
	initCmd.PrettyPrintJSONObject(initCmd.GetAllSettings())
	if !initCmd.confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Write to %s?", initCmd.ConfigFile),
		}
		initCmd.AskOne(prompt, &initCmd.confirmed)
	}
	if initCmd.confirmed {
		configDir := filepath.Dir(initCmd.ConfigFile)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.Mkdir(configDir, 0755)
			if err != nil {
				return err
			}
		}
		if err := initCmd.viperCfg.WriteConfigAs(initCmd.ConfigFile); err != nil {
			return err
		}
		initCmd.Println("\nOpsani CLI initialized")
	}
	return nil
}

// NewInitCommand returns a new `opsani init` command instance
func NewInitCommand(baseCommand *BaseCommand) *cobra.Command {
	initCmd := &initCommand{BaseCommand: baseCommand}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Opsani config",
		Long: `Initializes an Opsani config file and acquires the required settings:
	
	  * 'app':   Opsani app to control (OPSANI_APP).
	  * 'token': API token to authenticate with (OPSANI_TOKEN).
	`,
		Args: cobra.NoArgs,
		RunE: initCmd.RunInitCommand,
		PersistentPreRunE: initCmd.InitConfigRunE, // Skip loading the config file
	}
	cmd.Flags().BoolVar(&initCmd.confirmed, confirmedArg, false, "Write config without asking for confirmation")
	return cmd
}
