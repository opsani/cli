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
	"io/ioutil"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

type InitTestSuite struct {
	suite.Suite
	*test.OpsaniCommandExecutor
	*test.InteractiveCommandExecutor
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
	s.InteractiveCommandExecutor = test.NewInteractiveCommandExecutor(rootCmd)
}

func (s *InitTestSuite) TestRunningInitHelp() {
	output, err := s.Execute("init", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Initializes an Opsani config file")
}

func (s *InitTestSuite) TestTerminalInteraction() {
	var name string
	test.RunTestInInteractiveTerminal(s.T(), func(context *test.InteractiveExecutionContext) error {
		fmt.Printf("%v\n", context)
		return survey.AskOne(&survey.Input{
			Message: "What is your name?",
		}, &name, survey.WithStdio(context.Tty(), context.Tty(), context.Tty()))
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

func (s *InitTestSuite) TestTerminalConfirm() {
	var confirmed bool = true
	test.RunTestInInteractiveTerminal(s.T(), func(context *test.InteractiveExecutionContext) error {
		return survey.AskOne(&survey.Confirm{
			Message: "Delete file?",
		}, &confirmed, survey.WithStdio(context.Tty(), context.Tty(), context.Tty()))
	}, func(_ *test.InteractiveExecutionContext, c *expect.Console) error {
		s.RequireNoErr2(c.Expect(expect.RegexpPattern("Delete file?")))
		c.SendLine("N")
		c.ExpectEOF()
		return nil
	})
	s.Require().False(confirmed)
}

type InteractiveCommandTest struct {
	console *expect.Console
	context *test.InteractiveExecutionContext
	s       *InitTestSuite
}

func (ict *InteractiveCommandTest) Require() *require.Assertions {
	return ict.s.Require()
}

func (ict *InteractiveCommandTest) SendLine(s string) (int, error) {
	l, err := ict.console.SendLine(s)
	ict.Require().NoError(err)
	return l, err
}

func (ict *InteractiveCommandTest) ExpectEOF() (string, error) {
	return ict.console.ExpectEOF()
}

func (ict *InteractiveCommandTest) ExpectString(s string) (string, error) {
	return ict.console.ExpectString(s)
}

func (ict *InteractiveCommandTest) ExpectStringf(format string, args ...interface{}) (string, error) {
	return ict.ExpectString(fmt.Sprintf(format, args...))
}

func (ict *InteractiveCommandTest) ExpectMatch(opts ...expect.ExpectOpt) (string, error) {
	return ict.console.Expect(opts...)
}

func (ict *InteractiveCommandTest) ExpectMatches(opts ...expect.ExpectOpt) (string, error) {
	return ict.console.Expect(opts...)
}

func (ict *InteractiveCommandTest) RequireEOF() (string, error) {
	l, err := ict.console.ExpectEOF()
	ict.Require().NoErrorf(err, "Unexpected error %q: - ", err, ict.context.OutputBuffer().String())
	return l, err
}

func (ict *InteractiveCommandTest) RequireString(s string) (string, error) {
	l, err := ict.console.ExpectString(s)
	ict.Require().NoErrorf(err, "Failed while attempting to read %q: %v", s, err)
	return l, err
}

func (ict *InteractiveCommandTest) RequireStringf(format string, args ...interface{}) (string, error) {
	return ict.RequireString(fmt.Sprintf(format, args...))
}

func (ict *InteractiveCommandTest) RequireMatch(opts ...expect.ExpectOpt) (string, error) {
	l, err := ict.console.Expect(opts...)
	ict.Require().NoErrorf(err, "Failed while attempting to find a matcher for %q: %v", l, err)
	return l, err
}

func (ict *InteractiveCommandTest) RequireMatches(opts ...expect.ExpectOpt) (string, error) {
	l, err := ict.console.Expect(opts...)
	ict.Require().NoErrorf(err, "Failed while attempting to find a matcher for %q: %v", l, err)
	return l, err
}

func (s *InitTestSuite) ExecuteCommandInteractivelyE(
	ice *test.InteractiveCommandExecutor,
	args []string,
	testFunc func(*InteractiveCommandTest) error) (*test.InteractiveExecutionContext, error) {
	return ice.ExecuteInteractively(args, func(context *test.InteractiveExecutionContext, console *expect.Console) error {
		return testFunc(&InteractiveCommandTest{
			console: console,
			context: context,
			s:       s,
		})
	})
}

func (s *InitTestSuite) ExecuteCommandInteractively(
	args []string,
	testFunc func(*InteractiveCommandTest) error) (*test.InteractiveExecutionContext, error) {
	return s.ExecuteInteractively(args, func(context *test.InteractiveExecutionContext, console *expect.Console) error {
		return testFunc(&InteractiveCommandTest{
			console: console,
			context: context,
			s:       s,
		})
	})
}

func (s *InitTestSuite) TestInitWithExistingConfigDeclinedL() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})

	context, err := s.ExecuteCommandInteractively(test.Args("--config", configFile.Name(), "init"), func(t *InteractiveCommandTest) error {
		fmt.Printf("? Console = %v, Context = %v, t = %v, require = %v", t.console, t.context, t.s, t.s)
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("? Existing config found. Overwrite %s?", configFile.Name())
		t.SendLine("N")
		t.console.ExpectEOF()
		return nil
	})
	s.T().Logf("%v", context.OutputBuffer().String())
	s.Require().Error(err)
	s.Require().EqualError(err, terminal.InterruptErr.Error())
}

func (s *InitTestSuite) TestInitWithExistingConfigDeclined() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})

	context, err := s.ExecuteCommandInteractively(test.Args("--config", configFile.Name(), "init"), func(t *InteractiveCommandTest) error {
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("? Existing config found. Overwrite %s?", configFile.Name())
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.T().Logf("%v", context.OutputBuffer().String())
	s.Require().Error(err)
	s.Require().EqualError(err, terminal.InterruptErr.Error())
}

func (s *InitTestSuite) TestInitWithExistingConfigAccepted() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})

	context, err := s.ExecuteCommandInteractively(test.Args("--config", configFile.Name(), "init"), func(t *InteractiveCommandTest) error {
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("? Existing config found. Overwrite %s?", configFile.Name())
		t.SendLine("Y")
		t.ExpectMatch(expect.RegexpPattern("Opsani app"))
		t.SendLine("dev.opsani.com/amazing-app")
		t.RequireMatch(expect.RegexpPattern("API Token"))
		t.SendLine("123456")
		t.RequireMatch(expect.RegexpPattern(fmt.Sprintf("Write to %s?", configFile.Name())))

		t.SendLine("Y")
		t.RequireMatch(expect.RegexpPattern("Opsani CLI initialized"))
		return nil
	})
	s.Require().NoError(err, context.OutputBuffer().String())

	// Check the config file
	var config = map[string]interface{}{}
	body, err := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	s.Require().Equal("dev.opsani.com/amazing-app", config["app"])
	s.Require().Equal("123456", config["token"])
}
