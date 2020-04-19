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

	"github.com/AlecAivazis/survey/v2"
	"github.com/mgutz/ansi"
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
)

const usernameArg = "username"
const passwordArg = "password"

func runLoginCommand(cmd *cobra.Command, args []string) error {
	username, err := cmd.Flags().GetString(usernameArg)
	if err != nil {
		return err
	}
	password, err := cmd.Flags().GetString(passwordArg)
	if err != nil {
		return err
	}
	fmt.Println("Logging into", opsani.GetBaseURLHostnameAndPort())

	whiteBold := ansi.ColorCode("white+b")
	if username == "" {
		err := survey.AskOne(&survey.Input{
			Message: "Username:",
		}, &username, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%si %sUsername: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, username, ansi.Reset)
	}

	if password == "" {
		err := survey.AskOne(&survey.Password{
			Message: "Password:",
		}, &password, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}
	return nil
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the Opsani API",
	Long:  `Login to the Opsani API and persist access credentials.`,
	Args:  cobra.NoArgs,
	RunE:  runLoginCommand,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP(usernameArg, "u", "", "Opsani Username")
	loginCmd.Flags().StringP(passwordArg, "p", "", "Password")
}
