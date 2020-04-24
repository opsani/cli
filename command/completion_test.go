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

type CompletionTestSuite struct {
	test.OpsaniTestSuite
}

func TestCompletionTestSuite(t *testing.T) {
	suite.Run(t, new(CompletionTestSuite))
}

func (s *CompletionTestSuite) SetupTest() {
	viper.Reset()
	s.OpsaniTestSuite.SetRootCommand(command.NewRootCommand())
}

func (s *CompletionTestSuite) TestRunningCompletionHelp() {
	output, err := s.ExecuteCommand("completion", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Generate shell completion scripts for Opsani CL")
}

func (s *CompletionTestSuite) TestRunningCompletionBash() {
	output, err := s.ExecuteCommand("completion", "--shell", "bash")
	s.Require().NoError(err)
	s.Require().Contains(output, "BASH_COMP_DEBUG_FILE")
}

func (s *CompletionTestSuite) TestRunningCompletionZsh() {
	output, err := s.ExecuteCommand("completion", "--shell", "zsh")
	s.Require().NoError(err)
	s.Require().Contains(output, "#compdef _opsani opsani")
}

func (s *CompletionTestSuite) TestRunningCompletionFish() {
	output, err := s.ExecuteCommand("completion", "--shell", "fish")
	s.Require().NoError(err)
	s.Require().Contains(output, "__fish_opsani_no_subcommand")
}

func (s *CompletionTestSuite) TestRunningCompletionPowershell() {
	output, err := s.ExecuteCommand("completion", "--shell", "powershell")
	s.Require().NoError(err)
	s.Require().Contains(output, "Register-ArgumentCompleter -Native -CommandName 'opsani'")
}
