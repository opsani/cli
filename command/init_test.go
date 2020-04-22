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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/spf13/viper"
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
	// Colors make the tests flaky
	core.DisableColor = true
	viper.Reset()
	rootCmd := command.NewRootCommand()

	s.OpsaniCommandExecutor = test.NewOpsaniCommandExecutor(rootCmd)
}

func (s *InitTestSuite) TestRunningInitHelp() {
	output, err := s.Execute("init", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Initializes an Opsani config file")
}

func (s *InitTestSuite) TestTerminalInteraction() {
	var name string
	test.RunTestInInteractiveTerminal(s.T(), func(context *test.InteractiveExecutionContext) error {
		proxy := NewPassthroughPipeFile(context.GetStdin())
		return survey.AskOne(&survey.Input{
			Message: "What is your name?",
		}, &name, survey.WithStdio(proxy, context.GetStdout(), context.GetStderr()))
	}, func(_ *test.InteractiveExecutionContext, c *expect.Console) error {
		s.RequireNoErr2(c.ExpectString("What is your name?"))
		c.SendLine("Blake Watters")
		c.ExpectEOF()
		return nil
	})
	s.Require().Equal(name, "Blake Watters")
}

func (s *InitTestSuite) RequireNoErr2(_ interface{}, err error) {
	s.Require().NoError(err)
}

type PassthroughPipeFile struct {
	*expect.PassthroughPipe
	file *os.File
}

func (s *PassthroughPipeFile) Fd() uintptr {
	return s.file.Fd()
}

func NewPassthroughPipeFile(stdin *os.File) *PassthroughPipeFile {
	file := os.NewFile(stdin.Fd(), "pipe")
	pipe, _ := expect.NewPassthroughPipe(stdin)
	pipe.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	return &PassthroughPipeFile{
		file:            file,
		PassthroughPipe: pipe,
	}
}

func (s *InitTestSuite) TestTerminalConfirm() {
	var confirmed bool = true
	test.RunTestInInteractiveTerminal(s.T(), func(context *test.InteractiveExecutionContext) error {
		return survey.AskOne(&survey.Confirm{
			Message: "Delete file?",
		}, &confirmed, survey.WithStdio(NewPassthroughPipeFile(context.GetStdin()), context.GetStdout(), context.GetStderr()))
	}, func(_ *test.InteractiveExecutionContext, c *expect.Console) error {
		s.RequireNoErr2(c.Expect(expect.RegexpPattern("Delete file?")))
		c.SendLine("N")
		c.ExpectEOF()
		return nil
	})
	s.Require().False(confirmed)
}

func (s *InitTestSuite) TestInitWithExistingConfig() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})

	rootCmd := command.NewRootCommand()
	ice := test.NewInteractiveCommandExecutor(rootCmd, expect.WithDefaultTimeout(1.0*time.Second))
	ice.PreExecutionFunc = func(context *test.InteractiveExecutionContext) error {
		// Attach the survey library to the console
		// This is necessary because of type safety fun with modeling around file readers
		command.Stdio = terminal.Stdio{NewPassthroughPipeFile(context.GetStdin()), context.GetStdout(), context.GetStderr()}
		return nil
	}
	_, err := ice.Execute(test.Args("--config", configFile.Name(), "init"), func(_ *test.InteractiveExecutionContext, console *expect.Console) error {
		if _, err := console.ExpectString(fmt.Sprintf("Using config from: %s", configFile.Name())); err != nil {
			return err
		}
		str := fmt.Sprintf("? Existing config found. Overwrite %s?", configFile.Name())
		_, err := console.ExpectString(str)
		s.NoError(err)
		_, err = console.SendLine("N")
		s.NoError(err)
		_, err = console.ExpectEOF()
		return nil
	})
	s.Require().Error(err)
	s.Require().EqualError(err, terminal.InterruptErr.Error())
}
