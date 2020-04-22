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

// NewRootCommand returns a new instance of the root command for Opsani CLI
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
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
	rootCmd.MarkPersistentFlagFilename("config", "*.yaml", "*.yml")
	rootCmd.SetVersionTemplate("Opsani CLI version {{.Version}}\n")
	rootCmd.Flags().Bool("version", false, "Display version and exit")
	rootCmd.PersistentFlags().Bool("help", false, "Display help and exit")
	rootCmd.PersistentFlags().MarkHidden("help")
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	// Add all sub-commands
	rootCmd.AddCommand(NewInitCommand().Command)
	rootCmd.AddCommand(NewAppCommand())
	rootCmd.AddCommand(NewLoginCommand())

	rootCmd.AddCommand(newDiscoverCommand())
	rootCmd.AddCommand(newIMBCommand())
	rootCmd.AddCommand(newPullCommand())

	rootCmd.AddCommand(NewConfigCommand().Command)
	rootCmd.AddCommand(NewCompletionCommand().Command)

	// See Execute()
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return &FlagError{Err: err}
	})

	// Load configuration before execution of every action
	rootCmd.PersistentPreRunE = ReduceRunEFuncs(InitConfigRunE, RequireConfigFileFlagToExistRunE)

	return rootCmd
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
func Execute() (cmd *cobra.Command, err error) {
	rootCmd := NewRootCommand()

	if err := initConfig(); err != nil {
		rootCmd.PrintErr(err)
		return rootCmd, err
	}

	executedCmd, err := rootCmd.ExecuteC()
	if err != nil {
		// Exit silently if the user bailed with control-c
		if errors.Is(err, terminal.InterruptErr) {
			return executedCmd, err
		}

		executedCmd.PrintErrf("%s: %s\n", executedCmd.Name(), err)

		// Display usage for invalid command and flag errors
		var flagError *FlagError
		if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
			if !strings.HasSuffix(err.Error(), "\n") {
				executedCmd.PrintErrln()
			}
			executedCmd.PrintErrln(executedCmd.UsageString())
		}
	}
	return executedCmd, err
}

// RunFunc is a Cobra Run function
type RunFunc func(cmd *cobra.Command, args []string)

// RunEFunc is a Cobra Run function that returns an error
type RunEFunc func(cmd *cobra.Command, args []string) error

// ReduceRunFuncs reduces a list of Cobra run functions into a single aggregate run function
func ReduceRunFuncs(runFuncs ...RunFunc) RunFunc {
	return func(cmd *cobra.Command, args []string) {
		for _, runFunc := range runFuncs {
			runFunc(cmd, args)
		}
	}
}

// ReduceRunEFuncs reduces a list of Cobra run functions that return an error into a single aggregate run function
func ReduceRunEFuncs(runFuncs ...RunEFunc) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, runFunc := range runFuncs {
			if err := runFunc(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// InitConfigRunE initializes client configuration and aborts execution if an error is encountered
func InitConfigRunE(cmd *cobra.Command, args []string) error {
	return initConfig()
}

// RequireConfigFileFlagToExistRunE aborts command execution with an error if the config file specified via a flag does not exist
func RequireConfigFileFlagToExistRunE(cmd *cobra.Command, args []string) error {
	if configFilePath, err := cmd.Root().PersistentFlags().GetString("config"); err == nil {
		if configFilePath != "" {
			if _, err := os.Stat(opsani.ConfigFile); os.IsNotExist(err) {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

// RequireInitRunE aborts command execution with an error if the client is not initialized
func RequireInitRunE(cmd *cobra.Command, args []string) error {
	if !opsani.IsInitialized() {
		return fmt.Errorf("command failed because client is not initialized. Run %q and try again", "opsani init")
	}

	return nil
}

func initConfig() error {
	if opsani.ConfigFile != "" {
		// Use config file from the flag. (TODO: Should we check if the file exists unless we are running init?)
		viper.SetConfigFile(opsani.ConfigFile)
	} else {
		// Find Opsani config in home directory
		viper.AddConfigPath(opsani.DefaultConfigPath())
		viper.SetConfigName("config")
		viper.SetConfigType(opsani.DefaultConfigType())
		opsani.ConfigFile = opsani.DefaultConfigFile()
	}

	// Set up environment variables
	viper.SetEnvPrefix(opsani.KeyEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Load the configuration
	if err := viper.ReadInConfig(); err == nil {
		opsani.ConfigFile = viper.ConfigFileUsed()
	} else {

		switch err.(type) {
		case *os.PathError:
		case *viper.ConfigFileNotFoundError:
			// Ignore missing config files
		default:
			return fmt.Errorf("error parsing configuration file: %w", err)
		}
	}
	return nil
}

// NewAPIClientFromConfig returns an Opsani API client configured using the active configuration
func NewAPIClientFromConfig() *opsani.Client {
	c := opsani.NewClient().
		SetBaseURL(opsani.GetBaseURL()).
		SetApp(opsani.GetApp()).
		SetAuthToken(opsani.GetAccessToken()).
		SetDebug(opsani.GetDebugModeEnabled())
	if opsani.GetRequestTracingEnabled() {
		c.EnableTrace()
	}

	// Set the output directory to pwd by default
	if dir, err := os.Getwd(); err == nil {
		c.SetOutputDirectory(dir)
	}
	return c
}
