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

type ServoAssemblyTestSuite struct {
	test.Suite
}

func TestServoAssemblyTestSuite(t *testing.T) {
	suite.Run(t, new(ServoAssemblyTestSuite))
}

func (s *ServoAssemblyTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *ServoAssemblyTestSuite) TestRootHelp() {
	output, err := s.Execute("servo", "assembly", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Manage Servo assemblies")
}

func (s *ServoAssemblyTestSuite) TestListtHelp() {
	output, err := s.Execute("servo", "assembly", "list", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "List Servo assemblies")
}

func (s *ServoAssemblyTestSuite) TestSearchHelp() {
	output, err := s.Execute("servo", "assembly", "search", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Search for Servo assemblies")
}

func (s *ServoAssemblyTestSuite) TestInfoHelp() {
	output, err := s.Execute("servo", "assembly", "info", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Get info about a Servo assembly")
}

func (s *ServoAssemblyTestSuite) TestCloneHelp() {
	output, err := s.Execute("servo", "assembly", "clone", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Clone a Servo assembly with Git")
}

func (s *ServoAssemblyTestSuite) TestForkHelp() {
	output, err := s.Execute("servo", "assembly", "fork", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Fork a Servo assembly on GitHub")
}

func (s *ServoAssemblyTestSuite) TestCreateHelp() {
	output, err := s.Execute("servo", "assembly", "create", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Create a new Servo assembly")
}
