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
	"testing"

	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/stretchr/testify/suite"
)

type IgniteTestSuite struct {
	test.Suite
}

func (s *IgniteTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func TestIgniteTestSuite(t *testing.T) {
	suite.Run(t, new(IgniteTestSuite))
}

func (s *IgniteTestSuite) TestRunningIgniteHelp() {
	output, err := s.Execute("ignite", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Light up an interactive demo")
}

func (s *IgniteTestSuite) TestRunningIgniteNoConfig() {
	output, err := s.Execute("ignite")
	fmt.Println(output)
	s.Require().EqualError(err, "command failed because client is not initialized. Run \"opsani init\" and try again")
}

func (s *IgniteTestSuite) TestRunningIgniteEmptyConfig() {
	configFile := test.TempConfigFileWithBytes([]byte{})
	_, err := s.ExecuteArgs(ConfigFileArgs(configFile, "ignite"))
	s.Require().EqualError(err, "command failed because client is not initialized. Run \"opsani init\" and try again")
}

func (s *IgniteTestSuite) TestRunningIgniteBadConfigExt() {
	output, err := s.Execute("ignite", "--config", "foo.ini")
	fmt.Println(output)
	s.Require().EqualError(err, "config file does not exist. Run \"opsani init\" and try again (stat foo.ini: no such file or directory)")
}
