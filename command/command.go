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
	"encoding/json"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/go-resty/resty/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// Command is a wrapper around cobra.Command that adds Opsani functionality
type Command struct {
	*cobra.Command

	// Shadow all Cobra functions with Opsani equivalents
	PersistentPreRun func(cmd *Command, args []string)
	// PersistentPreRunE: PersistentPreRun but returns an error.
	PersistentPreRunE func(cmd *Command, args []string) error
	// PreRun: children of this command will not inherit.
	PreRun func(cmd *Command, args []string)
	// PreRunE: PreRun but returns an error.
	PreRunE func(cmd *Command, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	Run func(cmd *Command, args []string)
	// RunE: Run but returns an error.
	RunE func(cmd *Command, args []string) error
	// PostRun: run after the Run command.
	PostRun func(cmd *Command, args []string)
	// PostRunE: PostRun but returns an error.
	PostRunE func(cmd *Command, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	PersistentPostRun func(cmd *Command, args []string)
	// PersistentPostRunE: PersistentPostRun but returns an error.
	PersistentPostRunE func(cmd *Command, args []string) error
}

// Survey method wrappers
// NOTE: These are necessary because of how the Survey library models in, out, and err

var globalStdio terminal.Stdio

func SetStdio(stdio terminal.Stdio) {
	globalStdio = stdio
}

func (cmd *Command) Stdio() terminal.Stdio {
	if globalStdio != (terminal.Stdio{}) {
		return globalStdio
	} else {
		return terminal.Stdio{
			In:  os.Stdin,
			Out: os.Stdout,
			Err: os.Stderr,
		}
	}
}

// Ask is a wrapper for survey.AskOne that executes with the command's stdio
func (cmd *Command) Ask(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
	stdio := cmd.Stdio()
	return survey.Ask(qs, response, append(opts, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))...)
}

// AskOne is a wrapper for survey.AskOne that executes with the command's stdio
func (cmd *Command) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	stdio := cmd.Stdio()
	return survey.AskOne(p, response, append(opts, survey.WithStdio(stdio.In, stdio.Out, stdio.Err))...)
}

// PrettyPrintJSONObject prints the given object as pretty printed JSON
func (cmd *Command) PrettyPrintJSONObject(obj interface{}) error {
	s, err := prettyjson.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(s))
	return err
}

// PrettyPrintJSONBytes prints the given byte array as pretty printed JSON
func (cmd *Command) PrettyPrintJSONBytes(bytes []byte) error {
	s, err := prettyjson.Format(bytes)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(s))
	return err
}

// PrettyPrintJSONString prints the given string as pretty printed JSON
func (cmd *Command) PrettyPrintJSONString(str string) error {
	return PrettyPrintJSONBytes([]byte(str))
}

// PrettyPrintJSONResponse prints the given API response as pretty printed JSON
func (cmd *Command) PrettyPrintJSONResponse(resp *resty.Response) error {
	if resp.IsSuccess() {
		if r := resp.Result(); r != nil {
			return PrettyPrintJSONObject(r)
		}
	} else if resp.IsError() {
		if e := resp.Error(); e != nil {
			return PrettyPrintJSONObject(e)
		}
	}
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return err
	}
	return PrettyPrintJSONObject(result)
}

// ReduceRunEFuncsO reduces a list of Cobra run functions that return an error into a single aggregate run function
func ReduceRunEFuncsO(runFuncs ...RunEFunc) func(cmd *Command, args []string) error {
	return func(cmd *Command, args []string) error {
		for _, runFunc := range runFuncs {
			if err := runFunc(cmd.Command, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// NewCommandWithCobraCommand returns a new Opsani CLI command with a given Cobra command
func NewCommandWithCobraCommand(cobraCommand *cobra.Command, configFunc func(*Command)) *Command {
	opsaniCommand := &Command{
		Command: cobraCommand,
	}
	cobraCommand.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if opsaniCommand.PersistentPreRun != nil {
			opsaniCommand.PersistentPreRun(opsaniCommand, args)
		}
	}
	cobraCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if opsaniCommand.PersistentPreRunE != nil {
			return opsaniCommand.PersistentPreRunE(opsaniCommand, args)
		}
		return nil
	}
	cobraCommand.PreRun = func(cmd *cobra.Command, args []string) {
		if opsaniCommand.PreRun != nil {
			opsaniCommand.PreRun(opsaniCommand, args)
		}
	}
	cobraCommand.PreRunE = func(cmd *cobra.Command, args []string) error {
		if opsaniCommand.PreRunE != nil {
			return opsaniCommand.PreRunE(opsaniCommand, args)
		}
		return nil
	}
	cobraCommand.Run = func(cmd *cobra.Command, args []string) {
		if opsaniCommand.Run != nil {
			opsaniCommand.Run(opsaniCommand, args)
		}
	}
	cobraCommand.RunE = func(cmd *cobra.Command, args []string) error {
		if opsaniCommand.RunE != nil {
			return opsaniCommand.RunE(opsaniCommand, args)
		}
		return nil
	}
	cobraCommand.PostRun = func(cmd *cobra.Command, args []string) {
		if opsaniCommand.PostRun != nil {
			opsaniCommand.PostRun(opsaniCommand, args)
		}
	}
	cobraCommand.PostRunE = func(cmd *cobra.Command, args []string) error {
		if opsaniCommand.PostRunE != nil {
			return opsaniCommand.PostRunE(opsaniCommand, args)
		}
		return nil
	}
	cobraCommand.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if opsaniCommand.PersistentPostRun != nil {
			opsaniCommand.PersistentPostRun(opsaniCommand, args)
		}
	}
	cobraCommand.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if opsaniCommand.PersistentPostRunE != nil {
			return opsaniCommand.PersistentPostRunE(opsaniCommand, args)
		}
		return nil
	}

	if configFunc != nil {
		configFunc(opsaniCommand)
	}

	return opsaniCommand
}
