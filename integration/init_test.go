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
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/opsani/cli/command"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const packagePath string = "github.com/opsani/cli"

var defaultConfig = struct {
	Profiles []command.Profile `yaml:"profiles"`
}{
	[]command.Profile{
		{
			Name:  "default",
			App:   "example.com/app1",
			Token: "123456",
		},
	},
}

var (
	opsaniBinaryPath string
	opsaniConfigPath string
	expect           *require.Assertions
	server           *httptest.Server
)

func TestMain(m *testing.M) {
	tmpDir, err := ioutil.TempDir("", "integration-opsani")
	if err != nil {
		panic("failed to create temp dir")
	}
	defer os.RemoveAll(tmpDir)

	opsaniBinaryPath = filepath.Join(tmpDir, path.Base(packagePath))
	if runtime.GOOS == "windows" {
		opsaniBinaryPath += ".exe"
	}
	opsaniConfigPath = filepath.Join(tmpDir, "config.yaml")

	cmd := exec.Command("go", "build", "-o", opsaniBinaryPath, packagePath)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to build opsani: %s", output))
	}

	os.Exit(m.Run())
}

func WriteConfigFile(config interface{}) {
	data, err := yaml.Marshal(config)
	if err != nil {
		panic("failed to marshal YAML")
	}
	err = ioutil.WriteFile(opsaniConfigPath, data, 0644)
	if err != nil {
		panic("failed to marshal YAML")
	}
}
