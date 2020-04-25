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
	"github.com/alecthomas/assert"
	"github.com/hinshun/vt10x"
	"github.com/opsani/cli/command"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

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
		return nil, fmt.Errorf("os.NewFile failed: is your file descriptor valid?")
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

// ExecuteInInteractiveConsole runs a pair of functions connected in an interactive virtual terminal environment
func ExecuteInInteractiveConsole(
	processFunc InteractiveProcessFunc, // Represents the process that the user is interacting with via the terminal
	userFunc InteractiveUserFunc, // Represents the user interacting with the process
	consoleOpts ...expect.ConsoleOpt) (*InteractiveExecutionContext, error) {
	consoleObserver := new(consoleObserver)
	closerProxy := new(closerProxy) // Create a proxy object to close our Tty proxy later
	outputBuffer := new(bytes.Buffer)
	consoleOpts = append([]expect.ConsoleOpt{
		expect.WithStdout(outputBuffer),
		expect.WithExpectObserver(consoleObserver.observeExpect),
		expect.WithSendObserver(consoleObserver.observeSend),
		expect.WithCloser(closerProxy),
		expect.WithDefaultTimeout(250 * time.Millisecond),
	}, consoleOpts...)
	console, terminalState, err := vt10x.NewVT10XConsole(consoleOpts...)
	if err != nil {
		return nil, err
	}
	defer console.Close()

	// Use the same timeout in effect on the slave (user) side on the master (process) side pf the PTY
	timeoutOpts := expect.ConsoleOpts{}
	for _, opt := range consoleOpts {
		if err := opt(&timeoutOpts); err != nil {
			return nil, err
		}
	}
	consoleObserver.readTimeout = timeoutOpts.ReadTimeout

	// Create the execution context
	executionContext := &InteractiveExecutionContext{
		outputBuffer:    outputBuffer,
		console:         console,
		terminalState:   terminalState,
		closerProxy:     closerProxy,
		consoleObserver: consoleObserver,
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

// InteractiveExecutionContext describes the state of an interactive terminal execution
type InteractiveExecutionContext struct {
	outputBuffer    *bytes.Buffer
	terminalState   *vt10x.State
	console         *expect.Console
	passthroughTty  *PassthroughPipeFile
	closerProxy     *closerProxy
	consoleObserver *consoleObserver
}

// ReadTimeout returns the read time for the process side of an interactive execution
// expect.Console takes care of establishing a proxy pipe on the master side of the Tty
// but in a unit testing situation we have read failures on the slave side where the process
// may be waiting for input from the user
func (ice *InteractiveExecutionContext) ReadTimeout() time.Duration {
	return *ice.consoleObserver.readTimeout
}

// Tty returns the Tty of the underlying expect.Console instance
// You probably want to interact with the PassthroughTty which supports deadline based timeouts
func (ice *InteractiveExecutionContext) Tty() *os.File {
	return ice.console.Tty()
}

// Stdin returns the io.Reader to be used as stdin during execution
func (ice *InteractiveExecutionContext) Stdin() io.Reader {
	return ice.PassthroughTty()
}

// Stdout returns the io.Writer to be used as stdout during execution
func (ice *InteractiveExecutionContext) Stdout() io.Writer {
	return ice.PassthroughTty()
}

// Stderr returns the io.Writer to be used as stdout during execution
func (ice *InteractiveExecutionContext) Stderr() io.Writer {
	return ice.PassthroughTty()
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

// PassthroughTty returns a wrapper for the Tty that supports deadline based timeouts
// The Std* family of methods are all aliases for the passthrough tty
func (ice *InteractiveExecutionContext) PassthroughTty() *PassthroughPipeFile {
	// Wrap the Tty into a PassthroughPipeFile to enable deadline support (NewPassthroughPipeFileReader??)
	if ice.passthroughTty == nil {
		passthroughTty, err := NewPassthroughPipeFile(ice.Tty())
		if err != nil {
			panic(err)
		}
		ice.passthroughTty = passthroughTty
		ice.closerProxy.target = passthroughTty
		ice.consoleObserver.passthroughPipe = passthroughTty.PassthroughPipe
		ice.consoleObserver.extendDeadline()
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

// Command returns the Cobra command the executor was initialized with
// This is typically the root command in a Cobra application
func (ice *InteractiveCommandExecutor) Command() *cobra.Command {
	return ice.command
}

// SetTimeout sets the timeout for command execution
func (ice *InteractiveCommandExecutor) SetTimeout(timeout time.Duration) {
	ice.consoleOpts = append(ice.consoleOpts, expect.WithDefaultTimeout(timeout))
}

// Timeout return the timeout for command execution
// A zero value indicates that no timeout is in effect
func (ice *InteractiveCommandExecutor) Timeout() time.Duration {
	timeoutOpts := expect.ConsoleOpts{}
	for _, opt := range ice.consoleOpts {
		if err := opt(&timeoutOpts); err != nil {
			break
		}
	}
	if timeoutOpts.ReadTimeout == nil {
		return 0
	}
	return *timeoutOpts.ReadTimeout
}

// Closes out the passthrough proxy
type closerProxy struct {
	target io.Closer
}

func (cp *closerProxy) Close() error {
	if cp.target != nil {
		return cp.target.Close()
	}
	return nil
}

// Extends the read deadline on the passthrough pipe after expect/send operations
type consoleObserver struct {
	passthroughPipe *expect.PassthroughPipe
	readTimeout     *time.Duration
}

func (eo *consoleObserver) observeExpect(matchers []expect.Matcher, buf string, err error) {
	if eo.passthroughPipe == nil || eo.readTimeout == nil || err != nil {
		return
	}
	eo.extendDeadline()
}

func (eo *consoleObserver) observeSend(msg string, num int, err error) {
	if err != nil {
		eo.extendDeadline()
	}
}

func (eo *consoleObserver) extendDeadline() {
	if readTimeout := eo.readTimeout; readTimeout != nil {
		err := eo.passthroughPipe.SetReadDeadline(time.Now().Add(*readTimeout))
		if err != nil {
			panic(err)
		}
	}
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

	return ExecuteInInteractiveConsole(commandExecutionFunc, interactionFunc, ice.consoleOpts...)
}

// ExecuteS executes the target command by splitting the args string at space boundaries
// This is a convenience interface suitable only for simple arguments that do not contain quoted values or literals
// If you need something more advanced please use the Execute() and Args() method to compose from a variadic list of arguments
func (ice *InteractiveCommandExecutor) ExecuteS(argsStr string, interactionFunc InteractiveUserFunc) (*InteractiveExecutionContext, error) {
	return ice.Execute(strings.Split(argsStr, " "), interactionFunc)
}

// InteractiveTestContext instances are vended when an interactive test is run
type InteractiveTestContext struct {
	console *expect.Console
	context *InteractiveExecutionContext

	*assert.Assertions
	require *require.Assertions

	t *testing.T
}

// Console returns the the underlying terminal for direct interaction
func (ict *InteractiveTestContext) Console() *expect.Console {
	return ict.console
}

// T retrieves the current *testing.T context.
func (ict *InteractiveTestContext) T() *testing.T {
	return ict.t
}

// Require returns a require context for suite.
func (ict *InteractiveTestContext) Require() *require.Assertions {
	if ict.require == nil {
		ict.require = require.New(ict.T())
	}
	return ict.require
}

// Assert returns an assert context for suite.  Normally, you can call
// `suite.NoError(expected, actual)`, but for situations where the embedded
// methods are overridden (for example, you might want to override
// assert.Assertions with require.Assertions), this method is provided so you
// can call `suite.Assert().NoError()`.
func (ict *InteractiveTestContext) Assert() *assert.Assertions {
	if ict.Assertions == nil {
		ict.Assertions = assert.New(ict.T())
	}
	return ict.Assertions
}

// SendLine sends a line of text to the console
func (ict *InteractiveTestContext) SendLine(s string) (int, error) {
	return ict.console.SendLine(s)
}

// ExpectEOF waits for an EOF or an error to be emitted on the console
func (ict *InteractiveTestContext) ExpectEOF() (string, error) {
	return ict.console.ExpectEOF()
}

// ExpectString waits for a string of text to ebe written to the console
func (ict *InteractiveTestContext) ExpectString(s string) (string, error) {
	return ict.console.ExpectString(s)
}

// ExpectStringf waits for a string of formatted text to ebe written to the console
func (ict *InteractiveTestContext) ExpectStringf(format string, args ...interface{}) (string, error) {
	return ict.ExpectString(fmt.Sprintf(format, args...))
}

// ExpectMatch waits for a matcher to evaluate to true against content on the console
func (ict *InteractiveTestContext) ExpectMatch(opts ...expect.ExpectOpt) (string, error) {
	return ict.console.Expect(opts...)
}

// ExpectMatches waits for a series of matchers to evaluate to true against content on the console
func (ict *InteractiveTestContext) ExpectMatches(opts ...expect.ExpectOpt) (string, error) {
	return ict.console.Expect(opts...)
}

// RequireEOF waits for an EOF to be written to the console and terminates the test on timeout
func (ict *InteractiveTestContext) RequireEOF() (string, error) {
	l, err := ict.console.ExpectEOF()
	ict.Require().NoErrorf(err, "Unexpected error %q: - ", err, ict.context.OutputBuffer().String())
	return l, err
}

// RequireString waits for a string of text to be written to the console and terminates the test on timeout
func (ict *InteractiveTestContext) RequireString(s string) (string, error) {
	l, err := ict.console.ExpectString(s)
	ict.Require().NoErrorf(err, "Failed while attempting to read %q: %v", s, err)
	return l, err
}

// RequireStringf waits for a formatted string of text to be written to the console and terminates the test on timeout
func (ict *InteractiveTestContext) RequireStringf(format string, args ...interface{}) (string, error) {
	return ict.RequireString(fmt.Sprintf(format, args...))
}

// RequireMatch waits for a matcher to evaluate truthfully to be written to the console and terminates the test on timeout
func (ict *InteractiveTestContext) RequireMatch(opts ...expect.ExpectOpt) (string, error) {
	l, err := ict.console.Expect(opts...)
	ict.Require().NoErrorf(err, "Failed while attempting to find a matcher for %q: %v", l, err)
	return l, err
}

// RequireMatches waits for a series if matcher to evaluate truthfully to be written to the console and terminates the test on timeout
func (ict *InteractiveTestContext) RequireMatches(opts ...expect.ExpectOpt) (string, error) {
	l, err := ict.console.Expect(opts...)
	ict.Require().NoErrorf(err, "Failed while attempting to find a matcher for %q: %v", l, err)
	return l, err
}

// NewInteractiveCommandTester returns a new command executor for working with interactive terminal commands
func NewInteractiveCommandTester(command *cobra.Command, executor *InteractiveCommandExecutor, consoleOpts ...expect.ConsoleOpt) *InteractiveCommandTester {
	ice := executor
	if ice == nil {
		ice = NewInteractiveCommandExecutor(command)
	}
	return &InteractiveCommandTester{
		executor: ice,
	}
}

// InteractiveCommandTester instances are yielded while executing
// a command via ExecuteCommandInteractively
// and provide information about the virtual terminal execution environment
type InteractiveCommandTester struct {
	executor *InteractiveCommandExecutor
	t        *testing.T
}

// T returns the test instance for this command tester
func (ict *InteractiveCommandTester) T() *testing.T {
	if ict.t == nil {
		panic("invalid configuration: the test instance cannot be nil")
	}
	return ict.t
}

// Executor returns the command executor for this command tester
func (ict *InteractiveCommandTester) Executor() *InteractiveCommandExecutor {
	if ict.executor == nil {
		panic("invalid configuration: the executor instance cannot be nil")
	}
	return ict.executor
}

// Execute runs a Cobra command interactively and yields control to a func
// for testing the command through a console decorated with testify assertions
func (ict *InteractiveCommandTester) Execute(
	t *testing.T,
	args []string,
	testFunc func(*InteractiveTestContext) error) (*InteractiveExecutionContext, error) {
	return ict.Executor().Execute(args, func(context *InteractiveExecutionContext, console *expect.Console) error {
		return testFunc(&InteractiveTestContext{
			console: console,
			context: context,
			t:       t,
		})
	})
}

// ExecuteInInteractiveConsoleT runs a test within an interactive console environment
// Execution requires a standard test instance, a pair of functions that execute the code
// under test and the test code, and any desired options for configuring the virtual terminal environment
// The raw ouytput and the terminal state are logged in the event of a test failure
func ExecuteInInteractiveConsoleT(t *testing.T,
	codeUnderTestFunc InteractiveProcessFunc,
	testFunc InteractiveUserFunc,
	consoleOpts ...expect.ConsoleOpt) (*InteractiveExecutionContext, error) {
	context, err := ExecuteInInteractiveConsole(codeUnderTestFunc, testFunc, consoleOpts...)
	t.Logf("Raw output: %q", context.OutputBuffer().String())
	t.Logf("\n\nterminal state: %s", expect.StripTrailingEmptyLines(context.TerminalState().String()))
	return context, err
}
