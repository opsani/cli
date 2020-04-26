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
	"io/ioutil"
	"testing"

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

type ServoTestSuite struct {
	test.Suite
}

func TestServoTestSuite(t *testing.T) {
	suite.Run(t, new(ServoTestSuite))
}

func (s *ServoTestSuite) SetupTest() {
	viper.Reset()
	core.DisableColor = true
	s.SetCommand(command.NewRootCommand())
}

func (s *ServoTestSuite) TestRunningServo() {
	output, err := s.Execute("servo")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage Servos")
	s.Require().Contains(output, "Usage:")
}

func (s *ServoTestSuite) TestRunningServoHelp() {
	output, err := s.Execute("servo", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage Servos")
}

func (s *ServoTestSuite) TestRunningServoInvalidPositionalArg() {
	output, err := s.Execute("servo", "--help", "sadasdsdas")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage Servos")
}

func (s *ServoTestSuite) TestRunningServoSSHHelp() {
	output, err := s.Execute("servo", "ssh", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "SSH into a Servo")
}

func (s *ServoTestSuite) TestRunningServoSSHInvalidServo() {
	_, err := s.Execute("servo", "ssh", "fake-name")
	s.Require().EqualError(err, `no such Servo "fake-name"`)
}

func (s *ServoTestSuite) TestRunningServoLogsHelp() {
	output, err := s.Execute("servo", "logs", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "View logs on a Servo")
}

func (s *ServoTestSuite) TestRunningServoLogsInvalidServo() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})
	_, _, err := s.ExecuteC(test.Args("--config", configFile.Name(), "servo", "logs", "fake-name")...)
	s.Require().EqualError(err, `no such Servo "fake-name"`)
}

func (s *ServoTestSuite) TestRunningServoFollowHelp() {
	output, err := s.Execute("servo", "logs", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Follow log output")
}

func (s *ServoTestSuite) TestRunningLogsTimestampsHelp() {
	output, err := s.Execute("servo", "logs", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Show timestamps")
}

func (s *ServoTestSuite) TestRunningAddHelp() {
	output, err := s.Execute("servo", "add", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Add a Servo")
}

func (s *ServoTestSuite) TestRunningAddNoInput() {
	configFile := test.TempConfigFileWithObj(map[string]string{
		"app":   "example.com/app",
		"token": "123456",
	})
	args := test.Args("--config", configFile.Name(), "servo", "add")
	context, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString("? Servo name?")
		t.SendLine("opsani-dev")
		t.RequireString("? User?")
		t.SendLine("blakewatters")
		t.RequireString("? Host?")
		t.SendLine("dev.opsani.com")
		t.RequireString("? Path? (optional)")
		t.SendLine("/servo")
		t.ExpectEOF()
		return nil
	})
	s.T().Logf("The output buffer is: %v", context.OutputBuffer().String())
	s.Require().NoError(err)

	// Check the config file
	var config = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	expected := []interface{}(
		[]interface{}{
			map[interface{}]interface{}{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	)
	s.Require().EqualValues(expected, config["servos"])
}

// TODO: Override port and specifying some values on CLI

func (s *ServoTestSuite) TestRunningRemoveHelp() {
	output, err := s.Execute("servo", "remove", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Remove a Servo")
}

// TODO: add -f
func (s *ServoTestSuite) TestRunningRemoveServoConfirmed() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
		"servos": []map[string]string{
			{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "servo", "remove", "opsani-dev")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`? Remove Servo "opsani-dev"?`)
		t.SendLine("Y")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check the config file
	var config = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	s.Require().EqualValues([]interface{}{}, config["servos"])
}

func (s *ServoTestSuite) TestRunningRemoveServoUnknown() {
	config := map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "servo", "remove", "unknown")
	s.Require().EqualError(err, `Unable to find Servo named "unknown"`)
}

func (s *ServoTestSuite) TestRunningRemoveServoForce() {
	config := map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
		"servos": []map[string]string{
			{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "servo", "remove", "-f", "opsani-dev")
	s.Require().NoError(err)

	// Check that the servo has been removed
	var configState = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &configState)
	s.Require().EqualValues([]interface{}{}, configState["servos"])
}

func (s *ServoTestSuite) TestRunningRemoveServoDeclined() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
		"servos": []map[string]string{
			{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "servo", "remove", "opsani-dev")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`? Remove Servo "opsani-dev"?`)
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check that the config file has not changed
	expected := []interface{}(
		[]interface{}{
			map[interface{}]interface{}{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	)
	body, _ := ioutil.ReadFile(configFile.Name())
	var configState = map[string]interface{}{}
	yaml.Unmarshal(body, &configState)
	s.Require().EqualValues(expected, configState["servos"])
}

func (s *ServoTestSuite) TestRunningServoList() {
	config := map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
		"servos": []map[string]string{
			{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "servo", "list")
	s.Require().NoError(err)
	s.Require().Contains(output, "opsani-dev	ssh://blakewatters@dev.opsani.com:/servo	")
}

func (s *ServoTestSuite) TestRunningServoListVerbose() {
	config := map[string]interface{}{
		"app":   "example.com/app",
		"token": "123456",
		"servos": []map[string]string{
			{
				"host": "dev.opsani.com",
				"name": "opsani-dev",
				"path": "/servo",
				"port": "",
				"user": "blakewatters",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "servo", "list", "-v")
	s.Require().NoError(err)
	s.Require().Contains(output, "NAME      	USER        	HOST          	PATH  ")
	s.Require().Contains(output, "opsani-dev	blakewatters	dev.opsani.com	/servo	")
}
