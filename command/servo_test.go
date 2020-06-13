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

	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
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
	s.SetCommand(command.NewRootCommand())
}

func (s *ServoTestSuite) TestRunningServo() {
	output, err := s.Execute("servo")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage servos")
	s.Require().Contains(output, "Usage:")
}

func (s *ServoTestSuite) TestRunningServoHelp() {
	output, err := s.Execute("servo", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage servos")
}

func (s *ServoTestSuite) TestRunningServoInvalidPositionalArg() {
	output, err := s.Execute("servo", "--help", "sadasdsdas")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage servos")
}

func (s *ServoTestSuite) TestRunningServoSSHHelp() {
	output, err := s.Execute("servo", "shell", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Open an interactive shell on the servo")
}

func (s *ServoTestSuite) TestRunningServoSSHInvalidServo() {
	configFile := test.TempConfigFileWithObj(map[string][]map[string]string{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
			},
		},
	})
	_, err := s.Execute(test.Args("--config", configFile.Name(), "servo", "shell")...)
	s.Require().EqualError(err, "no driver for servo type: \"\"")
}

func (s *ServoTestSuite) TestRunningServoLogsHelp() {
	output, err := s.Execute("servo", "logs", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "View servo logs")
}

func (s *ServoTestSuite) TestRunningServoLogsInvalidServo() {
	configFile := test.TempConfigFileWithObj(map[string][]map[string]string{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
			},
		},
	})
	_, _, err := s.ExecuteC(test.Args("--config", configFile.Name(), "servo", "logs")...)
	s.Require().EqualError(err, "no driver for servo type: \"\"")
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
	output, err := s.Execute("servo", "attach", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Attach servo to the active profile")
}

func (s *ServoTestSuite) TestRunningAddNoInput() {
	configFile := test.TempConfigFileWithObj(map[string][]map[string]string{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "servo", "attach")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString("Select deployment:")
		t.SendLine("d")
		t.RequireString("User?")
		t.SendLine("blakewatters")
		t.RequireString("Host?")
		t.SendLine("dev.opsani.com")
		t.RequireString("Path? (optional)")
		t.SendLine("/servo")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check the config file
	body, _ := ioutil.ReadFile(configFile.Name())
	expected := `profiles:
  - name: default
    optimizer: example.com/app
    token: '123456'
    servo:
      type: docker-compose
      user: blakewatters
      host: dev.opsani.com
      path: /servo`
	s.Require().YAMLEq(expected, string(body))
}

func (s *ServoTestSuite) TestRunningAddNoInputWithBastion() {
	configFile := test.TempConfigFileWithObj(map[string][]map[string]string{
		"profiles": {
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "servo", "attach", "--bastion")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString("Select deployment:")
		t.SendLine("d")
		t.RequireString("User?")
		t.SendLine("blakewatters")
		t.RequireString("Host?")
		t.SendLine("dev.opsani.com")
		t.RequireString("Path? (optional)")
		t.SendLine("/servo")
		t.RequireString("Bastion host? (format is user@host[:port])")
		t.SendLine("blake@ssh.opsani.com:5555")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check the config file
	body, _ := ioutil.ReadFile(configFile.Name())
	expected := `profiles:
- name: default
  optimizer: example.com/app
  token: "123456"
  servo:
    type: docker-compose
    user: blakewatters
    host: dev.opsani.com
    path: /servo
    bastion: blake@ssh.opsani.com:5555`
	s.Require().YAMLEq(expected, string(body))
}

// TODO: Override port and specifying some values on CLI

func (s *ServoTestSuite) TestRunningRemoveHelp() {
	output, err := s.Execute("servo", "detach", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Detach servo from the active profile")
}

func (s *ServoTestSuite) TestRunningConfigHelp() {
	output, err := s.Execute("servo", "config", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "View servo config file")
}

func (s *ServoTestSuite) TestRunningRemoveServoConfirmed() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]interface{}{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"servo": map[string]string{
					"host": "dev.opsani.com",
					"name": "opsani-dev",
					"path": "/servo",
					"port": "",
					"user": "blakewatters",
				},
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "servo", "detach")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`Detach servo from profile "default"?`)
		t.SendLine("Y")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	var config = map[string][]command.Profile{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	s.Require().Empty(config["profiles"][0].Servo)
}

func (s *ServoTestSuite) TestRunningRemoveServoUnknown() {
	config := map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "servo", "detach", "unknown")
	s.Require().EqualError(err, "unknown command \"unknown\" for \"opsani servo detach\"")
}

func (s *ServoTestSuite) TestRunningRemoveServoForce() {
	config := map[string]interface{}{
		"profiles": []map[string]interface{}{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"servo": map[string]string{
					"host": "dev.opsani.com",
					"path": "/servo",
					"port": "",
					"user": "blakewatters",
				},
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "servo", "detach", "-f")
	s.Require().NoError(err)

	// Check that the servo has been removed
	var configState = map[string][]command.Profile{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &configState)
	s.Require().Empty(configState["profiles"][0].Servo)
}

func (s *ServoTestSuite) TestRunningRemoveServoDeclined() {
	configData := map[string]interface{}{
		"profiles": []map[string]interface{}{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"servo": map[string]string{
					"host": "dev.opsani.com",
					"path": "/servo",
					"port": "",
					"user": "blakewatters",
				},
			},
		},
	}
	configFile := test.TempConfigFileWithObj(configData)
	args := test.Args("--config", configFile.Name(), "servo", "detach")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`Detach servo from profile "default"?`)
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	body, _ := ioutil.ReadFile(configFile.Name())
	var configState = map[string][]command.Profile{}
	yaml.Unmarshal(body, &configState)
	s.Require().NotNil(configState["profiles"][0].Servo)
}

func (s *ServoTestSuite) TestRunningServoList() {
	config := map[string]interface{}{
		"profiles": []map[string]interface{}{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"servo": map[string]string{
					"host": "dev.opsani.com",
					"type": "docker-compose",
					"path": "/servo",
					"port": "",
					"user": "blakewatters",
				},
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "servo", "list")
	s.Require().NoError(err)
	s.Require().Contains(output, "default	docker-compose	ssh://blakewatters@dev.opsani.com:/servo")
}

func (s *ServoTestSuite) TestRunningServoListVerbose() {
	config := map[string]interface{}{
		"profiles": []map[string]interface{}{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"servo": map[string]string{
					"host": "dev.opsani.com",
					"type": "docker-compose",
					"path": "/servo",
					"port": "",
					"user": "blakewatters",
				},
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "servo", "list", "-v")
	s.Require().NoError(err)
	s.Require().Contains(output, "NAME   	TYPE          	NAMESPACE	DEPLOYMENT	USER        	HOST          	PATH   ")
	s.Require().Contains(output, "default	docker-compose	         	          	blakewatters	dev.opsani.com	/servo	")
}
