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

package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	*OpsaniCommandExecutor
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupSuite() {
	s.OpsaniCommandExecutor = NewOpsaniCommandExecutor(rootCmd)
}

func (s *ConfigTestSuite) SetupTest() {
	viper.Reset()
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func (s *ConfigTestSuite) TestRunningConfigFileDoesntExist() {
	configFile := TempConfigFileWithBytes([]byte{})
	os.Remove(configFile.Name())

	_, err := s.ExecuteWithConfig(configFile, "config")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "no such file or directory")
}

func (s *ConfigTestSuite) TestRunningConfigFileEmpty() {
	configFile := TempConfigFileWithBytes([]byte{})
	defer os.Remove(configFile.Name())

	_, err := s.ExecuteWithConfig(configFile, "config")
	s.Require().EqualError(err, "command failed because client is not initialized. Run \"opsani init\" and try again")
}

func (s *ConfigTestSuite) TestRunningConfigWithInvalidFile() {
	configFile := TempConfigFileWithString("malformed:yaml:ysdsfsd")
	defer os.Remove(configFile.Name())
	_, err := s.ExecuteWithConfig(configFile, "config")
	s.Require().Error(err)
	s.Require().EqualError(err, "error parsing configuration file: While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `malform...` into map[string]interface {}")
}

func (s *ConfigTestSuite) TestRunningWithInitializedConfig() {
	configFile := TempConfigFileWithObj(map[string]interface{}{"app": "example.com/app1", "token": "123456"})
	defer os.Remove(configFile.Name())
	output, err := s.ExecuteWithConfig(configFile, "config")
	s.Require().NoError(err)
	s.Require().Contains(output, `"app": "example.com/app1"`)
	s.Require().Contains(output, `"token": "123456"`)
	s.Require().Contains(output, fmt.Sprintln("Using config from:", configFile.Name()))
}
