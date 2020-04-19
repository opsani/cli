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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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

/**
Lifecycle commands
*/
var appStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the app",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		resp, err := client.StartApp()
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the app",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		resp, err := client.StopApp()
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the app",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		resp, err := client.RestartApp()
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check app status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		resp, err := client.GetAppStatus()
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

/**
Config commands
*/

func openFileInEditor(filename string, editor string) error {
	components := strings.Split(editor, " ")
	editor, args := components[0], components[1:]
	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	args = append(args, filename)
	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

var appConfigEditCmd = &cobra.Command{
	Use:   "edit [PATH=VALUE ...]",
	Short: "Edit app config",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create temp file
		tempFile, err := ioutil.TempFile(os.TempDir(), "*.json")
		if err != nil {
			return err
		}
		filename := tempFile.Name()

		// Download config to temp
		client := NewAPIClientFromConfig()
		resp, err := client.GetConfig()
		if err != nil {
			return err
		}
		opsani.WritePrettyJSONBytesToFile(resp.Body(), filename)
		if err != nil {
			return err
		}

		// Defer removal of the temporary file in case any of the next steps fail.
		defer os.Remove(filename)

		if err = tempFile.Close(); err != nil {
			return err
		}

		// Apply any inline path edits
		if len(args) > 0 {
			config, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			config, err = SetJSONKeyPathValuesFromStringsOnBytes(args, config)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filename, config, 0755)
			if err != nil {
				return err
			}
		}

		// Edit interactively if necessary
		if len(args) == 0 || appConfig.Interactive {
			if err = openFileInEditor(filename, appConfig.Editor); err != nil {
				return err
			}
		}

		body, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		// Send it back
		resp, err = client.SetConfigFromBody(body, appConfig.ApplyNow)
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appConfigGetCmd = &cobra.Command{
	Use:   "get [PATH ...]",
	Short: "Get app config",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		resp, err := client.GetConfig()
		if err != nil {
			return err
		}

		// Non-filtered invocation
		if len(args) == 0 {
			if appConfig.OutputFile == "" {
				// Print to stdout
				PrettyPrintJSONResponse(resp)
			} else {
				// Write to file
				opsani.WritePrettyJSONBytesToFile(resp.Body(), appConfig.OutputFile)
			}
		} else {
			// Handle filtered invocation
			var jsonStrings []string
			results := gjson.GetManyBytes(resp.Body(), args...)
			for _, result := range results {
				if appConfig.OutputFile == "" {
					PrettyPrintJSONString(result.String())
				} else {
					jsonStrings = append(jsonStrings, result.String())
				}
			}

			// Handle file output
			if appConfig.OutputFile != "" {
				err := opsani.WritePrettyJSONStringsToFile(jsonStrings, appConfig.OutputFile)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

var appConfigSetCmd = &cobra.Command{
	Use:   "set [CONFIG]",
	Short: "Set app config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		var body interface{}
		if filename := appConfig.InputFile; filename != "" {
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			body = bytes
		} else {
			body = args[0]
		}
		resp, err := client.SetConfigFromBody(body, appConfig.ApplyNow)
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appConfigPatchCmd = &cobra.Command{
	Use:   "patch [CONFIG]",
	Short: "Patch app config",
	Long:  "Patch merges the incoming change into the existing configuration.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClientFromConfig()
		var body interface{}
		if filename := appConfig.InputFile; filename != "" {
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			body = bytes
		} else {
			body = args[0]
		}
		resp, err := client.PatchConfigFromBody(body, appConfig.ApplyNow)
		if err != nil {
			return err
		}
		return PrettyPrintJSONResponse(resp)
	},
}

var appConfig = struct {
	OutputFile  string
	InputFile   string
	ApplyNow    bool
	Editor      string
	Interactive bool
}{}

var appConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage app config",

	// Alias for app config get
	Args: appConfigGetCmd.Args,
	RunE: appConfigGetCmd.RunE,
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage apps",
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
	appConfigCmd.AddCommand(appConfigGetCmd)
	appConfigCmd.AddCommand(appConfigSetCmd)
	appConfigCmd.AddCommand(appConfigPatchCmd)
	appConfigCmd.AddCommand(appConfigEditCmd)

	// app config flags
	appConfigCmd.Flags().StringVarP(&appConfig.OutputFile, "output", "o", "", "Write output to file instead of stdout")

	// app config set & patch flags
	appConfigPatchCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing config to apply")
	appConfigPatchCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")
	appConfigSetCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing config to apply")
	appConfigSetCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")

	// app edit flags
	appConfigEditCmd.Flags().StringVarP(&appConfig.Editor, "editor", "e", os.Getenv("EDITOR"), "Edit the config with the given editor (overrides $EDITOR)")
	appConfigEditCmd.Flags().BoolVarP(&appConfig.Interactive, "interactive", "i", false, "Edit the config changes interactively")
}
