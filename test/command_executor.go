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
	"bytes"

	"github.com/spf13/cobra"
)

// NewCommandExecutor returns an executor for testing Cobra commands
func NewCommandExecutor(rootCmd *cobra.Command) *CommandExecutor {
	return &CommandExecutor{
		rootCmd: rootCmd,
	}
}

// CommandExecutor provides an interface for executing Cobra commands in tests
type CommandExecutor struct {
	rootCmd *cobra.Command
}

// Execute runs a Cobra command with the given arguments and returns the output captured
func (ce *CommandExecutor) Execute(args ...string) (output string, err error) {
	_, output, err = ce.ExecuteC(args...)
	return output, err
}

// ExecuteC runs a Cobra command with the given arguments and returns the Cobra command invoked and the output captured
func (ce *CommandExecutor) ExecuteC(args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	ce.rootCmd.SetOut(buf)
	ce.rootCmd.SetErr(buf)
	ce.rootCmd.SetArgs(args)

	c, err = ce.rootCmd.ExecuteC()
	return c, buf.String(), err
}
