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

package command_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InitTestSuite struct {
	suite.Suite
	*test.OpsaniCommandExecutor
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	viper.Reset()
	rootCmd := command.NewRootCommand()

	s.OpsaniCommandExecutor = test.NewOpsaniCommandExecutor(rootCmd)
}

func (s *InitTestSuite) TestRunningInitHelp() {
	output, err := s.Execute("init", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Initializes an Opsani config file")
}

func Stdio(c *expect.Console) terminal.Stdio {
	return terminal.Stdio{c.Tty(), c.Tty(), c.Tty()}
}

// type wantsStdio interface {
// 	WithStdio(terminal.Stdio)
// }

// type PromptTest struct {
// 	name      string
// 	prompt    survey.Prompt
// 	procedure func(*expect.Console)
// 	expected  interface{}
// }

// func RunPromptTest(t *testing.T, test PromptTest) {
// 	var answer interface{}
// 	RunTest(t, test.procedure, func(stdio terminal.Stdio) error {
// 		var err error
// 		if p, ok := test.prompt.(wantsStdio); ok {
// 			p.WithStdio(stdio)
// 		}

// 		answer, err = test.prompt.Prompt(defaultPromptConfig())
// 		return err
// 	})
// 	require.Equal(t, test.expected, answer)
// }

func RunTest(t *testing.T, procedure func(*expect.Console), test func(terminal.Stdio) error) {
	// t.Parallel()

	// Multiplex output to a buffer as well for the raw bytes.
	buf := new(bytes.Buffer)
	c, state, err := vt10x.NewVT10XConsole(expect.WithStdout(buf), expect.WithDefaultTimeout(time.Second))
	require.Nil(t, err)
	defer c.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		procedure(c)
	}()

	err = test(Stdio(c))
	require.Nil(t, err)

	// Close the slave end of the pty, and read the remaining bytes from the master end.
	c.Tty().Close()
	<-donec

	t.Logf("Raw output: %q", buf.String())

	// Dump the terminal's screen.
	t.Logf("\n\n\nterminal state: %s", expect.StripTrailingEmptyLines(state.String()))
}

func (s *InitTestSuite) TestTerminalInteraction() {
	var name string
	RunTest(s.T(), func(c *expect.Console) {
		c.ExpectString("? What is your name?")
		c.SendLine("Blake Watters")
		c.ExpectEOF()
	}, func(stdio terminal.Stdio) error {
		return survey.AskOne(&survey.Input{
			Message: "What is your name?",
		}, &name, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))
	})
	s.Require().Equal(name, "Blake Watters")
}

func (s *InitTestSuite) TestInitWithExistingConfig() {
	// core.DisableColor = true
	// configFile := test.TempConfigFileWithObj(map[string]string{
	// 	"app":   "example.com/app",
	// 	"token": "123456",
	// })
	// output, err := s.ExecuteWithConfig(configFile, "init")
	// s.Require().NoError(err)
	// s.Require().Contains(output, "Initializes an Opsani config file")

	// c, _ := expect.NewConsole()
	// question := survey.Question{
	// 	Name: "name",
	// 	Prompt: &survey.Input{
	// 		Message: "What is your name?",
	// 	},
	// }
}

// existing config
// confirmed
// set creds via args
// set creds via env
