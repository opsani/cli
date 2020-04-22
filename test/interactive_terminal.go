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
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/spf13/cobra"
)

// RunTestInInteractiveTerminal runs a test within an interactive terminal environment
// Executin requires a standard test instance, a pair of functions that execute the code
// under test and the test code, and any desired options for configuring the virtual terminal environment
func RunTestInInteractiveTerminal(t *testing.T,
	codeUnderTestFunc InteractiveProcessFunc,
	testFunc InteractiveUserFunc,
	consoleOpts ...expect.ConsoleOpt) (*InteractiveExecutionContext, error) {
	context, err := ExecuteInInteractiveTerminal(codeUnderTestFunc, testFunc, consoleOpts...)
	t.Logf("Raw output: %q", context.GetOutputBuffer().String())
	t.Logf("\n\nterminal state: %s", expect.StripTrailingEmptyLines(context.GetTerminalState().String()))
	return context, err
}

type InteractiveExecutionContext struct {
	outputBuffer  *bytes.Buffer
	terminalState *vt10x.State
	console       *expect.Console
}

func (ice *InteractiveExecutionContext) GetStdin() *os.File {
	return ice.console.Tty()
}

func (ice *InteractiveExecutionContext) GetStdout() *os.File {
	return ice.console.Tty()
}

func (ice *InteractiveExecutionContext) GetStderr() *os.File {
	return ice.console.Tty()
}

func (ice *InteractiveExecutionContext) GetOutputBuffer() *bytes.Buffer {
	return ice.outputBuffer
}

func (ice *InteractiveExecutionContext) GetTerminalState() *vt10x.State {
	return ice.terminalState
}

func (ice *InteractiveExecutionContext) GetConsole() *expect.Console {
	return ice.console
}

// Args is a convenience function that converts a variadic list of strings into an array
func Args(args ...string) []string {
	return args
}

// InteractiveProcessFunc instances are functions that represent the process side of an interactive terminal
type InteractiveProcessFunc func(*InteractiveExecutionContext) error

// InteractiveUserFunc iinstances are functions that represent the user side of an interactive terminal
type InteractiveUserFunc func(*InteractiveExecutionContext, *expect.Console) error

// InteractiveCommandExecutor executes a Cobra command interactively within a virtual terminal
// The command executor orchestrates an underlying expect.Console and vt10x.VT virtual terminal and executes the
// target command within the environment. This allows for interaction with a command that is performing readline operations,
// emitting ANSI output, and otherwise exposing an interactive user experience.
type InteractiveCommandExecutor struct {
	command           *cobra.Command
	consoleOpts       []expect.ConsoleOpt
	PreExecutionFunc  InteractiveProcessFunc
	PostExecutionFunc InteractiveProcessFunc
}

// NewInteractiveCommandExecutor returns a new command executor for working with interactive terminal commands
func NewInteractiveCommandExecutor(command *cobra.Command, consoleOpts ...expect.ConsoleOpt) *InteractiveCommandExecutor {
	return &InteractiveCommandExecutor{
		consoleOpts: consoleOpts,
		command:     command,
	}
}

// SetTimeout sets the timeout for command execution
func (ice *InteractiveCommandExecutor) SetTimeout(timeout time.Duration) {
	ice.consoleOpts = append(ice.consoleOpts, expect.WithDefaultTimeout(timeout))
}

// ExecuteInInteractiveTerminal runs a pair of functions connected in an interactive virtual terminal environment
func ExecuteInInteractiveTerminal(
	processFunc InteractiveProcessFunc, // Represents the process that the user is interacting with via the terminal
	userFunc InteractiveUserFunc, // Represents the user interacting with the process
	consoleOpts ...expect.ConsoleOpt) (*InteractiveExecutionContext, error) {
	outputBuffer := new(bytes.Buffer)
	console, terminalState, err := vt10x.NewVT10XConsole(
		append([]expect.ConsoleOpt{
			expect.WithStdout(outputBuffer),
			expect.WithDefaultTimeout(300 * time.Millisecond),
		}, consoleOpts...)...)
	if err != nil {
		return nil, err
	}
	defer console.Close()

	// Create the execution context
	executionContext := &InteractiveExecutionContext{
		outputBuffer:  outputBuffer,
		console:       console,
		terminalState: terminalState,
	}

	// Execute our function within a channel and wait for exit
	exitChannel := make(chan struct{})
	go func() {
		defer close(exitChannel)
		userFunc(executionContext, console)
	}()

	// Run the process for the user to interact with
	err = processFunc(executionContext)
	if err != nil {
		fmt.Println("Process failed", err)
	}

	// Close the slave end of the pty, and read the remaining bytes from the master end.
	console.Tty().Close()
	<-exitChannel

	return executionContext, err
}

// Execute runs the specified command interactively and returns an execution context object upon completion
func (ice *InteractiveCommandExecutor) Execute(args []string, interactionFunc InteractiveUserFunc) (*InteractiveExecutionContext, error) {
	// Wrap our execution func with setup for Command execution
	commandExecutionFunc := func(context *InteractiveExecutionContext) error {
		ice.command.SetIn(context.GetStdin())
		ice.command.SetOut(context.GetStdout())
		ice.command.SetErr(context.GetStderr())
		ice.command.SetArgs(args)

		if ice.PreExecutionFunc != nil {
			ice.PreExecutionFunc(context)
		}
		_, err := ice.command.ExecuteC()
		if ice.PostExecutionFunc != nil {
			ice.PostExecutionFunc(context)
		}
		return err
	}

	return ExecuteInInteractiveTerminal(commandExecutionFunc, interactionFunc, ice.consoleOpts...)
}

// ExecuteString executes the target command by splitting the args string at space boundaries
// This is a convenience interface suitable only for simple arguments that do not contain quoted values or literals
// If you need something more advanced please use the Execute() and Args() method to compose from a variadic list of arguments
func (ice *InteractiveCommandExecutor) ExecuteString(argsStr string, interactionFunc InteractiveUserFunc) (*InteractiveExecutionContext, error) {
	return ice.Execute(strings.Split(argsStr, " "), interactionFunc)
}
