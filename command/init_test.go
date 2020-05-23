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
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

type InitTestSuite struct {
	test.Suite
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *InitTestSuite) TestRunningInitHelp() {
	output, err := s.Execute("init", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Initializes an Opsani config file")
}

// NOTE: This test intentionally uses Tty() instead of PassthroughTty()
// In the event of breakage the test may block. Use a go test timeoout or debug on another test case
func (s *InitTestSuite) TestTerminalInteraction() {
	var name string
	test.ExecuteInInteractiveConsoleT(s.T(), func(context *test.InteractiveExecutionContext) error {
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
	test.ExecuteInInteractiveConsoleT(s.T(), func(context *test.InteractiveExecutionContext) error {
		return survey.AskOne(&survey.Confirm{
			Message: "Delete file?",
		}, &confirmed, survey.WithStdio(context.PassthroughTty(), context.PassthroughTty(), context.PassthroughTty()))
	}, func(_ *test.InteractiveExecutionContext, c *expect.Console) error {
		s.RequireNoErr2(c.Expect(expect.RegexpPattern("Delete file?")))
		c.SendLine("N")
		c.ExpectEOF()
		return nil
	})
	s.Require().False(confirmed)
}

func (s *InitTestSuite) TestInitWithExistingConfigDeclinedL() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"app":   "example.com/app",
				"token": "123456",
			},
		},
	})

	context, err := s.ExecuteTestInteractively(test.Args("--config", configFile.Name(), "init"), func(t *test.InteractiveTestContext) error {
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("Existing config found. Overwrite %s?", configFile.Name())
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.T().Logf("The output buffer is: %v", context.OutputBuffer().String())
	s.Require().Error(err)
	s.Require().EqualError(err, terminal.InterruptErr.Error())
}

func (s *InitTestSuite) TestInitWithExistingConfigDeclined() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"app":   "example.com/app",
				"token": "123456",
			},
		},
	})

	context, err := s.ExecuteTestInteractively(test.Args("--config", configFile.Name(), "init"), func(t *test.InteractiveTestContext) error {
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("Existing config found. Overwrite %s?", configFile.Name())
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.T().Logf("Output buffer = %v", context.OutputBuffer().String())
	s.Require().Error(err)
	s.Require().EqualError(err, terminal.InterruptErr.Error())
}

// TODO: There is a missing test case with initializing against the default file (no --config)

func (s *InitTestSuite) TestInitWithExistingConfigAccepted() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":  "default",
				"app":   "example.com/app",
				"token": "123456",
			},
		},
	})

	context, err := s.ExecuteTestInteractively(test.Args("--config", configFile.Name(), "init"), func(t *test.InteractiveTestContext) error {
		t.RequireStringf("Using config from: %s", configFile.Name())
		t.RequireStringf("Existing config found. Overwrite %s?", configFile.Name())
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
	s.T().Logf("Output buffer = %v", context.OutputBuffer().String())

	// Check the config file
	var config struct {
		Profiles []command.Profile `yaml:"profiles"`
	}
	body, err := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	s.Require().Equal("dev.opsani.com/amazing-app", config.Profiles[1].App)
	s.Require().Equal("123456", config.Profiles[1].Token)
}
