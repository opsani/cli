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

var baseURL string
var cfgFile string
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "https://api.opsani.com/", "Base URL for accessing the Opsani API")
	rootCmd.PersistentFlags().MarkHidden("base-url")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Location of config file (default \"%s\")", filepath.Join(home, ".opsani", "config.yaml")))
	rootCmd.PersistentFlags().BoolVarP(&printVersion, "version", "v", false, "Print version information and quit")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".opsani/config.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
