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
	"strings"
	"time"

	"github.com/opsani/cli/command"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// Suite provides a single struct for importing all testing functionality
// with sensible defaults with disambiguated names designed to read clearly in
// test cases.
type Suite struct {
	suite.Suite
	cmd *cobra.Command

	ce  CommandExecutor
	ice *InteractiveCommandExecutor
	ict *InteractiveCommandTester

	args []string
}

// Command returns the Cobra command under test
func (h *Suite) Command() *cobra.Command {
	if h.cmd == nil {
		panic("invalid configuration: the Command instance cannot be nil")
	}
	return h.cmd
}

// SetCommand sets the Opsani command under test
func (h *Suite) SetCommand(cmd *command.BaseCommand) {
	h.SetCobraCommand(cmd.RootCobraCommand())
}

// SetCobraCommand sets the Cobra command under test
// Changing the command will reset the associated command executor and tester instances
func (h *Suite) SetCobraCommand(cmd *cobra.Command) {
	if cmd != nil {
		cmdExecutor := NewCommandExecutor(cmd)
		ice := NewInteractiveCommandExecutor(cmd)
		cmdTester := NewInteractiveCommandTester(cmd, ice)

		h.cmd = cmd
		h.ce = *cmdExecutor
		h.ice = ice
		h.ict = cmdTester
	} else {
		h.cmd = cmd
		h.ice = nil
		h.ict = nil
	}
}

// InteractiveCommandExecutor returns the interactive command executor
func (h *Suite) InteractiveCommandExecutor() *InteractiveCommandExecutor {
	if h.ice == nil {
		h.ice = NewInteractiveCommandExecutor(h.Command())
	}
	return h.ice
}

// InteractiveCommandTester returns the interactive command tester
func (h *Suite) InteractiveCommandTester() *InteractiveCommandTester {
	if h.ict == nil {
		h.ict = NewInteractiveCommandTester(h.Command(), h.InteractiveCommandExecutor())
	}
	return h.ict
}

// SetInteractiveExecutionTimeout sets the timeout for interactive command executions and tests
func (h *Suite) SetInteractiveExecutionTimeout(timeout time.Duration) {
	h.InteractiveCommandExecutor().SetTimeout(timeout)
}

// InteractiveExecutionTimeout returns the timeout for interactive command executions and tests
func (h Suite) InteractiveExecutionTimeout() time.Duration {
	return h.InteractiveCommandExecutor().Timeout()
}

// DefaultArgs returns the default arguments
func (h Suite) DefaultArgs() []string {
	return h.args
}

// SetDefaultArgs sets the default arguments
func (h *Suite) SetDefaultArgs(args []string) {
	h.args = args
}

// AddDefaultArg adds a default argument
func (h *Suite) AddDefaultArg(arg string) {
	h.args = append(h.args, arg)
}

// AddDefaultArgs adds a series of default arguments
func (h *Suite) AddDefaultArgs(args ...string) {
	h.args = append(h.args, args...)
}

///
/// Command Execution Functions
///

// Execute runs a Cobra command with the given arguments and returns the output captured
// Default arguments are prepended before execution
func (h *Suite) Execute(args ...string) (string, error) {
	_, output, err := h.ExecuteC(args...)
	return output, err
}

// ExecuteArgs runs a Cobra command with the given arguments and returns the output captured
// Default arguments are prepended before execution
func (h *Suite) ExecuteArgs(args []string) (string, error) {
	_, output, err := h.ExecuteC(args...)
	return output, err
}

// ExecuteC runs a Cobra command with the given arguments and returns the Cobra command invoked and the output captured
// Default arguments are prepended before execution
func (h *Suite) ExecuteC(args ...string) (*cobra.Command, string, error) {
	return h.ce.ExecuteC(h.appendArgsToDefaults(args)...)
}

// ExecuteString executes the target command by splitting the args string at space boundaries
// This is a convenience interface suitable only for simple arguments that do not contain quoted values or literals
// If you need something more advanced please use the Execute() and Args() method to compose from a variadic list of arguments
// Default arguments are prepended before execution
func (h *Suite) ExecuteString(argsStr string) (string, error) {
	return h.Execute(h.splitArgsStringAndAppendToDefaults(argsStr)...)
}

// ExecuteInteractively executes a command using the given arguments interactively
// Default arguments are prepended before execution
func (h *Suite) ExecuteInteractively(args []string, interactionFunc InteractiveUserFunc) (*InteractiveExecutionContext, error) {
	return h.InteractiveCommandExecutor().Execute(h.appendArgsToDefaults(args), interactionFunc)
}

// ExecuteStringInteractively splits the input string at space boundaires and executes the resulting arguments interactively
// Default arguments are prepended before execution
func (h *Suite) ExecuteStringInteractively(argsStr string, interactionFunc InteractiveUserFunc) (*InteractiveExecutionContext, error) {
	return h.ice.Execute(h.splitArgsStringAndAppendToDefaults(argsStr), interactionFunc)
}

// ExecuteTestInteractively executes the given arguments in an interactive test context
// Default arguments are prepended before execution
func (h *Suite) ExecuteTestInteractively(args []string, testFunc func(*InteractiveTestContext) error) (*InteractiveExecutionContext, error) {
	return h.ict.Execute(h.T(), h.appendArgsToDefaults(args), testFunc)
}

// ExecuteTestStringInteractively splits the input string at space boundaires and executes the resulting arguments in an interactive test context
// Default arguments are prepended before execution
func (h *Suite) ExecuteTestStringInteractively(argsStr string, testFunc func(*InteractiveTestContext) error) (*InteractiveExecutionContext, error) {
	return h.ict.Execute(h.T(), h.splitArgsStringAndAppendToDefaults(argsStr), testFunc)
}

///
/// Unexported Utility Methods
///

func (h *Suite) appendArgsToDefaults(args []string) []string {
	return append(h.DefaultArgs(), args...)
}

func (h *Suite) splitArgsStringAndAppendToDefaults(argsStr string) []string {
	return h.appendArgsToDefaults(strings.Split(argsStr, " "))
}
