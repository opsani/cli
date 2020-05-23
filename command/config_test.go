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
	"os"
	"regexp"
	"testing"

	"github.com/opsani/cli/command"
	"github.com/opsani/cli/test"
	"github.com/stretchr/testify/suite"
)

func ConfigFileArgs(file *os.File, args ...string) []string {
	return test.Args(append([]string{"--config", file.Name()}, args...)...)
}

type ConfigTestSuite struct {
	test.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.SetCommand(command.NewRootCommand())
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func (s *ConfigTestSuite) TestRunningConfigFileDoesntExist() {
	configFile := test.TempConfigFileWithBytes([]byte{})
	os.Remove(configFile.Name())

	_, err := s.ExecuteArgs(ConfigFileArgs(configFile, "config"))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "no such file or directory")
}

func (s *ConfigTestSuite) TestRunningConfigFileEmpty() {
	configFile := test.TempConfigFileWithBytes([]byte{})
	_, err := s.ExecuteArgs(ConfigFileArgs(configFile, "config"))
	s.Require().EqualError(err, "command failed because client is not initialized. Run \"opsani init\" and try again")
}

func (s *ConfigTestSuite) TestRunningConfigWithInvalidFile() {
	configFile := test.TempConfigFileWithString("malformed:yaml:ysdsfsd")
	_, err := s.ExecuteArgs(ConfigFileArgs(configFile, "config"))
	s.Require().Error(err)
	s.Require().EqualError(err, "error parsing configuration file: While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `malform...` into map[string]interface {}")
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func Strip(str string) string {
	return re.ReplaceAllString(str, "")
}

func (s *ConfigTestSuite) TestRunningWithInitializedConfig() {
	configFile := test.TempConfigFileWithObj(map[string]interface{}{"profiles": []map[string]string{{"app": "example.com/app1", "token": "123456"}}})
	output, err := s.ExecuteArgs(ConfigFileArgs(configFile, "config"))
	s.Require().NoError(err)
	yaml := Strip(output)
	s.Require().Contains(yaml, `app: example.com/app1`)
	s.Require().Contains(yaml, `token: "123456`)
	s.Require().Contains(yaml, fmt.Sprintln("Using config from:", configFile.Name()))
}
