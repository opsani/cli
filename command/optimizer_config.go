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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

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

// NewOptimizerConfigEditCommand returns a new Opsani CLI app config edit action
func NewOptimizerConfigEditCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "edit [PATH=VALUE ...]",
		Short: "Edit optimizer config",
		Args:  ValidSetJSONKeyPathArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create temp file
			tempFile, err := ioutil.TempFile(os.TempDir(), "*.json")
			if err != nil {
				return err
			}
			filename := tempFile.Name()

			// Download config to temp
			client := baseCmd.NewAPIClient()
			resp, err := client.GetConfig()
			if err != nil {
				return err
			}
			if err = opsani.WritePrettyJSONBytesToFile(resp.Body(), filename); err != nil {
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

				if err = ioutil.WriteFile(filename, config, 0755); err != nil {
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
}

// NewOptimizerConfigGetCommand returns a new Opsani CLI `app config get` action
func NewOptimizerConfigGetCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "get [PATH ...]",
		Short: "Get optimizer config",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			resp, err := client.GetConfig()
			if err != nil {
				return err
			}

			// Non-filtered invocation
			if len(args) == 0 {
				if appConfig.OutputFile == "" {
					// Print to stdout
					if err = PrettyPrintJSONResponse(resp); err != nil {
						return err
					}
				} else {
					// Write to file
					if err = opsani.WritePrettyJSONBytesToFile(resp.Body(), appConfig.OutputFile); err != nil {
						return err
					}
				}
			} else {
				// Handle filtered invocation
				var jsonStrings []string
				results := gjson.GetManyBytes(resp.Body(), args...)
				for _, result := range results {
					if appConfig.OutputFile == "" {
						if err = PrettyPrintJSONString(result.String()); err != nil {
							return err
						}
					} else {
						jsonStrings = append(jsonStrings, result.String())
					}
				}

				// Handle file output
				if appConfig.OutputFile != "" {
					if err := opsani.WritePrettyJSONStringsToFile(jsonStrings, appConfig.OutputFile); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}
}

func bodyForConfigUpdateWithArgs(args []string) (interface{}, error) {
	if filename := appConfig.InputFile; filename != "" {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(bytes, &map[string]interface{}{})
		if err != nil {
			return nil, fmt.Errorf("file %v is not valid JSON: %w", filename, err)
		}
		return bytes, nil
	} else {
		if len(args) == 0 {
			return nil, fmt.Errorf("cannot patch without a JSON config argument")
		}
		return args[0], nil
	}
}

// NewOptimizerConfigSetCommand returns a new Opsani CLI `app config set` action
func NewOptimizerConfigSetCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "set [CONFIG]",
		Short: "Set optimizer config",
		Args:  RangeOfValidJSONArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			body, err := bodyForConfigUpdateWithArgs(args)
			if err != nil {
				return err
			}

			resp, err := client.SetConfigFromBody(body, appConfig.ApplyNow)
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

// NewOptimizerConfigPatchCommand returns a new Opsani CLI `app config patch` action
func NewOptimizerConfigPatchCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "patch [CONFIG]",
		Short: "Patch optimizer config",
		Long:  "Patch merges the incoming change into the existing configuration.",
		Args:  RangeOfValidJSONArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			body, err := bodyForConfigUpdateWithArgs(args)
			if err != nil {
				return err
			}

			resp, err := client.PatchConfigFromBody(body, appConfig.ApplyNow)
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

var appConfig = struct {
	OutputFile  string
	InputFile   string
	ApplyNow    bool
	Editor      string
	Interactive bool
}{}

// NewOptimizerConfigCommand returns a new Opsani CLI `app config` action
func NewOptimizerConfigCommand(baseCmd *BaseCommand) *cobra.Command {
	appConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage optimizer configuration",
	}

	appConfigGetCmd := NewOptimizerConfigGetCommand(baseCmd)
	appConfigSetCmd := NewOptimizerConfigSetCommand(baseCmd)
	appConfigPatchCmd := NewOptimizerConfigPatchCommand(baseCmd)
	appConfigEditCmd := NewOptimizerConfigEditCommand(baseCmd)

	appConfigCmd.AddCommand(appConfigGetCmd)
	appConfigCmd.AddCommand(appConfigSetCmd)
	appConfigCmd.AddCommand(appConfigPatchCmd)
	appConfigCmd.AddCommand(appConfigEditCmd)

	// alias for app config get
	appConfigCmd.Args = appConfigGetCmd.Args
	appConfigCmd.RunE = appConfigGetCmd.RunE

	// app config flags
	appConfigCmd.Flags().StringVarP(&appConfig.OutputFile, "output", "o", "", "Write output to file instead of stdout")
	appConfigCmd.MarkFlagFilename("output")

	// app config set & patch flags
	updateGlobs := []string{"*.json", "*.yaml", "*.yml"}
	appConfigPatchCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing config to apply")
	appConfigPatchCmd.MarkFlagFilename("file", updateGlobs...)
	appConfigPatchCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")
	appConfigSetCmd.Flags().StringVarP(&appConfig.InputFile, "file", "f", "", "File containing config to apply")
	appConfigSetCmd.MarkFlagFilename("file", updateGlobs...)
	appConfigSetCmd.Flags().BoolVarP(&appConfig.ApplyNow, "apply", "a", true, "Apply the config changes immediately")

	// app edit flags
	appConfigEditCmd.Flags().StringVarP(&appConfig.Editor, "editor", "e", os.Getenv("EDITOR"), "Edit the config with the given editor (overrides $EDITOR)")
	appConfigEditCmd.Flags().BoolVarP(&appConfig.Interactive, "interactive", "i", false, "Edit the config changes interactively")

	return appConfigCmd
}

// ValidSetJSONKeyPathArgs checks that positional arguments are valid key paths for setting values
func ValidSetJSONKeyPathArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	for _, arg := range args {
		if matched, err := regexp.Match("^(.+)=(.+)$", []byte(arg)); err != nil {
			return err
		} else if !matched {
			return fmt.Errorf("argument '%s' is not of the form [PATH]=[VALUE]", arg)
		}
	}
	return nil
}

// RangeOfValidJSONArgs ensures that the number of args are within the range and are all valid JSON
func RangeOfValidJSONArgs(min int, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", min, max, len(args))
		}
		for i, arg := range args {
			if err := json.Unmarshal([]byte(arg), &map[string]interface{}{}); err != nil {
				return fmt.Errorf("argument %v (\"%s\") is not valid JSON: %w", i, arg, err)
			}
		}
		return nil
	}
}
