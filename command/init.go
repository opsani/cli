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
	"github.com/go-resty/resty/v2"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

const confirmedArg = "confirmed"

type initCommand struct {
	*BaseCommand

	confirmed bool
}

func (initCmd *initCommand) RunInitWithTokenCommand(_ *cobra.Command, args []string) error {
	initToken := args[0]

	initCmd.Printf("Initializing with token: %s...\n", initToken)

	configFile := initCmd.viperCfg.ConfigFileUsed()
	// NOTE: On first launch with no config file, Viper returns ""
	// because no config file was resolved. There's probably a cleaner solution...
	if configFile == "" {
		configFile = initCmd.DefaultConfigFile()
	}
	var profile Profile
	URL := fmt.Sprintf("http://localhost:5678/init/%s", initToken)
	client := resty.New()
	resp, err := client.R().
		SetResult(&profile).
		Get(URL)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("Failed initialization with token %q (%s)", initToken, resp.Body())
	}

	// Confirm that the user wants to write this config
	registry := NewProfileRegistry(initCmd.viperCfg)
	registry.AddProfile(profile)

	initCmd.Printf("\nOpsani config initialized:\n")
	initCmd.PrettyPrintYAMLObject(initCmd.GetAllSettings())
	if !initCmd.confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Write to %s?", configFile),
		}
		initCmd.AskOne(prompt, &initCmd.confirmed)
	}
	if initCmd.confirmed {
		configDir := filepath.Dir(configFile)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.Mkdir(configDir, 0755)
			if err != nil {
				return err
			}
		}
		if err := initCmd.viperCfg.WriteConfigAs(configFile); err != nil {
			return err
		}
		initCmd.Println("\nOpsani CLI initialized")
	}
	initCmd.Println("\nBegin optimizing by working with an interactive demo via `opsani ignite`")
	initCmd.Println("Or jump right in to connecting your app by running `opsani vital`")
	return nil
}

// RunInitCommand initializes Opsani CLI config
func (initCmd *initCommand) RunInitCommand(_ *cobra.Command, args []string) error {
	// Handle reinitialization case
	overwrite := false
	configFile := initCmd.viperCfg.ConfigFileUsed()
	// NOTE: On first launch with no config file, Viper returns ""
	// because no config file was resolved. There's probably a cleaner solution...
	if configFile == "" {
		configFile = initCmd.DefaultConfigFile()
	}
	if _, err := os.Stat(configFile); !os.IsNotExist(err) && !initCmd.confirmed {
		initCmd.Println("Using config from:", configFile)
		initCmd.PrettyPrintYAMLObject(initCmd.GetAllSettings())

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Existing config found. Overwrite %s?", configFile),
		}
		err := initCmd.AskOne(prompt, &overwrite)
		if err != nil {
			return err
		}
		if !overwrite {
			return terminal.InterruptErr
		}
	}

	profile := Profile{
		Name:    "default",
		App:     initCmd.App(),
		Token:   initCmd.AccessToken(),
		BaseURL: initCmd.BaseURL(),
	}
	whiteBold := ansi.ColorCode("white+b")

	if overwrite || profile.App == "" {
		err := initCmd.AskOne(&survey.Input{
			Message: "Opsani app (i.e. domain.com/app):",
			Default: profile.App,
		}, &profile.App, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		initCmd.Printf("%si %sApp: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, profile.App, ansi.Reset)
	}

	if overwrite || profile.Token == "" {
		err := initCmd.AskOne(&survey.Input{
			Message: "API Token:",
			Default: profile.Token,
		}, &profile.Token, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		initCmd.Printf("%si %sAPI Token: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, profile.Token, ansi.Reset)
	}

	// Confirm that the user wants to write this config
	registry := NewProfileRegistry(initCmd.viperCfg)
	registry.AddProfile(profile)

	initCmd.Printf("\nOpsani config initialized:\n")
	initCmd.PrettyPrintYAMLObject(initCmd.GetAllSettings())
	if !initCmd.confirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Write to %s?", configFile),
		}
		initCmd.AskOne(prompt, &initCmd.confirmed)
	}
	if initCmd.confirmed {
		configDir := filepath.Dir(configFile)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.Mkdir(configDir, 0755)
			if err != nil {
				return err
			}
		}
		if err := initCmd.viperCfg.WriteConfigAs(configFile); err != nil {
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
		Use:   "init [INIT_TOKEN]",
		Short: "Initialize Opsani config",
		Long: `Initializes an Opsani config file and acquires the required settings:
	
    * 'app':   Opsani app to control (OPSANI_APP).
    * 'token': API token to authenticate with (OPSANI_TOKEN).
	`,
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: initCmd.InitConfigRunE, // Skip loading the config file
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 1 {
				return initCmd.RunInitWithTokenCommand(c, args)
			} else {
				return initCmd.RunInitCommand(c, args)
			}
		},
	}
	cmd.Flags().BoolVar(&initCmd.confirmed, confirmedArg, false, "Write config without asking for confirmation")
	return cmd
}
