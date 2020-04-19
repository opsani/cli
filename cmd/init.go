/*
Copyright © 2020 Blake Watters <blake@opsani.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Opsani config",
	Long: `Initializes an Opsani config file and acquires the required settings:

  * 'app':   Opsani app to control (OPSANI_APP).
  * 'token': API token to authenticate with (OPSANI_TOKEN).
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		confirmed, err := cmd.Flags().GetBool("confirmed")
		if err != nil {
			return err
		}

		// Handle reinitialization case
		overwrite := false
		if _, err := os.Stat(opsani.ConfigFile); !os.IsNotExist(err) && !confirmed {
			fmt.Println("Using config from:", opsani.ConfigFile)
			PrettyPrintJSONObject(opsani.GetAllSettings())

			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Existing config found. Overwrite %s?", opsani.ConfigFile),
			}
			err := survey.AskOne(prompt, &overwrite)
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
			err := survey.AskOne(&survey.Input{
				Message: "Opsani app (i.e. domain.com/app):",
				Default: opsani.GetApp(),
			}, &app, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("%si %sApp: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, app, ansi.Reset)
		}

		if overwrite || token == "" {
			err := survey.AskOne(&survey.Input{
				Message: "API Token:",
				Default: opsani.GetAccessToken(),
			}, &token, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("%si %sAPI Token: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, token, ansi.Reset)
		}

		// Confirm that the user wants to write this config
		opsani.SetApp(app)
		opsani.SetAccessToken(token)

		fmt.Printf("\nOpsani config initialized:\n")
		PrettyPrintJSONObject(opsani.GetAllSettings())
		if !confirmed {
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Write to %s?", opsani.ConfigFile),
			}
			survey.AskOne(prompt, &confirmed)
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
			fmt.Println("\nOpsani CLI initialized")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("confirmed", false, "Write config without asking for confirmation")
}
