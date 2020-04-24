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

package test

import (
	"os"

	"github.com/spf13/cobra"
)

// OpsaniCommandExecutor provides an interface for executing Opsani CLI commands in tests
type OpsaniCommandExecutor struct {
	*CommandExecutor
}

// ExecuteWithConfig runs an Opsani CLI command with the given config file and arguments and returns the output captured
func (oce *OpsaniCommandExecutor) ExecuteWithConfig(configFile *os.File, args ...string) (output string, err error) {
	return oce.ExecuteCommand(append([]string{"--config", configFile.Name()}, args...)...)
}

// ExecuteCWithConfig runs an Opsani CLI command with the given config file and arguments and returns the Opsani CLI command invoked
func (oce *OpsaniCommandExecutor) ExecuteCWithConfig(configFile *os.File, args ...string) (c *cobra.Command, output string, err error) {
	defer os.Remove(configFile.Name())
	return oce.ExecuteCommandC(append([]string{"--config", configFile.Name()}, args...)...)
}

// NewOpsaniCommandExecutor returns an executor for testing Opsani CLI commands
func NewOpsaniCommandExecutor(rootCmd *cobra.Command) *OpsaniCommandExecutor {
	return &OpsaniCommandExecutor{
		CommandExecutor: NewCommandExecutor(rootCmd),
	}
}
