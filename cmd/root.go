/*
Copyright © 2020 Blake Watters <blake@opsani.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opsani",
	Short: "The official CLI for Opsani",
	Long: `Work with Opsani from the command line.

Opsani CLI is in early stages of development. 
We'd love to hear your feedback at <https://github.com/opsani/cli>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       "0.0.1",
}

// FlagError is the kind of error raised in flag processing
type FlagError struct {
	Err error
}

func (fe FlagError) Error() string {
	return fe.Err.Error()
}

func (fe FlagError) Unwrap() error {
	return fe.Err
}

func subCommandPath(rootCmd *cobra.Command, cmd *cobra.Command) string {
	path := make([]string, 0, 1)
	currentCmd := cmd
	if rootCmd == cmd {
		return ""
	}
	for {
		path = append([]string{currentCmd.Name()}, path...)
		if currentCmd.Parent() == rootCmd {
			return strings.Join(path, " ")
		}
		currentCmd = currentCmd.Parent()
	}
}

// Execute is the entry point for executing all commands from main
// All commands with RunE will bubble errors back here
func Execute() {
	if cmd, err := rootCmd.ExecuteC(); err != nil {
		// Exit silently if the user bailed with control-c
		if err == terminal.InterruptErr {
			os.Exit(0)
		}

		if rootCmd != cmd {
			fmt.Fprintf(os.Stderr, "%q %s\n", subCommandPath(rootCmd, cmd), err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}

		// Display usage for invalid command and flag errors
		var flagError *FlagError
		if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Fprintln(os.Stderr)
			}
			fmt.Fprintln(os.Stderr, cmd.UsageString())
		}
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String(opsani.KeyBaseURL, opsani.DefaultBaseURL, "Base URL for accessing the Opsani API")
	rootCmd.PersistentFlags().MarkHidden(opsani.KeyBaseURL)
	viper.BindPFlag(opsani.KeyBaseURL, rootCmd.PersistentFlags().Lookup(opsani.KeyBaseURL))
	rootCmd.PersistentFlags().String(opsani.KeyApp, "", "App to control (overrides config file and OPSANI_APP)")
	viper.BindPFlag(opsani.KeyApp, rootCmd.PersistentFlags().Lookup(opsani.KeyApp))
	rootCmd.PersistentFlags().String(opsani.KeyToken, "", "API token to authenticate with (overrides config file and OPSANI_TOKEN)")
	viper.BindPFlag(opsani.KeyToken, rootCmd.PersistentFlags().Lookup(opsani.KeyToken))
	rootCmd.PersistentFlags().BoolP(opsani.KeyDebugMode, "D", false, "Enable debug mode")
	viper.BindPFlag(opsani.KeyDebugMode, rootCmd.PersistentFlags().Lookup(opsani.KeyDebugMode))
	rootCmd.PersistentFlags().Bool(opsani.KeyRequestTracing, false, "Enable request tracing")
	viper.BindPFlag(opsani.KeyRequestTracing, rootCmd.PersistentFlags().Lookup(opsani.KeyRequestTracing))

	rootCmd.PersistentFlags().StringVar(&opsani.ConfigFile, "config", "", fmt.Sprintf("Location of config file (default \"%s\")", opsani.DefaultConfigFile()))
	rootCmd.SetVersionTemplate("Opsani CLI version {{.Version}}\n")

	// See Execute()
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return &FlagError{Err: err}
	})
}

func initConfig() {
	if opsani.ConfigFile != "" {
		// Use config file from the flag. (TODO: Should we check if the file exists unless we are running init?)
		viper.SetConfigFile(opsani.ConfigFile)
	} else {
		// Find Opsani config in home directory
		viper.AddConfigPath(opsani.DefaultConfigPath())
		viper.SetConfigName("config")
		viper.SetConfigType(opsani.DefaultConfigType())
	}

	// Set up environment variables
	viper.SetEnvPrefix(opsani.KeyEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Load the configuration
	if err := viper.ReadInConfig(); err == nil {
		opsani.ConfigFile = viper.ConfigFileUsed()
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			opsani.ConfigFile = opsani.DefaultConfigFile()
		} else {
			fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing configuration file: %s", err))
			os.Exit(1)
		}
	}
}
