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
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Survey method wrappers
// Survey needs access to a file descriptor for configuring the terminal but Cobra wants to model
// stdio as streams.
var globalStdio terminal.Stdio

// SetStdio is global package helper for testing where access to a file
// descriptor for the TTY is required
func SetStdio(stdio terminal.Stdio) {
	globalStdio = stdio
}

// BaseCommand is the foundational command structure for the Opsani CLI
// It contains the root command for Cobra and is designed for embedding
// into other command structures to add subcommand functionality
type BaseCommand struct {
	rootCobraCommand *cobra.Command
	viperCfg         *viper.Viper

	configFile            string
	profileName           string
	profile               *Profile
	requestTracingEnabled bool
	debugModeEnabled      bool
	disableColors         bool
}

// stdio is a test helper for returning terminal file descriptors usable by Survey
func (cmd *BaseCommand) stdio() terminal.Stdio {
	if globalStdio != (terminal.Stdio{}) {
		return globalStdio
	} else {
		return terminal.Stdio{
			In:  os.Stdin,
			Out: os.Stdout,
			Err: os.Stderr,
		}
	}
}

// RootCobraCommand returns the root Cobra command of the Opsani CLI command
func (cmd *BaseCommand) RootCobraCommand() *cobra.Command {
	return cmd.rootCobraCommand
}

// Viper returns the Viper configuration object underlying the Opsani CLI command
func (cmd *BaseCommand) Viper() *viper.Viper {
	return cmd.viperCfg
}

// Proxy the Cobra I/O methods for convenience

// OutOrStdout returns output to stdout.
func (cmd *BaseCommand) OutOrStdout() io.Writer {
	return cmd.rootCobraCommand.OutOrStdout()
}

// Print is a convenience method to Print to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Print(i ...interface{}) {
	cmd.rootCobraCommand.Print(i...)
}

// Println is a convenience method to Println to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Println(i ...interface{}) {
	cmd.rootCobraCommand.Println(i...)
}

// Printf is a convenience method to Printf to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Printf(format string, i ...interface{}) {
	cmd.rootCobraCommand.Printf(format, i...)
}

// PrintErr is a convenience method to Print to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErr(i ...interface{}) {
	cmd.rootCobraCommand.PrintErr(i...)
}

// PrintErrln is a convenience method to Println to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErrln(i ...interface{}) {
	cmd.rootCobraCommand.PrintErrln(i...)
}

// PrintErrf is a convenience method to Printf to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErrf(format string, i ...interface{}) {
	cmd.rootCobraCommand.PrintErrf(format, i...)
}

// Proxy the Survey library to follow our output directives

// Ask is a wrapper for survey.AskOne that executes with the command's stdio
func (cmd *BaseCommand) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	stdio := cmd.stdio()
	return survey.Ask(qs, response, append(opts, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))...)
}

// AskOne is a wrapper for survey.AskOne that executes with the command's stdio
func (cmd *BaseCommand) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	stdio := cmd.stdio()
	return survey.AskOne(p, response, append(opts, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))...)
}

// PrettyPrintJSONObject prints the given object as pretty printed JSON
func (cmd *BaseCommand) PrettyPrintJSONObject(obj interface{}) error {
	s, err := prettyjson.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(s))
	return err
}

// PrettyPrintJSONBytes prints the given byte array as pretty printed JSON
func (cmd *BaseCommand) PrettyPrintJSONBytes(bytes []byte) error {
	s, err := prettyjson.Format(bytes)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(s))
	return err
}

// PrettyPrintJSONString prints the given string as pretty printed JSON
func (cmd *BaseCommand) PrettyPrintJSONString(str string) error {
	return PrettyPrintJSONBytes([]byte(str))
}

// PrettyPrintJSONResponse prints the given API response as pretty printed JSON
func (cmd *BaseCommand) PrettyPrintJSONResponse(resp *resty.Response) error {
	if resp.IsSuccess() {
		if r := resp.Result(); r != nil {
			return PrettyPrintJSONObject(r)
		}
	} else if resp.IsError() {
		if e := resp.Error(); e != nil {
			return PrettyPrintJSONObject(e)
		}
	}
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return err
	}
	return PrettyPrintJSONObject(result)
}

const escape = "\x1b"

func format(attr color.Attribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

// PrettyPrintYAMLObject pretty prints the given object marshalled into YAML
func (cmd *BaseCommand) PrettyPrintYAMLObject(obj interface{}) error {
	yaml, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	return cmd.PrettyPrintYAML(yaml, false)
}

// PrettyPrintYAML pretty prints the given YAML byte array, optionally including line numbers
func (cmd *BaseCommand) PrettyPrintYAML(bytes []byte, lineNumbers bool) error {
	tokens := lexer.Tokenize(string(bytes))
	var p printer.Printer
	p.LineNumber = lineNumbers
	if cmd.ColorOutput() {
		p.LineNumberFormat = func(num int) string {
			fn := color.New(color.Bold, color.FgHiWhite).SprintFunc()
			return fn(fmt.Sprintf("%2d | ", num))
		}
		p.Bool = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiMagenta),
				Suffix: format(color.Reset),
			}
		}
		p.Number = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiMagenta),
				Suffix: format(color.Reset),
			}
		}
		p.MapKey = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiCyan),
				Suffix: format(color.Reset),
			}
		}
		p.Anchor = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiYellow),
				Suffix: format(color.Reset),
			}
		}
		p.Alias = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiYellow),
				Suffix: format(color.Reset),
			}
		}
		p.String = func() *printer.Property {
			return &printer.Property{
				Prefix: format(color.FgHiGreen),
				Suffix: format(color.Reset),
			}
		}
	}

	// writer := colorable.NewColorableStdout()
	cmd.OutOrStdout().Write([]byte(p.PrintTokens(tokens) + "\n"))
	return nil
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
func (cmd *BaseCommand) PersistentFlags() *pflag.FlagSet {
	return cmd.rootCobraCommand.PersistentFlags()
}

// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func (cmd *BaseCommand) Flags() *pflag.FlagSet {
	return cmd.rootCobraCommand.Flags()
}

// BaseURL returns the Opsani API base URL
// The BaseURL is determined by config (args/env), active profile, or default
func (cmd *BaseCommand) BaseURL() string {
	if baseURL := cmd.valueFromFlagOrEnv(KeyBaseURL, "OPSANI_BASE_URL"); baseURL != "" {
		return baseURL
	}
	if cmd.profile != nil {
		// return cmd.profile.BaseURL
	}
	return DefaultBaseURL
}

func (cmd *BaseCommand) valueFromFlagOrEnv(flagKey string, envKey string) string {
	if value, _ := cmd.PersistentFlags().GetString(flagKey); value != "" {
		return value
	}
	if value, set := os.LookupEnv(envKey); set {
		return value
	}
	return ""
}

// BaseURLHostnameAndPort returns the hostname and port portion of Opsani base URL for summary display
func (cmd *BaseCommand) BaseURLHostnameAndPort() string {
	u, err := url.Parse(cmd.BaseURL())
	if err != nil {
		return cmd.GetBaseURL()
	}
	baseURLDescription := u.Hostname()
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		baseURLDescription = baseURLDescription + ":" + port
	}
	return baseURLDescription
}

func (cmd *BaseCommand) baseURLFromFlagsOrEnv() string {
	return cmd.valueFromFlagOrEnv(KeyBaseURL, "OPSANI_BASE_URL")
}

func (cmd *BaseCommand) appFromFlagsOrEnv() string {
	return cmd.valueFromFlagOrEnv(KeyApp, "OPSANI_APP")
}

func (cmd *BaseCommand) tokenFromFlagsOrEnv() string {
	return cmd.valueFromFlagOrEnv(KeyToken, "OPSANI_TOKEN")
}

// LoadProfile loads the configuration for the specified profile
func (cmd *BaseCommand) LoadProfile() (*Profile, error) {
	registry := NewProfileRegistry(cmd.viperCfg)
	profiles, err := registry.Profiles()
	if err != nil || len(profiles) == 0 {
		return nil, nil
	}

	var profile *Profile
	if cmd.profileName == "" {
		profile = &profiles[0]
	} else {
		profile = registry.ProfileNamed(cmd.profileName)
		if profile == nil {
			return nil, fmt.Errorf("no profile %q", cmd.profileName)
		}
	}

	// Apply any config overrides
	if profile != nil {
		if baseURL := cmd.baseURLFromFlagsOrEnv(); baseURL != "" {
			profile.BaseURL = baseURL
		}
		if app := cmd.appFromFlagsOrEnv(); app != "" {
			profile.App = app
		}
		if token := cmd.tokenFromFlagsOrEnv(); token != "" {
			profile.Token = token
		}

		cmd.profile = profile
		registry.Set(profiles)
	}

	return profile, nil
}

// AccessToken returns the Opsani API access token
func (cmd *BaseCommand) AccessToken() string {
	if token := cmd.valueFromFlagOrEnv(KeyToken, "OPSANI_TOKEN"); token != "" {
		return token
	}
	if cmd.profile != nil {
		return cmd.profile.Token
	}
	return ""
}

// App returns the target Opsani app
func (cmd *BaseCommand) App() string {
	if app := cmd.valueFromFlagOrEnv(KeyApp, "OPSANI_APP"); app != "" {
		return app
	}
	if cmd.profile != nil {
		return cmd.profile.App
	}
	return ""
}

// AppComponents returns the organization name and app ID as separate path components
func (cmd *BaseCommand) AppComponents() (orgSlug string, appSlug string) {
	app := cmd.App()
	org := filepath.Dir(app)
	appID := filepath.Base(app)
	return org, appID
}

// AllSettings returns all configuration settings
func (cmd *BaseCommand) AllSettings() map[string]interface{} {
	return cmd.viperCfg.AllSettings()
}

// DebugModeEnabled returns a boolean value indicating if debugging is enabled
func (cmd *BaseCommand) DebugModeEnabled() bool {
	return cmd.debugModeEnabled
}

// RequestTracingEnabled returns a boolean value indicating if request tracing is enabled
func (cmd *BaseCommand) RequestTracingEnabled() bool {
	return cmd.requestTracingEnabled
}

// ColorOutput indicates if ANSI colors will be used for output
func (cmd *BaseCommand) ColorOutput() bool {
	return !cmd.disableColors
}

// SetColorOutput sets whether or not ANSI colors will be used for output
func (cmd *BaseCommand) SetColorOutput(colorOutput bool) {
	cmd.disableColors = !colorOutput
}
