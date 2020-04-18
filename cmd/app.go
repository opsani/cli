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
	"io/ioutil"
	"os"

	"github.com/hokaccha/go-prettyjson"
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
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

	// Set the output directory to pwd by default
	dir, err := os.Getwd()
	if err == nil {
		c.SetOutputDirectory(dir)
	}
	return c
}

// PrettyPrintJSON prints the given object as pretty printed JSON
func PrettyPrintJSON(obj interface{}) {
	s, _ := prettyjson.Marshal(obj)
	fmt.Println(string(s))
}

/**
Lifecycle commands
*/
var appStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the app",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		status, err := client.StartApp()
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

var appStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the app",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		status, err := client.StopApp()
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

var appRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the app",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		status, err := client.RestartApp()
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

var appStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check app status",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		status, err := client.GetAppStatus()
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

/**
Config commands
*/

var appConfigEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit app configuration interactively via $EDITOR",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edit called")
		// TODO: edit the config
	},
}

var appConfigSetCmd = &cobra.Command{
	Use:   "set [CONFIG]",
	Short: "Set application configuration",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		var body interface{}
		if filename := appConfig.InputFile; filename != "" {
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			body = bytes
		} else {
			// Read literal from the positional argument
			// TODO: support JSON Path/literal format
			body = args[0]
		}
		status, err := client.SetConfigFromBody(body, appConfig.ApplyNow)
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

var appConfigPatchCmd = &cobra.Command{
	Use:   "patch [CONFIG]",
	Short: "Patch the existing application configuration",
	Long:  "Patch merges the incoming change into the existing configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		var body interface{}
		if filename := appConfig.InputFile; filename != "" {
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			body = bytes
		} else {
			// Read literal from the positional argument
			// TODO: support JSON Path/literal format
			body = args[0]
		}
		status, err := client.PatchConfigFromBody(body, appConfig.ApplyNow)
		if err != nil {
			panic(err)
		}
		PrettyPrintJSON(status)
	},
}

var appConfig = struct {
	OutputFile string
	InputFile  string
	ApplyNow   bool
}{}

var appConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewAPIClientFromConfig()
		if appConfig.OutputFile == "" {
			config, err := client.GetConfig()
			if err != nil {
				panic(err)
			}
			PrettyPrintJSON(config)
		} else {
			err := client.GetConfigToOutput(appConfig.OutputFile)
			if err == nil {
				fmt.Printf("Output written to \"%s\"\n", appConfig.OutputFile)
			} else {
				panic(err)
			}
		}
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
	rootCmd.AddCommand(appCmd)

	// Lifecycle
	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appRestartCmd)
	appCmd.AddCommand(appStatusCmd)

	// Config
	appCmd.AddCommand(appConfigCmd)
	appConfigCmd.AddCommand(appConfigEditCmd)
	appConfigCmd.AddCommand(appConfigSetCmd)
	appConfigCmd.AddCommand(appConfigPatchCmd)

	// app config flags
	appConfigCmd.Flags().StringVarP(&appConfig.OutputFile, "output", "o", "", "Write output to file instead of stdout")

	// app config set flags
	// TODO: Read from standard input
	//
	appConfigPatchCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing configuration to apply")
	appConfigPatchCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")
	appConfigSetCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing configuration to apply")
	appConfigSetCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")
}
