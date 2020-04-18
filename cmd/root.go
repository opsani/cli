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
	"fmt"
	"os"
	"strings"

	"github.com/opsani/cli/opsani"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var printVersion bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opsani",
	Short: "The official CLI for Opsani",
	Long: `Work with Opsani from the command line.

Opsani CLI is in early stages of development. 
We'd love to hear your feedback at <https://github.com/opsani/cli>`,
	Run: func(cmd *cobra.Command, args []string) {
		if printVersion {
			fmt.Println("Opsani CLI version 0.0.1")
			os.Exit(0)
		}

		// If we aren't printing the version display usage and exit
		cmd.Help()
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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
	rootCmd.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "Print version information and quit")
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
			panic(fmt.Errorf("error parsing configuration file: %s", err))
		}
	}
}
