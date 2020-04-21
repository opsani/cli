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

type AppLifecycleTestSuite struct {
	suite.Suite
	*test.OpsaniCommandExecutor
}

func TestAppLifecycleTestSuite(t *testing.T) {
	suite.Run(t, new(AppLifecycleTestSuite))
}

func (s *AppLifecycleTestSuite) SetupTest() {
	viper.Reset()
	rootCmd := command.NewRootCommand()

	s.OpsaniCommandExecutor = test.NewOpsaniCommandExecutor(rootCmd)
}

func (s *AppLifecycleTestSuite) TestRunningAppStartHelp() {
	output, err := s.Execute("app", "start", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Start the app")
}

func (s *AppLifecycleTestSuite) TestRunningAppStopHelp() {
	output, err := s.Execute("app", "stop", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Stop the app")
}

func (s *AppLifecycleTestSuite) TestRunningAppRestartHelp() {
	output, err := s.Execute("app", "restart", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Restart the app")
}

func (s *AppLifecycleTestSuite) TestRunningAppStatusHelp() {
	output, err := s.Execute("app", "status", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Check app status")
}
