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
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	rootCmd  *cobra.Command
	viperCfg *viper.Viper

	ConfigFile            string
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

// CobraCommand returns the Cobra instance underlying the Opsani CLI command
func (cmd *BaseCommand) CobraCommand() *cobra.Command {
	return cmd.rootCmd
}

// Viper returns the Viper configuration object underlying the Opsani CLI command
// func (cmd *BaseCommand) Viper() *viper {
// 	return viper
// }

// Proxy the Cobra I/O methods for convenience

// OutOrStdout returns output to stdout.
func (cmd *BaseCommand) OutOrStdout() io.Writer {
	return cmd.rootCmd.OutOrStdout()
}

// Print is a convenience method to Print to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Print(i ...interface{}) {
	cmd.rootCmd.Print(i...)
}

// Println is a convenience method to Println to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Println(i ...interface{}) {
	cmd.rootCmd.Println(i...)
}

// Printf is a convenience method to Printf to the defined output, fallback to Stderr if not set.
func (cmd *BaseCommand) Printf(format string, i ...interface{}) {
	cmd.rootCmd.Printf(format, i...)
}

// PrintErr is a convenience method to Print to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErr(i ...interface{}) {
	cmd.rootCmd.PrintErr(i...)
}

// PrintErrln is a convenience method to Println to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErrln(i ...interface{}) {
	cmd.rootCmd.PrintErrln(i...)
}

// PrintErrf is a convenience method to Printf to the defined Err output, fallback to Stderr if not set.
func (cmd *BaseCommand) PrintErrf(format string, i ...interface{}) {
	cmd.rootCmd.PrintErrf(format, i...)
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

func (cmd *BaseCommand) prettyPrintYAML(bytes []byte, lineNumbers bool) error {
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

// BaseURL returns the Opsani API base URL
func (cmd *BaseCommand) BaseURL() string {
	return cmd.viperCfg.GetString(KeyBaseURL)
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

// SetBaseURL sets the Opsani API base URL
func (cmd *BaseCommand) SetBaseURL(baseURL string) {
	cmd.viperCfg.Set(KeyBaseURL, baseURL)
}

// AccessToken returns the Opsani API access token
func (cmd *BaseCommand) AccessToken() string {
	return cmd.viperCfg.GetString(KeyToken)
}

// SetAccessToken sets the Opsani API access token
func (cmd *BaseCommand) SetAccessToken(accessToken string) {
	cmd.viperCfg.Set(KeyToken, accessToken)
}

// App returns the target Opsani app
func (cmd *BaseCommand) App() string {
	return cmd.viperCfg.GetString(KeyApp)
}

// SetApp sets the target Opsani app
func (cmd *BaseCommand) SetApp(app string) {
	cmd.viperCfg.Set(KeyApp, app)
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

// Servos returns the Servos in the configuration
func (cmd *BaseCommand) Servos() ([]Servo, error) {
	servos := make([]Servo, 0)
	err := cmd.viperCfg.UnmarshalKey("servos", &servos)
	if err != nil {
		return nil, err
	}
	return servos, nil
}

// lookupServo named returns the Servo with the given name and its index in the config
func (cmd *BaseCommand) lookupServo(name string) (*Servo, int) {
	var servo *Servo
	servos, err := cmd.Servos()
	if err != nil {
		return nil, 0
	}
	var index int
	for i, s := range servos {
		if s.Name == name {
			servo = &s
			index = i
			break
		}
	}

	return servo, index
}

// ServoNamed named returns the Servo with the given name
func (cmd *BaseCommand) ServoNamed(name string) *Servo {
	servo, _ := cmd.lookupServo(name)
	return servo
}

// AddServo adds a Servo to the config
func (cmd *BaseCommand) AddServo(servo Servo) error {
	servos, err := cmd.Servos()
	if err != nil {
		return err
	}

	servos = append(servos, servo)
	cmd.viperCfg.Set("servos", servos)
	return cmd.viperCfg.WriteConfig()
}

// RemoveServoNamed removes a Servo from the config with the given name
func (cmd *BaseCommand) RemoveServoNamed(name string) error {
	s, index := cmd.lookupServo(name)
	if s == nil {
		return fmt.Errorf("no such Servo %q", name)
	}
	servos, err := cmd.Servos()
	if err != nil {
		return err
	}
	servos = append(servos[:index], servos[index+1:]...)
	cmd.viperCfg.Set("servos", servos)
	return cmd.viperCfg.WriteConfig()
}

// RemoveServo removes a Servo from the config
func (cmd *BaseCommand) RemoveServo(servo Servo) error {
	return cmd.RemoveServoNamed(servo.Name)
}

// Servo represents a deployed Servo assembly running somewhere
type Servo struct {
	Name string
	User string
	Host string
	Port string
	Path string
}

func (s Servo) HostAndPort() string {
	h := s.Host
	p := s.Port
	if p == "" {
		p = "22"
	}
	return strings.Join([]string{h, p}, ":")
}

func (s Servo) DisplayHost() string {
	v := s.Host
	if s.Port != "" && s.Port != "22" {
		v = v + ":" + s.Port
	}
	return v
}

func (s Servo) DisplayPath() string {
	if s.Path != "" {
		return s.Path
	}
	return "~/"
}

func (s Servo) URL() string {
	pathComponent := ""
	if s.Path != "" {
		if s.Port != "" && s.Port != "22" {
			pathComponent = pathComponent + ":"
		}
		pathComponent = pathComponent + s.Path
	}
	return fmt.Sprintf("ssh://%s@%s:%s", s.User, s.DisplayHost(), pathComponent)
}
