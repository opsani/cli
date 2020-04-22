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
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const confirmedArg = "confirmed"

// RunInitCommand initializes Opsani CLI config
func RunInitCommand(cmd *Command, args []string) error {
	confirmed, err := cmd.Flags().GetBool(confirmedArg)
	if err != nil {
		return err
	}

	// Handle reinitialization case
	overwrite := false
	if _, err := os.Stat(opsani.ConfigFile); !os.IsNotExist(err) && !confirmed {
		cmd.Println("Using config from:", opsani.ConfigFile)
		cmd.PrettyPrintJSONObject(opsani.GetAllSettings())

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Existing config found. Overwrite %s?", opsani.ConfigFile),
		}
		err := cmd.AskOne(prompt, &overwrite)
		if err != nil {
			return err
		}
		if !overwrite {
			return terminal.InterruptErr
		}
	}
	app := opsani.GetApp()
	token := opsani.GetAccessToken()
	whiteBold := ansi.ColorCode("white+b")

	if overwrite || app == "" {
		err := cmd.AskOne(&survey.Input{
			Message: "Opsani app (i.e. domain.com/app):",
			Default: opsani.GetApp(),
		}, &app, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		cmd.Printf("%si %sApp: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, app, ansi.Reset)
	}

	if overwrite || token == "" {
		err := cmd.AskOne(&survey.Input{
			Message: "API Token:",
			Default: opsani.GetAccessToken(),
		}, &token, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		cmd.Printf("%si %sAPI Token: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, token, ansi.Reset)
	}

	// Confirm that the user wants to write this config
	opsani.SetApp(app)
	opsani.SetAccessToken(token)

	cmd.Printf("\nOpsani config initialized:\n")
	cmd.PrettyPrintJSONObject(opsani.GetAllSettings())
	if !confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Write to %s?", opsani.ConfigFile),
		}
		cmd.AskOne(prompt, &confirmed)
	}
	if confirmed {
		configDir := filepath.Dir(opsani.ConfigFile)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.Mkdir(configDir, 0755)
			if err != nil {
				return err
			}
		}
		if err := viper.WriteConfigAs(opsani.ConfigFile); err != nil {
			return err
		}
		cmd.Println("\nOpsani CLI initialized")
	}
	return nil
}

// NewInitCommand returns a new `opsani init` command instance
func NewInitCommand() *Command {
	return NewCommandWithCobraCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize Opsani config",
		Long: `Initializes an Opsani config file and acquires the required settings:
	
	  * 'app':   Opsani app to control (OPSANI_APP).
	  * 'token': API token to authenticate with (OPSANI_TOKEN).
	`,
		Args: cobra.NoArgs,
	}, func(cmd *Command) {
		cmd.RunE = RunInitCommand
		cmd.Flags().Bool(confirmedArg, false, "Write config without asking for confirmation")
	})
}
