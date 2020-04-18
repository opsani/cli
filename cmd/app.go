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

	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"gopkg.in/ffmt.v1"
)

// NewAPIClientFromConfig returns an Opsani API client configured using the active configuration
func NewAPIClientFromConfig() *opsani.Client {
	c := opsani.NewClient().
		SetBaseURL(opsani.GetBaseURL()).
		SetApp(opsani.GetApp()).
		SetAuthToken(opsani.GetAccessToken()).
		SetDebug(opsani.GetDebugModeEnabled())
	tracingEnabled := opsani.GetRequestTracingEnabled()
	if tracingEnabled {
		c.EnableTrace()
	}
	return c
}

var appConfigEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit app configuration interactively via $EDITOR",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edit called")
		// TODO: edit the config
	},
}

var appConfigSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set application configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("set called")
		// TODO: set the config
	},
}

var appConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		config, err := client.GetConfig()
		if err != nil {
			panic(err)
		}
		ffmt.Pjson(config)
	},
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage apps",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("app called")
	},
}

func init() {
	configCmd.AddCommand(appConfigEditCmd)
	configCmd.AddCommand(appConfigSetCmd)
	appCmd.AddCommand(appConfigCmd)
	rootCmd.AddCommand(appCmd)

	appConfigSetCmd.Flags().StringP("file", "f", "", "File containing configuration to apply")
}
