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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/ffmt.v1"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Opsani config",
	Long: `Initializes an Opsani config file and acquires the required settings:

  * 'app':   Opsani app to control (OPSANI_APP).
  * 'token': API token to authenticate with (OPSANI_TOKEN).
`,
	Run: func(cmd *cobra.Command, args []string) {
		whiteBold := ansi.ColorCode("white+b")
		if opsaniConfig.App == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Opsani app (i.e. domain.com/app):",
			}, &opsaniConfig.App, survey.WithValidator(survey.Required))
			if err == terminal.InterruptErr {
				os.Exit(0)
			} else if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("%si %sApp: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, opsaniConfig.App, ansi.Reset)
		}

		if opsaniConfig.Token == "" {
			err := survey.AskOne(&survey.Input{
				Message: "API Token:",
			}, &opsaniConfig.Token, survey.WithValidator(survey.Required))
			if err == terminal.InterruptErr {
				os.Exit(0)
			} else if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("%si %sAPI Token: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, opsaniConfig.Token, ansi.Reset)
		}

		// Confirm that the user wants to write this config
		viper.Set("app", opsaniConfig.App)
		viper.Set("token", opsaniConfig.Token)

		fmt.Printf("\nOpsani config initialized:\n")
		ffmt.Print(viper.AllSettings())
		confirmed := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Write to %s?", opsaniConfig.ConfigFile),
		}
		survey.AskOne(prompt, &confirmed)
		if confirmed {
			configDir := filepath.Dir(opsaniConfig.ConfigFile)
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				err = os.Mkdir(configDir, 0755)
				if err != nil {
					panic(err)
				}
			}
			if err := viper.WriteConfigAs(opsaniConfig.ConfigFile); err != nil {
				panic(err)
			}
			fmt.Println("\nOpsani CLI initialized")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
