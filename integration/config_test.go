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

package integration

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestRunningConfigFileDoesntExist() {
	cmd := exec.Command(opsaniBinaryPath,
		"--config", opsaniConfigPath,
		"config",
	)

	output, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Require().Contains(string(output), "no such file or directory")
}

func (s *ConfigTestSuite) TestRunningConfigUninitialized() {
	cmd := exec.Command(opsaniBinaryPath,
		"--config", opsaniConfigPath,
		"config",
	)

	WriteConfigFile(nil)
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Require().Contains(string(output), "command failed because client is not initialized")
}

func (s *ConfigTestSuite) TestRunningConfigInitialized() {
	WriteConfigFile(defaultConfig)
	cmd := exec.Command(opsaniBinaryPath,
		"--config", opsaniConfigPath,
		"--no-colors",
		"config",
	)

	output, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Require().Contains(string(output), `app: example.com/app1`)
	s.Require().Contains(string(output), `token: "123456`)
}

func (s *ConfigTestSuite) TestRunningConfigFileInvalidData() {
	ioutil.WriteFile(opsaniConfigPath, []byte("this\nwill\nnot\nparse"), 0644)
	cmd := exec.Command(opsaniBinaryPath,
		"--config", opsaniConfigPath,
		"config",
	)

	output, err := cmd.CombinedOutput()
	s.Require().NoError(err)
	s.Require().Contains(string(output), "error parsing configuration file")
}
