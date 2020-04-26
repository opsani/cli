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

type AuthTestSuite struct {
	test.Suite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func (s *AuthTestSuite) TestLoginHelp() {
	output, err := s.Execute("auth", "login", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Login to Opsani")
}

func (s *AuthTestSuite) TestLogoutHelp() {
	output, err := s.Execute("auth", "logout", "--help")
	s.Require().NoError(err)
	s.Require().Contains(output, "Logout from Opsani")
}
