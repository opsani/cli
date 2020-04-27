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

type authCommand struct {
	*BaseCommand
}

// NewAuthCommand returns a new `opani login` command instance
func NewAuthCommand(baseCmd *BaseCommand) *cobra.Command {
	authCommand := authCommand{BaseCommand: baseCmd}

	authCobra := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Args:  cobra.NoArgs,
		PersistentPreRunE: ReduceRunEFuncs(
			baseCmd.InitConfigRunE,
			baseCmd.RequireConfigFileFlagToExistRunE,
			baseCmd.RequireInitRunE,
		),
	}

	loginCobra := &cobra.Command{
		Use:   "login",
		Short: "Login to Opsani",
		Args:  cobra.NoArgs,
		RunE:  authCommand.runLoginCommand,
	}

	loginCobra.Flags().StringP("username", "u", "", "Opsani Username")
	loginCobra.Flags().StringP("password", "p", "", "Password")
	authCobra.AddCommand(loginCobra)

	authCobra.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout from Opsani",
		Args:  cobra.NoArgs,
		RunE:  authCommand.runLogoutCommand,
	})

	return authCobra
}

func (cmd *authCommand) runLoginCommand(c *cobra.Command, args []string) error {
	fmt.Println("Logging into", cmd.GetBaseURLHostnameAndPort())

	whiteBold := ansi.ColorCode("white+b")
	username, _ := c.Flags().GetString("username")
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

	password, _ := c.Flags().GetString("password")
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

func (cmd *authCommand) runLogoutCommand(c *cobra.Command, args []string) error {
	fmt.Println("Logged out from", cmd.GetBaseURLHostnameAndPort())
	return nil
}
