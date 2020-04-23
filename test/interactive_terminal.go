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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/opsani/cli/command"
	"github.com/spf13/cobra"
)

// TODO: Rename to PassthroughTty
// PassthroughPipeFile wraps a file with a PassthroughPipe to add read deadline support
type PassthroughPipeFile struct {
	*expect.PassthroughPipe
	file *os.File
}

// Write proxies the write operation to the underlying file
func (s *PassthroughPipeFile) Write(p []byte) (n int, err error) {
	return s.file.Write(p)
}

// Fd proxies the file descriptor of the underyling file
// This is necessary because the survey library treats the stdio descriptors as
// concrete os.File instances instead of reader/writer interfaces
func (s *PassthroughPipeFile) Fd() uintptr {
	return s.file.Fd()
}

// NewPassthroughPipeFile returns a new PassthroughPipeFile that wraps the input file to enable read deadline support
func NewPassthroughPipeFile(inFile *os.File) (*PassthroughPipeFile, error) {
	file := os.NewFile(inFile.Fd(), "pipe")
	if file == nil {
		fmt.Errorf("os.NewFile failed: is your file descriptor valid?")
	}
	pipe, err := expect.NewPassthroughPipe(inFile)
	if err != nil {
		return nil, err
	}
	return &PassthroughPipeFile{
		file:            file,
		PassthroughPipe: pipe,
	}, nil
}

// RunTestInInteractiveTerminal runs a test within an interactive terminal environment
// Executin requires a standard test instance, a pair of functions that execute the code
// under test and the test code, and any desired options for configuring the virtual terminal environment
func RunTestInInteractiveTerminal(t *testing.T,
	codeUnderTestFunc InteractiveProcessFunc,
	testFunc InteractiveUserFunc,
	consoleOpts ...expect.ConsoleOpt) (*InteractiveExecutionContext, error) {
	context, err := ExecuteInInteractiveTerminal(codeUnderTestFunc, testFunc, consoleOpts...)
	t.Logf("Raw output: %q", context.OutputBuffer().String())
	t.Logf("\n\nterminal state: %s", expect.StripTrailingEmptyLines(context.TerminalState().String()))
	return context, err
}

// InteractiveExecutionContext describes the state of an interactive terminal execution
type InteractiveExecutionContext struct {
	outputBuffer   *bytes.Buffer
	terminalState  *vt10x.State
	console        *expect.Console
	passthroughTty *PassthroughPipeFile
}

// Tty returns the io.Reader to be used as stdin during execution
// func (ice *InteractiveExecutionContext) Tty() *PassthroughPipeFile {
// 	return ice.passthroughTty
// }
func (ice *InteractiveExecutionContext) Tty() *os.File {
	// return ice.passthroughTty
	return ice.console.Tty()
}

// Stdin returns the io.Reader to be used as stdin during execution
func (ice *InteractiveExecutionContext) Stdin() io.Reader {
	// if ice.passthroughTty == nil {

	// }
	// return ice.passthroughTty
	// // return ice.passthroughTty
	return ice.console.Tty()
}

// Stdout returns the io.Writer to be used as stdout during execution
func (ice *InteractiveExecutionContext) Stdout() io.Writer {
	// return ice.passthroughTty
	return ice.console.Tty()
}

// Stderr returns the io.Writer to be used as stdout during execution
func (ice *InteractiveExecutionContext) Stderr() io.Writer {
	// return ice.passthroughTty
	return ice.console.Tty()
}

// OutputBuffer returns the output buffer read from the tty
func (ice *InteractiveExecutionContext) OutputBuffer() *bytes.Buffer {
	return ice.outputBuffer
}

// TerminalState returns the state if the terminal
func (ice *InteractiveExecutionContext) TerminalState() *vt10x.State {
	return ice.terminalState
}

// Console returns the console for interacting with the terminal
func (ice *InteractiveExecutionContext) Console() *expect.Console {
	return ice.console
}

func (ice *InteractiveExecutionContext) PassthroughTty() *PassthroughPipeFile {
	// Wrap the Tty into a PassthroughPipeFile to enable deadline support (NewPassthroughPipeFileReader??)
	if ice.passthroughTty == nil {
		passthroughTty, err := NewPassthroughPipeFile(ice.Tty())
		// TODO: SETTING THE DEADLINE IS CRITICAL
		passthroughTty.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		if err != nil {
			panic(err)
		}
		ice.passthroughTty = passthroughTty
	}
	return ice.passthroughTty
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
			expect.WithDefaultTimeout(100 * time.Millisecond),
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
		// passthroughTty: passthroughTty,
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
		ice.command.SetIn(context.Stdin())
		ice.command.SetOut(context.Stdout())
		ice.command.SetErr(context.Stderr())
		ice.command.SetArgs(args)

		command.SetStdio(terminal.Stdio{In: context.PassthroughTty(), Out: context.PassthroughTty(), Err: context.PassthroughTty()})

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
