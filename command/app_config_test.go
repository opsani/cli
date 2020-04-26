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

type AppConfigTestSuite struct {
	test.Suite
}

func TestAppConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AppConfigTestSuite))
}

func (s *AppConfigTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *AppConfigTestSuite) TestRunningAppConfigEditHelp() {
	output, err := s.Execute("app", "config", "edit", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Edit app config")
}

func (s *AppConfigTestSuite) TestRunningAppConfigGetHelp() {
	output, err := s.Execute("app", "config", "get", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Get app config")
}

func (s *AppConfigTestSuite) TestRunningAppConfigPatchHelp() {
	output, err := s.Execute("app", "config", "patch", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Patch merges the incoming")
}

func (s *AppConfigTestSuite) TestRunningAppConfigSetHelp() {
	output, err := s.Execute("app", "config", "set", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Set app config")
}
