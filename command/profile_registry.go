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

	"github.com/spf13/viper"
)

// Profile represents an Opsani app, token, and base URL
type Profile struct {
	Name    string `yaml:"name" mapstructure:"name"`
	App     string `yaml:"app" mapstructure:"app"`
	Token   string `yaml:"token" mapstructure:"token"`
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
}

// ProfileRegistry provides an interface for managing configuration of app profiles
type ProfileRegistry struct {
	viper *viper.Viper
}

// NewProfileRegistry returns a new registry of configured app profiles
func NewProfileRegistry(viper *viper.Viper) ProfileRegistry {
	return ProfileRegistry{
		viper: viper,
	}
}

// Profiles returns the Profiles in the configuration
func (pr *ProfileRegistry) Profiles() ([]Profile, error) {
	profiles := make([]Profile, 0)
	err := pr.viper.UnmarshalKey("profiles", &profiles)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// lookupProfile named returns the Profile with the given name and its index in the config
func (pr *ProfileRegistry) lookupProfile(name string) (*Profile, int) {
	var profile *Profile
	profiles, err := pr.Profiles()
	if err != nil {
		return nil, 0
	}
	var index int
	for i, s := range profiles {
		if s.Name == name {
			profile = &s
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
	profiles, err := pr.Profiles()
	if err != nil {
		return err
	}

	profiles = append(profiles, profile)
	pr.viper.Set("profiles", profiles)
	return nil
}

// RemoveProfileNamed removes a Profile from the config with the given name
func (pr *ProfileRegistry) RemoveProfileNamed(name string) error {
	s, index := pr.lookupProfile(name)
	if s == nil {
		return fmt.Errorf("no such profile %q", name)
	}
	profiles, err := pr.Profiles()
	if err != nil {
		return err
	}
	profiles = append(profiles[:index], profiles[index+1:]...)
	pr.viper.Set("profiles", profiles)
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
	return pr.viper.WriteConfig()
}
