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

package cmd

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
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

// NewOpsaniCommandExecutor returns an executor for testing Opsani CLI commands
func NewOpsaniCommandExecutor(rootCmd *cobra.Command) *OpsaniCommandExecutor {
	return &OpsaniCommandExecutor{
		CommandExecutor: NewCommandExecutor(rootCmd),
	}
}

// OpsaniCommandExecutor provides an interface for executing Opsani CLI commands in tests
type OpsaniCommandExecutor struct {
	*CommandExecutor
}

// ExecuteWithConfig runs an Opsani CLI command with the given config file and arguments and returns the output captured
func (oce *OpsaniCommandExecutor) ExecuteWithConfig(configFile *os.File, args ...string) (output string, err error) {
	return oce.Execute(append([]string{"--config", configFile.Name()}, args...)...)
}

// ExecuteCWithConfig runs an Opsani CLI command with the given config file and arguments and returns the Opsani CLI command invoked
func (oce *OpsaniCommandExecutor) ExecuteCWithConfig(configFile *os.File, args ...string) (c *cobra.Command, output string, err error) {
	return oce.ExecuteC(append([]string{"--config", configFile.Name()}, args...)...)
}

// TempConfigFileWithBytes returns a temporary YAML config file with the given byte array content
func TempConfigFileWithBytes(bytes []byte) *os.File {
	tmpFile, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		panic("failed to create temp file")
	}
	if _, err = tmpFile.Write(bytes); err != nil {
		panic("failed writing to temp file")
	}
	return tmpFile
}

// TempConfigFileWithString returns a temporary YAML config file with the given string content
func TempConfigFileWithString(str string) *os.File {
	return TempConfigFileWithBytes([]byte(str))
}

// TempConfigFileWithObj returns a temporary YAML config file with the given object serialized to YAML
func TempConfigFileWithObj(obj interface{}) *os.File {
	if data, err := yaml.Marshal(obj); data != nil {
		return TempConfigFileWithBytes(data)
	} else if err != nil {
		panic("failed serializing config to YAML")
	}
	return nil
}
