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

// NewLoginCommand returns a new `opani login` command instance
func NewLoginCommand() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the Opsani API",
		Long:  `Login to the Opsani API and persist access credentials.`,
		Args:  cobra.NoArgs,
		RunE:  runLoginCommand,
	}

	loginCmd.Flags().StringP(usernameArg, "u", "", "Opsani Username")
	loginCmd.Flags().StringP(passwordArg, "p", "", "Password")

	return loginCmd
}
