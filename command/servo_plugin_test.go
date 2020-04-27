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
	"testing"

	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/stretchr/testify/suite"
)

type ServoPluginTestSuite struct {
	test.Suite
}

func TestServoPluginTestSuite(t *testing.T) {
	suite.Run(t, new(ServoPluginTestSuite))
}

func (s *ServoPluginTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *ServoPluginTestSuite) TestRootHelp() {
	output, err := s.Execute("servo", "plugin", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage Servo plugins")
}

func (s *ServoPluginTestSuite) TestListtHelp() {
	output, err := s.Execute("servo", "plugin", "list", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "List Servo plugins")
}

func (s *ServoPluginTestSuite) TestSearchHelp() {
	output, err := s.Execute("servo", "plugin", "search", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Search for Servo plugins")
}

func (s *ServoPluginTestSuite) TestInfoHelp() {
	output, err := s.Execute("servo", "plugin", "info", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Get info about a Servo plugin")
}

func (s *ServoPluginTestSuite) TestCloneHelp() {
	output, err := s.Execute("servo", "plugin", "clone", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Clone a Servo plugin with Git")
}

func (s *ServoPluginTestSuite) TestForkHelp() {
	output, err := s.Execute("servo", "plugin", "fork", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Fork a Servo plugin on GitHub")
}

func (s *ServoPluginTestSuite) TestCreateHelp() {
	output, err := s.Execute("servo", "plugin", "create", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Create a new Servo plugin")
}
