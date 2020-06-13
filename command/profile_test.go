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

type ProfileTestSuite struct {
	test.Suite
}

func TestProfileTestSuite(t *testing.T) {
	suite.Run(t, new(ProfileTestSuite))
}

func (s *ProfileTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *ProfileTestSuite) TestRunningProfile() {
	output, err := s.Execute("profile")
	s.Require().NoError(err)
	s.Require().Contains(output, "Profiles provide an interface for ")
	s.Require().Contains(output, "Usage:")
}

func (s *ProfileTestSuite) TestRunningProfileHelp() {
	output, err := s.Execute("profile", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Profiles provide an interface for interacting with an optimizer backend and a servo deployment as a unit")
}

func (s *ProfileTestSuite) TestRunningProfileInvalidPositionalArg() {
	output, err := s.Execute("profile", "--help", "sadasdsdas")
	s.Require().NoError(err)
	s.Require().Contains(output, "Profiles provide an interface for interacting with an optimizer backend and a servo deployment as a unit")
}

func (s *ProfileTestSuite) TestRunningAddHelp() {
	output, err := s.Execute("profile", "add", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Add a profile to the configuration")
}

func (s *ProfileTestSuite) TestRunningAddNoInput() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "profile", "add")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString("Profile name?")
		t.SendLine("opsani-dev")
		t.RequireString("Opsani optimizer (e.g. domain.com/app)?")
		t.SendLine("dev.opsani.com/blake")
		t.RequireString("API Token?")
		t.SendLine("123456")
		t.RequireString("Attach servo to new profile?")
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check the config file
	var config = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	expected := []interface{}(
		[]interface{}{
			map[interface{}]interface{}{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
			map[interface{}]interface{}{
				"optimizer": "dev.opsani.com/blake",
				"name":      "opsani-dev",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	)
	s.Require().EqualValues(expected, config["profiles"])
}

func (s *ProfileTestSuite) TestRunningRemoveHelp() {
	output, err := s.Execute("profile", "remove", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Remove a profile from the configuration")
}

func (s *ProfileTestSuite) TestRunningRemoveProfileConfirmed() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "profile", "remove", "default")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`Remove profile "default"?`)
		t.SendLine("Y")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check the config file
	var config = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &config)
	s.Require().EqualValues([]interface{}{}, config["profiles"])
}

func (s *ProfileTestSuite) TestRunningRemoveProfileUnknown() {
	config := map[string]interface{}{
		"profiles": []map[string]string{
			{
				"optimizer": "dev.opsani.com",
				"name":      "opsani-dev",
				"token":     "/profile",
				"base_url":  "https://api.opsani.com/",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "profile", "remove", "unknown")
	s.Require().EqualError(err, `Unable to find profile "unknown"`)
}

func (s *ProfileTestSuite) TestRunningRemoveProfileForce() {
	config := map[string]interface{}{
		"profiles": []map[string]string{
			{
				"optimizer": "dev.opsani.com",
				"name":      "opsani-dev",
				"token":     "/profile",
				"base_url":  "https://api.opsani.com/",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	_, err := s.Execute("--config", configFile.Name(), "profile", "remove", "-f", "opsani-dev")
	s.Require().NoError(err)

	// Check that the profile has been removed
	var configState = map[string]interface{}{}
	body, _ := ioutil.ReadFile(configFile.Name())
	yaml.Unmarshal(body, &configState)
	s.Require().EqualValues([]interface{}{}, configState["profiles"])
}

func (s *ProfileTestSuite) TestRunningRemoveProfileDeclined() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	})
	args := test.Args("--config", configFile.Name(), "profile", "remove", "default")
	_, err := s.ExecuteTestInteractively(args, func(t *test.InteractiveTestContext) error {
		t.RequireString(`Remove profile "default"?`)
		t.SendLine("N")
		t.ExpectEOF()
		return nil
	})
	s.Require().NoError(err)

	// Check that the config file has not changed
	expected := []interface{}(
		[]interface{}{
			map[interface{}]interface{}{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	)
	body, _ := ioutil.ReadFile(configFile.Name())
	var configState = map[string]interface{}{}
	yaml.Unmarshal(body, &configState)
	s.Require().EqualValues(expected, configState["profiles"])
}

func (s *ProfileTestSuite) TestRunningProfileList() {
	config := map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "profile", "list")
	s.Require().NoError(err)
	s.Require().Contains(output, "default	example.com/app	123456")
}

func (s *ProfileTestSuite) TestRunningProfileListVerbose() {
	config := map[string]interface{}{
		"profiles": []map[string]string{
			{
				"name":      "default",
				"optimizer": "example.com/app",
				"token":     "123456",
				"base_url":  "https://api.opsani.com/",
			},
		},
	}
	configFile := test.TempConfigFileWithObj(config)
	output, err := s.Execute("--config", configFile.Name(), "profile", "list", "-v")
	s.Require().NoError(err)
	s.Require().Contains(output, "NAME   	APP            	TOKEN 	SERVO ")
	s.Require().Contains(output, "default	example.com/app	123456")
}
