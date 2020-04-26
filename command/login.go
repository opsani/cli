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
	"github.com/spf13/cobra"
)

type loginCommand struct {
	*BaseCommand

	username string
	password string
}

func (loginCmd *loginCommand) runLoginCommand(_ *cobra.Command, args []string) error {
	fmt.Println("Logging into", loginCmd.GetBaseURLHostnameAndPort())

	whiteBold := ansi.ColorCode("white+b")
	if loginCmd.username == "" {
		err := survey.AskOne(&survey.Input{
			Message: "Username:",
		}, &loginCmd.username, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%si %sUsername: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, loginCmd.username, ansi.Reset)
	}

	if loginCmd.password == "" {
		err := survey.AskOne(&survey.Password{
			Message: "Password:",
		}, &loginCmd.password, survey.WithValidator(survey.Required))
		if err != nil {
			return err
		}
	}
	return nil
}

// NewLoginCommand returns a new `opani login` command instance
func NewLoginCommand(baseCmd *BaseCommand) *cobra.Command {
	loginCmd := loginCommand{BaseCommand: baseCmd}
	c := &cobra.Command{
		Use:   "login",
		Short: "Login to the Opsani API",
		Long:  `Login to the Opsani API and persist access credentials.`,
		Args:  cobra.NoArgs,
		RunE:  loginCmd.runLoginCommand,
	}

	c.Flags().StringVarP(&loginCmd.username, "username", "u", "", "Opsani Username")
	c.Flags().StringVarP(&loginCmd.password, "password", "p", "", "Password")

	return c
}
