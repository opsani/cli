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
package opsani

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// ConfigFile stores the path to the active Opsani configuration file
var ConfigFile string = DefaultConfigFile()

// DefaultConfigFile returns the full path to the default Opsani configuration file
func DefaultConfigFile() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return filepath.Join(home, ".opsani", "config.yaml")
}

// DefaultConfigPath returns the path to the directory storing the Opsani configuration file
func DefaultConfigPath() string {
	return filepath.Dir(DefaultConfigFile())
}

// DefaultConfigType returns the
func DefaultConfigType() string {
	return "yaml"
}

// The Config struct wraps Viper configuration
type Config struct{}

// GetBaseURL returns the Opsani API base URL
func GetBaseURL() string {
	return viper.GetString(KeyBaseURL)
}

// GetBaseURLHostnameAndPort returns the hostname and port portion of Opsani base URL for summary display
func GetBaseURLHostnameAndPort() string {
	u, err := url.Parse(GetBaseURL())
	if err != nil {
		panic(err)
	}
	baseURLDescription := u.Hostname()
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		baseURLDescription = baseURLDescription + ":" + port
	}
	return baseURLDescription
}

// SetBaseURL sets the Opsani API base URL
func SetBaseURL(baseURL string) {
	viper.Set(KeyBaseURL, baseURL)
}

// GetAccessToken returns the Opsani API access token
func GetAccessToken() string {
	return viper.GetString(KeyToken)
}

// SetAccessToken sets the Opsani API access token
func SetAccessToken(accessToken string) {
	viper.Set(KeyToken, accessToken)
}

// GetApp returns the target Opsani app
func GetApp() string {
	return viper.GetString(KeyApp)
}

// SetApp sets the target Opsani app
func SetApp(app string) {
	viper.Set(KeyApp, app)
}

// GetAllSettings returns all configuration settings
func GetAllSettings() map[string]interface{} {
	return viper.AllSettings()
}

// GetDebugModeEnabled returns a boolean value indicating if debugging is enabled
func GetDebugModeEnabled() bool {
	return viper.GetBool(KeyDebugMode)
}

// GetRequestTracingEnabled returns a boolean value indicating if request tracing is enabled
func GetRequestTracingEnabled() bool {
	return viper.GetBool(KeyRequestTracing)
}
