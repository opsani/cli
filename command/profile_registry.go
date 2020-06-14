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
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Servo represents a deployed Servo assembly running somewhere
type Servo struct {
	Type string `yaml:"type" mapstructure:"type"`

	// Docker Compose
	User    string `yaml:"user,omitempty" mapstructure:"user,omitempty"`
	Host    string `yaml:"host,omitempty" mapstructure:"host,omitempty"`
	Port    string `yaml:"port,omitempty" mapstructure:"port,omitempty"`
	Path    string `yaml:"path,omitempty" mapstructure:"path,omitempty"`
	Bastion string `yaml:"bastion,omitempty" mapstructure:"bastion,omitempty"`

	// Kubernetes
	Namespace  string `yaml:"namespace,omitempty" mapstructure:"namespace,omitempty"`
	Deployment string `yaml:"deployment,omitempty" mapstructure:"deployment,omitempty"`
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

// Profile represents an Opsani app, token, and base URL
type Profile struct {
	Name      string `yaml:"name" mapstructure:"name" json:"name"`
	Optimizer string `yaml:"optimizer" mapstructure:"optimizer" json:"optimizer"`
	Token     string `yaml:"token" mapstructure:"token" json:"token"`
	BaseURL   string `yaml:"base_url,omitempty" mapstructure:"base_url,omitempty" json:"base_url,omitempty"`
	Servo     Servo  `yaml:"servo,omitempty" mapstructure:"servo,omitempty" json:"servo,omitempty"`
}

// Organization returns the domain of the organization that owns the app
func (p Profile) Organization() string {
	return filepath.Dir(p.Optimizer)
}

// AppName returns the name of the app
func (p Profile) AppName() string {
	return filepath.Base(p.Optimizer)
}

// IsActive indicates if the profile is active
func (p Profile) IsActive() bool {
	return p.Name == "default"
}

// ProfileRegistry provides an interface for managing configuration of app profiles
type ProfileRegistry struct {
	viper    *viper.Viper
	profiles []*Profile
}

// NewProfileRegistry returns a new registry of configured app profiles
func NewProfileRegistry(viper *viper.Viper) (*ProfileRegistry, error) {
	profiles := make([]*Profile, 0)
	err := viper.UnmarshalKey("profiles", &profiles)
	if err != nil {
		return nil, err
	}

	return &ProfileRegistry{
		viper:    viper,
		profiles: profiles,
	}, nil
}

// Profiles returns the Profiles in the configuration
func (pr *ProfileRegistry) Profiles() []*Profile {
	return pr.profiles
}

// lookupProfile named returns the Profile with the given name and its index in the config
func (pr *ProfileRegistry) lookupProfile(name string) (*Profile, int) {
	var profile *Profile
	var index int
	for i, s := range pr.profiles {
		if s.Name == name {
			profile = s
			index = i
			break
		}
	}

	return profile, index
}

// ProfileNamed named returns the Profile with the given name
func (pr *ProfileRegistry) ProfileNamed(name string) *Profile {
	profile, _ := pr.lookupProfile(name)
	return profile
}

// AddProfile adds a Profile to the config
func (pr *ProfileRegistry) AddProfile(profile Profile) error {
	pr.profiles = append(pr.profiles, &profile)
	pr.viper.Set("profiles", pr.profiles)
	return nil
}

// RemoveProfileNamed removes a Profile from the config with the given name
func (pr *ProfileRegistry) RemoveProfileNamed(name string) error {
	s, index := pr.lookupProfile(name)
	if s == nil {
		return fmt.Errorf("no such profile %q", name)
	}
	pr.profiles = append(pr.profiles[:index], pr.profiles[index+1:]...)
	pr.viper.Set("profiles", pr.profiles)
	return nil
}

// RemoveProfile removes a Profile from the config
func (pr *ProfileRegistry) RemoveProfile(profile Profile) error {
	return pr.RemoveProfileNamed(profile.Name)
}

// Set sets the Profile on the registry
func (pr *ProfileRegistry) Set(profiles []Profile) {
	pr.viper.Set("profiles", profiles)
}

// Save the data back to the config
func (pr *ProfileRegistry) Save() error {
	pr.viper.Set("profiles", pr.profiles)
	return pr.viper.WriteConfig()
}
