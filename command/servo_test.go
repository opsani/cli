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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ServoTestSuite struct {
	test.Suite
}

func TestServoTestSuite(t *testing.T) {
	suite.Run(t, new(ServoTestSuite))
}

func (s *ServoTestSuite) SetupTest() {
	viper.Reset()
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
