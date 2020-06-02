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

package command

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// ServoRegistry manages a collection of configured Servos
type ServoRegistry struct {
	viper *viper.Viper
}

// NewServoRegistry builds a Servo registry from the Viper configuration
func NewServoRegistry(viper *viper.Viper) ServoRegistry {
	return ServoRegistry{
		viper: viper,
	}
}

// Servos returns the Servos in the configuration
func (sr *ServoRegistry) Servos() ([]Servo, error) {
	servos := make([]Servo, 0)
	err := sr.viper.UnmarshalKey("servos", &servos)
	if err != nil {
		return nil, err
	}
	return servos, nil
}

// lookupServo named returns the Servo with the given name and its index in the config
func (sr *ServoRegistry) lookupServo(name string) (*Servo, int) {
	var servo *Servo
	servos, err := sr.Servos()
	if err != nil {
		return nil, 0
	}
	var index int
	for i, s := range servos {
		if s.Name == name {
			servo = &s
			index = i
			break
		}
	}

	return servo, index
}

// ServoNamed named returns the Servo with the given name
func (sr *ServoRegistry) ServoNamed(name string) *Servo {
	servo, _ := sr.lookupServo(name)
	return servo
}

// AddServo adds a Servo to the config
func (sr *ServoRegistry) AddServo(servo Servo) error {
	servos, err := sr.Servos()
	if err != nil {
		return err
	}

	servos = append(servos, servo)
	sr.viper.Set("servos", servos)
	return sr.viper.WriteConfig()
}

// RemoveServoNamed removes a Servo from the config with the given name
func (sr *ServoRegistry) RemoveServoNamed(name string) error {
	s, index := sr.lookupServo(name)
	if s == nil {
		return fmt.Errorf("no such Servo %q", name)
	}
	servos, err := sr.Servos()
	if err != nil {
		return err
	}
	servos = append(servos[:index], servos[index+1:]...)
	sr.viper.Set("servos", servos)
	return sr.viper.WriteConfig()
}

// RemoveServo removes a Servo from the config
func (sr *ServoRegistry) RemoveServo(servo Servo) error {
	return sr.RemoveServoNamed(servo.Name)
}

// Servo represents a deployed Servo assembly running somewhere
type Servo struct {
	Name string `yaml:"name" mapstructure:"name"`
	Type string `yaml:"type" mapstructure:"type"`

	// Docker Compose
	User string `yaml:"user" mapstructure:"user"`
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
	Path string `yaml:"path" mapstructure:"path"`

	// Kubernetes
	Namespace  string `yaml:"namespace" mapstructure:"namespace"`
	Deployment string `yaml:"deployment" mapstructure:"deployment"`

	Bastion string `yaml:"bastion" mapstructure:"bastion"`
}

// Description returns a textual description of the servo
func (s Servo) Description() string {
	if s.Type == "docker-compose" {
		return s.URL()
	} else if s.Type == "kubernetes" {
		return fmt.Sprintf("namespaces/%s/deployments/%s", s.Namespace, s.Deployment)
	}
	return ""
}

// HostAndPort returns a string with the host and port when different from 22
func (s Servo) HostAndPort() string {
	h := s.Host
	p := s.Port
	if p == "" {
		p = "22"
	}
	return strings.Join([]string{h, p}, ":")
}

// DisplayHost returns the hostname and port if different from 22
func (s Servo) DisplayHost() string {
	v := s.Host
	if s.Port != "" && s.Port != "22" {
		v = v + ":" + s.Port
	}
	return v
}

// DisplayPath returns the path that the Servo is installed at
func (s Servo) DisplayPath() string {
	if s.Type != "docker-compose" {
		return ""
	}
	if s.Path != "" {
		return s.Path
	}
	return "~/"
}

// URL returns an ssh:// URL for accessing the Servo
func (s Servo) URL() string {
	pathComponent := ""
	if s.Path != "" {
		if s.Port != "" && s.Port != "22" {
			pathComponent = pathComponent + ":"
		}
		pathComponent = pathComponent + s.Path
	}
	return fmt.Sprintf("ssh://%s@%s:%s", s.User, s.DisplayHost(), pathComponent)
}

// BastionComponents splits the bastion host identifier into user and host components
func (s Servo) BastionComponents() (string, string) {
	components := strings.Split(s.Bastion, "@")
	user := components[0]
	host := components[1]
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}
	return user, host
}
