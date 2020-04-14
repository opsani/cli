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
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

// Configuration options bound via Cobra
var loginConfig = struct {
	Username string
	Password string
}{}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the Opsani API",
	Long:  `Login to the Opsani API and persist access credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		u, err := url.Parse(baseURL)
		if err != nil {
			panic(err)
		}
		baseURLDescription := u.Hostname()
		if port := u.Port(); port != "" && port != "80" && port != "443" {
			baseURLDescription = baseURLDescription + ":" + port
		}
		fmt.Println("Logging into", baseURLDescription)

		whiteBold := ansi.ColorCode("white+b")
		if loginConfig.Username == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Username:",
			}, &loginConfig.Username, survey.WithValidator(survey.Required))
			if err == terminal.InterruptErr {
				os.Exit(0)
			} else if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("%si %sUsername: %s%s%s%s\n", ansi.Blue, whiteBold, ansi.Reset, ansi.LightCyan, loginConfig.Username, ansi.Reset)
		}

		if loginConfig.Password == "" {
			err := survey.AskOne(&survey.Password{
				Message: "Password:",
			}, &loginConfig.Password, survey.WithValidator(survey.Required))
			if err == terminal.InterruptErr {
				os.Exit(0)
			} else if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&loginConfig.Username, "username", "u", "", "Opsani Username")
	loginCmd.Flags().StringVarP(&loginConfig.Password, "password", "p", "", "Password")
}
