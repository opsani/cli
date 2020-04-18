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
	"path/filepath"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Configuration options bound via Cobra
var opsaniConfig = struct {
	BaseURL    string
	ConfigFile string
	App        string
	Token      string
}{}
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

func defaultConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".opsani", "config.yaml")
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&opsaniConfig.BaseURL, "base-url", "https://api.opsani.com/", "Base URL for accessing the Opsani API")
	rootCmd.PersistentFlags().MarkHidden("base-url")
	viper.BindPFlag("base-url", rootCmd.PersistentFlags().Lookup("base-url"))
	rootCmd.PersistentFlags().StringVar(&opsaniConfig.ConfigFile, "config", "", fmt.Sprintf("Location of config file (default \"%s\")", defaultConfigFile()))
	rootCmd.PersistentFlags().StringVar(&opsaniConfig.App, "app", "", "App to control (overrides config file and OPSANI_APP)")
	viper.BindPFlag("app", rootCmd.PersistentFlags().Lookup("app"))
	rootCmd.PersistentFlags().StringVar(&opsaniConfig.Token, "token", "", "API token to authenticate with (overrides config file and OPSANI_TOKEN)")
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	rootCmd.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "Print version information and quit")
}

func initConfig() {
	if opsaniConfig.ConfigFile != "" {
		// Use config file from the flag. (TODO: Should we check if the file exists unless we are running init?)
		viper.SetConfigFile(opsaniConfig.ConfigFile)
	} else {
		// Find Opsani config in home directory
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(filepath.Join(home, ".opsani"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		opsaniConfig.ConfigFile = defaultConfigFile()
	}

	// Set up environment variables
	viper.SetEnvPrefix("OPSANI")
	viper.AutomaticEnv()

	// Load the configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			panic(fmt.Errorf("error parsing configuration file: %s", err))
		}
	}
}
