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

package test

import (
	"io/ioutil"
	"os"

	"sigs.k8s.io/yaml"
)

// TempConfigFileWithBytes returns a temporary YAML config file with the given byte array content
func TempConfigFileWithBytes(bytes []byte) *os.File {
	tmpFile, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		panic("failed to create temp file")
	}
	if _, err = tmpFile.Write(bytes); err != nil {
		panic("failed writing to temp file")
	}
	return tmpFile
}

// TempConfigFileWithString returns a temporary YAML config file with the given string content
func TempConfigFileWithString(str string) *os.File {
	return TempConfigFileWithBytes([]byte(str))
}

// TempConfigFileWithObj returns a temporary YAML config file with the given object serialized to YAML
func TempConfigFileWithObj(obj interface{}) *os.File {
	if data, err := yaml.Marshal(obj); data != nil {
		return TempConfigFileWithBytes(data)
	} else if err != nil {
		panic("failed serializing config to YAML")
	}
	return nil
}
